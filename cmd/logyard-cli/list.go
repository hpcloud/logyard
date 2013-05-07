package main

import (
	"flag"
	"fmt"
	"logyard"
)

type list struct {
}

func (cmd *list) Name() string {
	return "list"
}

func (cmd *list) DefineFlags(fs *flag.FlagSet) {
}

func (cmd *list) Run(args []string) error {
	Init("list")
	config := logyard.GetConfig()
	for _, name := range sortedKeys(config.Drains) {
		uri := config.Drains[name]
		fmt.Printf("%-20s\t%s\n", name, uri)
	}
	return nil
}
