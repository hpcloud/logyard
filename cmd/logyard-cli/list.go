package main

import (
	"flag"
	"fmt"
	"logyard"
	"sort"
)

type list struct {
}

func (cmd *list) Name() string {
	return "list"
}

func (cmd *list) DefineFlags(fs *flag.FlagSet) {
}

func (cmd *list) Run(args []string) error {
	config := logyard.GetConfig()
	for _, name := range sortedKeysStringMap(config.Drains) {
		uri := config.Drains[name]
		fmt.Printf("%-20s\t%s\n", name, uri)
	}
	return nil
}

func sortedKeysStringMap(m map[string]string) []string {
	keys := make([]string, len(m))
	idx := 0
	for key, _ := range m {
		keys[idx] = key
		idx++
	}
	sort.Strings(keys)
	return keys
}
