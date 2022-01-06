package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigFile  string `yaml:"-"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Author      string `yaml:"author,omitempty"`
	Steps       []Step `yaml:"steps"`
}
type Parallel struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Trigger     string `yaml:"trigger"`
	ProjectId   string `yaml:"project-id"`
	Type        string `yaml:"type,omitempty"`
	DependsOn   string `yaml:"depends-on,omitempty"`
}
type Step struct {
	Name        string     `yaml:"name,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Trigger     string     `yaml:"trigger,omitempty"`
	ProjectId   string     `yaml:"project-id,omitempty"`
	Type        string     `yaml:"type,omitempty"`
	DependsOn   string     `yaml:"depends-on,omitempty"`
	Parallel    []Parallel `yaml:"parallel,omitempty"`
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
	// fmt.Printf("%#v\n", config)
	// empJSON, err := yaml.Marshal(config)
	// if err != nil {
	// 	log.Fatalf(err.Error())
	// }
	// fmt.Printf("%s\n", string(empJSON))
	return config
}
