package main

import (
	"context"
	"flag"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/sharvanath/kromium/core"
	"github.com/sharvanath/kromium/schema"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
)

const version = "0.1.7"

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	render := flag.Bool("render", true, "Render UI")
	runConfig := flag.String("run", "", "Run the schema")
	validate := flag.String("validate", "", "Validate the pipeline schema")
	parallelism := flag.Int("P", runtime.GOMAXPROCS(0), "The parallelism for the run loop")
	flag.Parse()

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	if *printVersion {
		fmt.Printf("Kromium Version: %s\n", version)
		os.Exit(0)
	} else if validate != nil && *validate != "" {
		err := schema.ValidatePipelineConfig(*validate)
		if err != nil {
			fmt.Printf("Error validating %s: %v", *validate, err)
			os.Exit(1)
		}
		fmt.Printf("Successfuly validated %s\n", *validate)
	} else if runConfig != nil && *runConfig != "" {
		fmt.Printf("Running %s\n", *runConfig)
		config, err := schema.ConvertToPipelineConfig(*runConfig)
		if err != nil {
			fmt.Println("Error reading the schema:", err)
			os.Exit(1)
		}
		if err = config.Init(context.Background()); err != nil {
			fmt.Println("Error initializing the schema:", err)
			os.Exit(1)
		}
		defer config.Close()
		go func() {
			if *render {
				for e := range ui.PollEvents() {
					if e.Type == ui.KeyboardEvent {
						break
					}
				}
				ui.Close()
			}
		}()

		err = core.RunPipelineLoop(context.Background(), config, *parallelism, *render)
		if err != nil {
			ui.Close()
			fmt.Println("Error running pipeline:", err)
			os.Exit(1)
		}
	} else {
		flag.Usage()
		os.Exit(0)
	}
}
