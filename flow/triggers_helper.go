package flow

import (
	"cork/gcp"
	"cork/utils"
	"log"
	"regexp"
	"time"
)

const (
	gitShaRegex = "^[0-9a-f]{5,40}$"
)

func setExactRef(exactRef *string, build *gcp.BuildOperation) {
	if matched, _ := regexp.MatchString(gitShaRegex, *exactRef); !matched {
		*exactRef = build.CommitSha
	}
}

func getSourceRepo(ref string) gcp.RepoSource {
	if matched, _ := regexp.MatchString(gitShaRegex, ref); matched {
		return gcp.RepoSource{
			CommitSha: ref,
		}
	}
	return gcp.RepoSource{
		BranchName: ref,
	}
}

func getRef(cloudBuildRef string, exactRef string) string {
	if exactRef == "" {
		return cloudBuildRef
	}
	return exactRef
}

func waitForBuild(projectId string, buildId string) string {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	retries := 3
	for range ticker.C {
		status, err := gcp.GetBuild(projectId, buildId)
		if err != nil {
			if retries == 0 {
				log.Fatalln(err)
			}
			retries -= 1
		}
		if utils.Contains([]string{gcp.SUCCESS, gcp.FAILURE, gcp.CANCELLED}, status) {
			return status
		}
	}
	return ""
}
