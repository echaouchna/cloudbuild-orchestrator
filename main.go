package main

import (
	"cork/cmd"
	"cork/config"
	"cork/flow"
)

func main() {

	options := cmd.Parse()

	c := config.Unmarshal(options.Filename)

	filteredConfig := c.Filter(options.Included, options.Excluded)

	flow.Execute([]config.Config{filteredConfig}, options)
}
