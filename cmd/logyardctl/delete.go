package main

import (
	"flag"
	"fmt"
	"log"
	"logyard"
)

type delete struct {
}

func (cmd *delete) Name() string {
	return "delete"
}

func (cmd *delete) DefineFlags(fs *flag.FlagSet) {
}

func (cmd *delete) Run(args []string) error {
	Init()
	if len(args) == 0 {
		return fmt.Errorf("need at least one positional argument")
	}
	for _, name := range args {
		err := logyard.Config.DeleteDrain(name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Deleted drain %s\n", name)
	}
	return nil
}
