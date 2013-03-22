package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
)

type recv struct {
	hideprefix *bool
	filter     *string
}

func (cmd *recv) Name() string {
	return "recv"
}

func (cmd *recv) DefineFlags(fs *flag.FlagSet) {
	cmd.hideprefix = fs.Bool("hideprefix", false, "hide message prefix")
	cmd.filter = fs.String("filter", "", "filter by message key pattern")
}

func (cmd *recv) Run(args []string) error {
	sub := logyard.Broker.Subscribe(*cmd.filter)
	for msg := range sub.Ch {
		if *cmd.hideprefix {
			fmt.Println(msg.Value)
		} else {
			fmt.Println(msg.Key, msg.Value)
		}
	}
	err := sub.Wait()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
