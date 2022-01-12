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
	NoFastFailing bool
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
	for _, c := range configs {
		config := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			run(config, options)
		}()
	}
}
