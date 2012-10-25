package main

import (
	"flag"
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

func (cmd *delete) Run(args []string) {
	Init()
	for _, name := range args {
		err := logyard.Config.DeleteDrain(name)
		if err != nil {
			log.Fatal(err)
		}
	}
}
