package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sharvanath/kromium/core"
	"os"
)

const version = "beta-0.1"

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")
	runConfig := flag.String("run", "", "The config to run sync")
	flag.Parse()

	if *printVersion {
		fmt.Printf("Kromium Version: %s\n", version)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	if runConfig != nil && *runConfig != "" {
		fmt.Printf("Running %s\n", *runConfig)
		config, err := core.ReadPipelineConfigFile(*runConfig)
		if err != nil {
			fmt.Println("Error reading the config:", err)
		}

		err = core.RunPipeline(context.Background(), config)
		if err != nil {
			fmt.Println("Error running pipeline:", err)
		}
	}
}
