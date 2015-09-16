package commands

import (
	"flag"
	"fmt"
	"github.com/hpcloud/log"
	"logyard"
)

type recv struct {
	json       bool
	hideprefix bool
	filter     string
}

func (cmd *recv) Name() string {
	return "recv"
}

func (cmd *recv) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.json, "json", false, "Output result as JSON")
	fs.BoolVar(&cmd.hideprefix, "hideprefix", false, "hide message prefix")
	fs.StringVar(&cmd.filter, "filter", "", "filter by message key pattern")
}

func (cmd *recv) Run(args []string) (string, error) {
	if cmd.json {
		return "", fmt.Errorf("--json not supported by this subcommand")
	}
	sub := logyard.Broker.Subscribe(cmd.filter)
	for msg := range sub.Ch {
		if cmd.hideprefix {
			fmt.Println(msg.Value)
		} else {
			fmt.Println(msg.Key, msg.Value)
		}
	}
	err := sub.Wait()
	if err != nil {
		log.Fatal(err)
	}
	return "", nil
}
