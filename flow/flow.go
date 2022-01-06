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

type ExecutionFlowContext struct {
	reference string
	exactRef  string
	config    config.Config
	triggers  map[string]*gcp.BuildTrigger
	statuses  map[string]BuildStatus
	lock      sync.Mutex
}

var (
	wg sync.WaitGroup
)

func Execute(configs []config.Config, reference string) {
	defer wg.Wait()
	startLogger(&wg)
	for _, c := range configs {
		config := c
		wg.Add(1)
		go func() {
			run(config, reference)
		}()
	}
}
