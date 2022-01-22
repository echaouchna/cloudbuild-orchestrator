package cmd

import (
	"cork/utils"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/juliangruber/go-intersect"
)

type Options struct {
	version         bool
	NoFastFailing   bool
	Reference       string
	Included        []string
	Excluded        []string
	Filename        string
	NumParallelJobs int
}

var (
	options     Options
	corkVersion string
	included    string
	excluded    string
)

func init() {
	flag.StringVar(&options.Reference, "reference", "develop", "Reference to use for the build")
	flag.BoolVar(&options.version, "version", false, "Version")
	flag.StringVar(&included, "include", "", "Types to be included")
	flag.StringVar(&excluded, "exclude", "", "Types to be excluded")
	flag.BoolVar(&options.NoFastFailing, "no-fast-failing", false, "No fast failing")
	flag.IntVar(&options.NumParallelJobs, "parallel", 20, "The number of parallel jobs")
}

func Parse() Options {

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(), "Usage: %s "+
				"[-exclude \"<typeA,typeB,...>\"] "+
				"[-include \"<type1,type2,...>\"] "+
				"[-no-fast-failing] "+
				"[-reference <ref>] "+
				"[-parallel <number>] "+
				"<config_file>\n", os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()
	if options.version {
		fmt.Println(corkVersion)
		os.Exit(0)
	}

	if condition := len(flag.Args()) != 1; condition {
		flag.Usage()
		os.Exit(1)
	}
	options.Filename = flag.Arg(0)

	options.Included = utils.RemoveEmptyStrings(strings.Split(included, ","))
	options.Excluded = utils.RemoveEmptyStrings(strings.Split(excluded, ","))

	if intersect := intersect.Hash(options.Included, options.Excluded); len(intersect) > 0 {
		fmt.Printf("WARNING: The following types are included and excluded: %s\n", intersect)
	}

	fmt.Println("Using reference: " + options.Reference)
	fmt.Printf("Fast failing: %v\n", !options.NoFastFailing)

	return options
}
