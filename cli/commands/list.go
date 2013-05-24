package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"logyard"
	"sort"
)

type list struct {
	json bool
}

func (cmd *list) Name() string {
	return "list"
}

func (cmd *list) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.json, "json", false, "Output result as JSON")
}

func (cmd *list) Run(args []string) (string, error) {
	config := logyard.GetConfig()
	if cmd.json {
		data, err := json.Marshal(config.Drains)
		return string(data), err
	} else {
		for _, name := range sortedKeysStringMap(config.Drains) {
			uri := config.Drains[name]
			fmt.Printf("%-20s\t%s\n", name, uri)
		}
		return "", nil
	}
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
