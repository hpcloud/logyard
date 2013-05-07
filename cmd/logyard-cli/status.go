package main

import (
	"flag"
	"fmt"
	"logyard"
	"logyard/util/statecache"
	"stackato/server"
)

type status struct {
}

func (cmd *status) Name() string {
	return "status"
}

func (cmd *status) DefineFlags(fs *flag.FlagSet) {
}

func (cmd *status) Run(args []string) error {
	Init("status")
	cache := &statecache.StateCache{
		"logyard:drainstatus:",
		server.LocalIPMust(),
		logyard.NewRedisClientMust(server.Config.CoreIP+":6464", 0)}

	var drains []string
	if len(args) > 0 {
		drains = args
	} else {
		config := logyard.GetConfig()
		drains = sortedKeys(config.Drains)
	}

	for _, name := range drains {
		states, err := cache.GetState(name)
		if err != nil {
			return fmt.Errorf("Unable to retrieve cached state: %v", err)
		}
		for _, nodeip := range sortedKeys(states) {
			fmt.Printf("%-20s\t%s\t%s\n", name, nodeip, states[nodeip])
		}
	}
	return nil
}
