package main

import (
	"cork/config"
	"cork/config/validate"
	"cork/flow"
	"flag"
	"fmt"
	"os"
)

var (
	reference   string
	version     bool
	corkVersion string
)

func init() {
	flag.StringVar(&reference, "reference", "develop", "Reference to use for the build")
	flag.BoolVar(&version, "version", false, "version")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-version] [-reference <ref>] <config_file>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if version {
		fmt.Println(corkVersion)
		os.Exit(0)
	}

	if condition := len(flag.Args()) != 1; condition {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Using reference: " + reference)

	filename := flag.Arg(0)
	c := config.Unmarshal(filename)

	warnings, errors := validate.ValidateConfig(c)

	for _, warning := range warnings {
		fmt.Printf("WARNING: %s\n", warning.Message)
	}

	if len(errors) > 0 {
		es := ""
		for _, e := range errors {
			es += e + "\n"
		}
		fmt.Println(es)
		os.Exit(1)
	}

	flow.Execute([]config.Config{c}, reference)
}
