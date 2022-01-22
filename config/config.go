package config

import (
	"cork/dag"
	"cork/gcp"
	"cork/utils"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Author      string `yaml:"author,omitempty"`
	ConfigFile  string `yaml:"-"`
	Description string `yaml:"description,omitempty"`
	Name        string `yaml:"name"`
	Steps       []Step `yaml:"steps"`
}
type Step struct {
	DependsOn   []string `yaml:"depends-on,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Manual      bool     `yaml:"manual,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	ProjectId   string   `yaml:"project-id,omitempty"`
	Status      string   `yaml:"status,omitempty"`
	Tags        string   `yaml:"tags,omitempty"`
	Trigger     string   `yaml:"trigger,omitempty"`
	LogUrl      string   `yaml:"log-url,omitempty"`
}

func (step Step) GetKey() string {
	return step.Name
}

func (step Step) HasFinished() bool {
	return step.Status == gcp.SUCCESS ||
		step.Status == gcp.FAILURE ||
		step.Status == gcp.CANCELLED
}

func (step Step) IsSuccessful() bool {
	return step.Status == gcp.SUCCESS
}

func (step Step) HasStarted() bool {
	return step.Status != ""
}

type Steps []Step

func (steps Steps) Items() []dag.Task {
	tasks := []dag.Task{}
	for _, step := range steps {
		tasks = append(tasks, dag.Task(step))
	}
	return tasks
}

func shouldHandleTrigger(includedTags []string, excludedTags []string, stepTags string) bool {
	tags := strings.Split(stepTags, ",")
	retVal := false
	for _, tag := range tags {
		if (len(includedTags) == 0 ||
			utils.MatchAtLeastOne(includedTags, tag)) &&
			(len(excludedTags) == 0 ||
				!utils.MatchAtLeastOne(excludedTags, tag)) {
			retVal = true
		}
	}
	return retVal
}

func (config Config) Filter(includeTags []string, excludedTags []string) (filteredConfig Config) {
	filteredConfig.Author = config.Author
	filteredConfig.ConfigFile = config.ConfigFile
	filteredConfig.Description = config.Description
	filteredConfig.Name = config.Name

	steps := []Step{}

	for _, step := range config.Steps {
		if condition := shouldHandleTrigger(includeTags, excludedTags, step.Tags); condition {
			steps = append(steps, step)
		}
	}
	filteredConfig.Steps = steps
	return
}

func (config Config) GetLinks() (links map[string][]string) {
	links = map[string][]string{}
	for _, task := range config.Steps {
		if len(task.DependsOn) > 0 {
			links[task.Name] = task.DependsOn
		}
	}
	return
}

func Unmarshal(path string) Config {
	config := Config{ConfigFile: path}
	source, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
