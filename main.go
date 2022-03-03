package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sharvanath/kromium/core"
	"net/http"
	"os"
	_ "net/http/pprof"
)

const version = "beta-0.1"

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	runConfig := flag.String("run", "", "The config to run sync")
	parallelism := flag.Int("parallelism", 4, "The parallelism for the run loop")
	flag.Parse()

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	if *printVersion {
		fmt.Printf("Kromium Version: %s\n", version)
		os.Exit(0)
	} else if runConfig != nil && *runConfig != "" {
		fmt.Printf("Running %s\n", *runConfig)
		config, err := core.ReadPipelineConfigFile(context.Background(), *runConfig)
		if err != nil {
			fmt.Println("Error reading the config:", err)
		}
		defer config.Close()

		err = core.RunPipelineLoop(context.Background(), config, *parallelism, true)
		if err != nil {
			fmt.Println("Error running pipeline:", err)
		}
	} else {
		flag.Usage()
		os.Exit(0)
	}
}
