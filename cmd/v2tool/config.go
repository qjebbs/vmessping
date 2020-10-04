package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/qjebbs/v2tool/config"
)

func mergeConfig(args []string) {
	configCmd := flag.NewFlagSet("v2tool config", flag.ExitOnError)
	var files stringArrayFlags
	configCmd.Var(&files, "i", "input path, could be path of json or folder contains them")
	err := configCmd.Parse(args)

	if len(files) == 0 {
		configCmd.Usage()
		os.Exit(1)
	}

	c, err := config.MergeJSONs(files)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v", string(c))
}
