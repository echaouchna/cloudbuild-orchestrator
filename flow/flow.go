package flow

import (
	"cork/cmd"
	"cork/config"
	"cork/dag"
	"fmt"
	"log"
	"strings"
	"sync"
)

func Execute(configs []config.Config, options cmd.Options) {
	wg := sync.WaitGroup{}
	defer wg.Wait()
	for _, c := range configs {
		d, err := dag.BuildDag(config.Steps(c.Steps), c.GetLinks())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("# %s:\n", c.Name)
		fmt.Print(d)
		manualStep := []string{}
		for _, node := range d.Nodes {
			step := node.Task.(config.Step)
			if step.Manual {
				manualStep = append(manualStep, "\t"+step.Name)
			}
		}
		if len(manualStep) > 0 {
			fmt.Println("Manual steps:\n" + strings.Join(manualStep, "\n"))
		}
		wg.Add(1)
		go func(conf config.Config) {
			defer wg.Done()
			run(d, &conf, options)
		}(c)
	}
}
