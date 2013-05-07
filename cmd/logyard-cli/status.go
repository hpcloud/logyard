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
	if len(args) > 0 {
		return fmt.Errorf("not supported yet")
	}
	Init("status")
	cache := &statecache.StateCache{
		"logyard:drainstatus:",
		server.LocalIPMust(),
		logyard.NewRedisClientMust(server.Config.CoreIP+":6464", 0)}

	config := logyard.GetConfig()
	for _, name := range sortedKeys(config.Drains) {
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
