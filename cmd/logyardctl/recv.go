package main

import (
	"flag"
	"fmt"
	"github.com/srid/log2"
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
	c, err := logyard.NewClientGlobal()
	if err != nil {
		log2.Fatal(err)
	}
	ss, err := c.Recv([]string{*cmd.filter})
	if err != nil {
		log2.Fatal(err)
	}
	for msg := range ss.Ch {
		if *cmd.hideprefix {
			fmt.Println(msg.Value)
		} else {
			fmt.Println(msg.Key, msg.Value)
		}
	}
	err = ss.Wait()
	if err != nil {
		log2.Fatal(err)
	}
	return nil
}
