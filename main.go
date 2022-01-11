package main

import (
	"cork/config"
	"cork/config/validate"
	"cork/flow"
	"cork/utils"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/juliangruber/go-intersect"
)

var (
	corkVersion string
	flags       struct {
		version       bool
		NoFastFailing bool
		reference     string
		included      string
		excluded      string
	}
)

func init() {
	flag.StringVar(&flags.reference, "reference", "develop", "Reference to use for the build")
	flag.BoolVar(&flags.version, "version", false, "Version")
	flag.StringVar(&flags.included, "include", "", "Types to be included")
	flag.StringVar(&flags.excluded, "exclude", "", "Types to be excluded")
	flag.BoolVar(&flags.NoFastFailing, "no-fast-failing", false, "No fast failing")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-exclude \"<typeA,typeB,...>\"] [-include \"<type1,type2,...>\"] [-no-fast-failing] [-reference <ref>] <config_file>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flags.version {
		fmt.Println(corkVersion)
		os.Exit(0)
	}

	if condition := len(flag.Args()) != 1; condition {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Using reference: " + flags.reference)

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

	includedTypes := utils.RemoveEmptyStrings(strings.Split(flags.included, ","))
	excludedTypes := utils.RemoveEmptyStrings(strings.Split(flags.excluded, ","))

	if intersect := intersect.Hash(includedTypes, excludedTypes); len(intersect) > 0 {
		fmt.Printf("WARNING: The following types are included and excluded: %s\n", intersect)
	}

	flow.Execute([]config.Config{c}, flow.Options{
		Reference:     flags.reference,
		NoFastFailing: flags.NoFastFailing,
		IncludedTypes: includedTypes,
		ExcludedTypes: excludedTypes,
	})
}
