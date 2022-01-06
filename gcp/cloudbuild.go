package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
)

const (
	RUNNING   = "RUNNING"
	SUCCESS   = "SUCCESS"
	FAILURE   = "FAILURE"
	CANCELLED = "CANCELLED"
)

type BuildTrigger = cloudbuild.BuildTrigger

type RepoSource = cloudbuild.RepoSource

type BuildOperationMetadata struct {
	Type  string `json:"@type"`
	Build struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Source struct {
		} `json:"source"`
		CreateTime time.Time `json:"createTime"`
		Steps      []struct {
			Name       string   `json:"name"`
			Args       []string `json:"args"`
			Dir        string   `json:"dir"`
			ID         string   `json:"id"`
			Entrypoint string   `json:"entrypoint,omitempty"`
			SecretEnv  []string `json:"secretEnv,omitempty"`
		} `json:"steps"`
		Timeout        string `json:"timeout"`
		ProjectID      string `json:"projectId"`
		LogsBucket     string `json:"logsBucket"`
		BuildTriggerID string `json:"buildTriggerId"`
		Options        struct {
			SubstitutionOption   string   `json:"substitutionOption"`
			Logging              string   `json:"logging"`
			Env                  []string `json:"env"`
			DynamicSubstitutions bool     `json:"dynamicSubstitutions"`
			Pool                 struct {
			} `json:"pool"`
		} `json:"options"`
		LogURL        string            `json:"logUrl"`
		Substitutions map[string]string `json:"substitutions"`
		Tags          []string          `json:"tags"`
		Artifacts     struct {
			Objects struct {
				Location string   `json:"location"`
				Paths    []string `json:"paths"`
			} `json:"objects"`
		} `json:"artifacts"`
		QueueTTL         string `json:"queueTtl"`
		Name             string `json:"name"`
		AvailableSecrets struct {
			SecretManager []struct {
				VersionName string `json:"versionName"`
				Env         string `json:"env"`
			} `json:"secretManager"`
		} `json:"availableSecrets"`
	} `json:"build"`
}

type CloudBuildOperationError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
			Domain  string `json:"domain"`
			Reason  string `json:"reason"`
		} `json:"errors"`
		Status string `json:"status"`
	} `json:"error"`
}

type BuildOperation struct {
	ID        string
	LogURL    string
	CommitSha string
}

// TriggerCloudBuild triggers a GCP Cloudbuild.
func TriggerCloudBuild(projectId string, triggerId string, repoSource RepoSource) (*BuildOperation, error) {
	ctx := context.Background()
	cloudbuildService, _ := cloudbuild.NewService(ctx)
	operation, err := cloudbuildService.Projects.Triggers.Run(projectId, triggerId, &repoSource).Do()
	if apiErr, ok := err.(*googleapi.Error); ok {
		buildOperationError := CloudBuildOperationError{}
		json.Unmarshal([]byte(apiErr.Body), &buildOperationError)
		return nil, errors.New(buildOperationError.Error.Message)
	}
	buildOperationMetadata := BuildOperationMetadata{}
	json.Unmarshal(operation.Metadata, &buildOperationMetadata)
	return &BuildOperation{
		ID:        buildOperationMetadata.Build.ID,
		LogURL:    buildOperationMetadata.Build.LogURL,
		CommitSha: buildOperationMetadata.Build.Substitutions["REVISION_ID"],
	}, nil
}

func ListTriggers(projectId string) map[string]*BuildTrigger {
	buildTriggers := make(map[string]*BuildTrigger)
	ctx := context.Background()
	cloudbuildService, _ := cloudbuild.NewService(ctx)

	operation, err := cloudbuildService.Projects.Triggers.List(projectId).Do()
	// Google API verbose debugging info.
	if apiErr, ok := err.(*googleapi.Error); ok {
		log.Println(apiErr.Body)
	}

	for _, trigger := range operation.Triggers {
		buildTriggers[projectId+"/"+trigger.Name] = trigger
	}

	return buildTriggers
}

func GetBuild(projectId string, buildId string) (string, error) {
	ctx := context.Background()
	cloudbuildService, _ := cloudbuild.NewService(ctx)

	operation, err := cloudbuildService.Projects.Builds.Get(projectId, buildId).Do()
	// Google API verbose debugging info.
	if apiErr, ok := err.(*googleapi.Error); ok {
		log.Println(apiErr.Body)
		buildOperationError := CloudBuildOperationError{}
		json.Unmarshal([]byte(apiErr.Body), &buildOperationError)
		return "", errors.New(buildOperationError.Error.Message)
	}

	return operation.Status, nil
}
