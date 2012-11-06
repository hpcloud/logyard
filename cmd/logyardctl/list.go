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
	Init()
	for _, name := range sortedKeys(logyard.Config.Drains) {
		uri := logyard.Config.Drains[name]
		fmt.Printf("%-20s\t%s\n", name, uri)
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	idx := 0
	for key, _ := range m {
		keys[idx] = key
		idx++
	}
	sort.Strings(keys)
	return keys
}
