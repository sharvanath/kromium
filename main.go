package main

import (
	"context"
	"encoding/json"
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
		file, _ := os.Open(*runConfig)
		defer file.Close()
		decoder := json.NewDecoder(file)
		config := core.SyncConfig{}
		err := decoder.Decode(&config)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println(core.RunSync(context.Background(), config))
	}
}
