package commands

import (
	"flag"
	"fmt"
	"github.com/ActiveState/logyard"
)

type delete struct {
	json bool
}

func (cmd *delete) Name() string {
	return "delete"
}

func (cmd *delete) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.json, "json", false, "Output result as JSON")
}

func (cmd *delete) Run(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("need at least one positional argument")
	}
	for _, name := range args {
		err := logyard.DeleteDrain(name)
		// In case of an error, exit abrutly ignoring the rest of the
		// drains.
		if err != nil {
			return "", err
		}
		if !cmd.json {
			fmt.Printf("Deleted drain %s\n", name)
		}
	}
	if cmd.json {
		return "{}", nil
	} else {
		return "", nil
	}
}
