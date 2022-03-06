package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sharvanath/kromium/config"
	"github.com/sharvanath/kromium/core"
	"net/http"
	"os"
	_ "net/http/pprof"
)

const version = "0.1.0"

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	runConfig := flag.String("run", "", "Run the config")
	validate := flag.String("validate", "", "Validate the pipeline config")
	parallelism := flag.Int("P", 1, "The parallelism for the run loop")
	flag.Parse()

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	if *printVersion {
		fmt.Printf("Kromium Version: %s\n", version)
		os.Exit(0)
	} else if validate != nil && *validate != "" {
		err := config.ValidatePipelineConfig(*validate)
		if err != nil {
			fmt.Printf("Error validating %s: %v", *validate, err)
			os.Exit(1)
		}
		fmt.Printf("Successfuly validated %s\n", *validate)
	} else if runConfig != nil && *runConfig != "" {
		fmt.Printf("Running %s\n", *runConfig)
		config, err := config.ConvertToPipelineConfig(*runConfig)
		if err != nil {
			fmt.Println("Error reading the config:", err)
			os.Exit(1)
		}
		config.Init(context.Background())
		defer config.Close()
		err = core.RunPipelineLoop(context.Background(), config, *parallelism, true)
		if err != nil {
			fmt.Println("Error running pipeline:", err)
			os.Exit(1)
		}
	} else {
		flag.Usage()
		os.Exit(0)
	}
}
