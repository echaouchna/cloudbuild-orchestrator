package flow

import (
	"cork/config"
	"cork/gcp"
	"sync"
)

type BuildStatus struct {
	value  string
	logUrl string
}

type Options struct {
	Reference     string
	IncludedTypes []string
	ExcludedTypes []string
}

type ExecutionFlowContext struct {
	options  Options
	exactRef string
	config   config.Config
	triggers map[string]*gcp.BuildTrigger
	statuses map[string]BuildStatus
	lock     sync.Mutex
}

var (
	wg sync.WaitGroup
)

func Execute(configs []config.Config, options Options) {
	defer wg.Wait()
	startLogger(&wg)
	for _, c := range configs {
		config := c
		wg.Add(1)
		go func() {
			run(config, options)
		}()
	}
}
