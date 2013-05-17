package main

import (
	"flag"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"logyard"
	"logyard/util/statecache"
	"sort"
	"stackato/server"
	"strconv"
	"strings"
)

type status struct {
	notrunning *bool
}

func (cmd *status) Name() string {
	return "status"
}

func (cmd *status) DefineFlags(fs *flag.FlagSet) {
	cmd.notrunning = fs.Bool(
		"notrunning", false, "show all drains, including running ones")
}

func (cmd *status) Run(args []string) error {
	Init("status")
	cache := &statecache.StateCache{
		"logyard:drainstatus:",
		server.LocalIPMust(),
		logyard.NewRedisClientMust(
			server.GetClusterConfig().MbusIp+":6464",
			0)}

	var drains []string
	if len(args) > 0 {
		drains = args
	} else {
		config := logyard.GetConfig()
		drains = sortedKeysStringMap(config.Drains)
	}

	for _, name := range drains {
		states, err := cache.GetState(name)
		if err != nil {
			return fmt.Errorf("Unable to retrieve cached state: %v", err)
		}
		for _, nodeip := range sortedKeysStateMap(states) {
			running := strings.Contains(states[nodeip]["name"], "RUNNING")
			if *cmd.notrunning && running {
				continue
			}
			printStatus(name, nodeip, states[nodeip])
		}
	}
	return nil
}

func printStatus(name, nodeip string, info statecache.StateInfo) error {
	rev, err := strconv.Atoi(info["rev"])
	if err != nil {
		return fmt.Errorf("Corrupt drain status: %v", err)
	}
	state := info["name"]

	fmt.Printf("%-20s\t%s\t%s[%d]", name, nodeip, state, rev)
	if error, ok := info["error"]; ok {
		color.Printf("\t@r%s@|", error)
	}
	fmt.Println()
	return nil
}

func sortedKeysStateMap(m map[string]statecache.StateInfo) []string {
	keys := make([]string, len(m))
	idx := 0
	for key, _ := range m {
		keys[idx] = key
		idx++
	}
	sort.Strings(keys)
	return keys
}
