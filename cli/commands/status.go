package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"logyard"
	"logyard/util/golor"
	"logyard/util/statecache"
	"sort"
	"stackato/server"
	"strconv"
	"strings"
)

type status struct {
	json       bool
	prefix     bool
	notrunning bool
}

func (cmd *status) Name() string {
	return "status"
}

func (cmd *status) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.json, "json", false,
		"Output result as JSON")
	fs.BoolVar(&cmd.prefix, "prefix", false,
		"Treat drain names as prefix")
	fs.BoolVar(&cmd.notrunning, "notrunning", false,
		"show only drains not running")
}

func (cmd *status) GetDrains(args []string) ([]string, error) {
	var drains []string

	config := logyard.GetConfig()

	if cmd.prefix {
		// Return drains matching the given prefix (args[0])
		if len(args) != 1 {
			return nil, fmt.Errorf("Need exactly 1 position arg")
		}
		prefix := args[0]
		for name, _ := range config.Drains {
			if strings.HasPrefix(name, prefix) {
				drains = append(drains, name)
			}
		}
	} else if len(args) > 0 {
		drains = args
	} else {
		// Return all drains
		drains = sortedKeysStringMap(config.Drains)
	}

	return drains, nil
}

func (cmd *status) Run(args []string) (string, error) {
	cache := &statecache.StateCache{
		"logyard:drainstatus:",
		server.LocalIPMust(),
		server.NewRedisClientMust(
			server.GetClusterConfig().MbusIp+":6464",
			"",
			0)}

	drains, err := cmd.GetDrains(args)
	if err != nil {
		return "", err
	}
	data := make(map[string]map[string]statecache.StateInfo)

	for _, name := range drains {
		states, err := cache.GetState(name)
		if err != nil {
			return "", fmt.Errorf("Unable to retrieve cached state: %v", err)
		}
		data[name] = states
	}

	if cmd.json {
		b, err := json.Marshal(data)
		return string(b), err
	} else {
		for name, states := range data {
			for _, nodeip := range sortedKeysStateMap(states) {
				running := strings.Contains(states[nodeip]["name"], "RUNNING")
				if cmd.notrunning && running {
					continue
				}
				printStatus(name, nodeip, states[nodeip])
			}
		}
		return "", nil
	}
}

func printStatus(name, nodeip string, info statecache.StateInfo) error {
	rev, err := strconv.Atoi(info["rev"])
	if err != nil {
		return fmt.Errorf("Corrupt drain status: %v", err)
	}
	state := info["name"]

	fmt.Printf("%-20s\t%s\t%s[%d]", name, nodeip, state, rev)
	if error, ok := info["error"]; ok {
		fmt.Printf("\t%s", golor.Colorize(error, golor.RGB(5, 0, 0), -1))
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
