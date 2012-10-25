package main

import (
	"flag"
	"fmt"
	"log"
	"logyard"
	"net/url"
	"strings"
)

type add struct {
	scheme *string
	host   *string
	filter *string
	format *string
}

func (cmd *add) Name() string {
	return "add"
}

func (cmd *add) DefineFlags(fs *flag.FlagSet) {
	cmd.scheme = fs.String("scheme", "", "Drain scheme (eg: tcp, udp, redis)")
	cmd.host = fs.String("host", "", "Drain hostname/port (eg: logs.loggly.com:12345)")
	cmd.filter = fs.String("filter", "", "Message filters separated by ; (eg: systail;events)")
	cmd.format = fs.String("format", "", "Template to format the json record (eg: {{.Node}}: {{.Text}})")
}

func (cmd *add) Run(args []string) {
	name := args[0]
	Init()

	// TODO: format
	uri := fmt.Sprintf("%s://%s/?", *cmd.scheme, *cmd.host)
	noQuery := true

	for _, filter := range strings.Split(*cmd.filter, ";") {
		if !noQuery {
			uri += "&"
		}
		uri += "filter=" + url.QueryEscape(filter)
		noQuery = false
	}

	if *cmd.format != "" {
		if !noQuery {
			uri += "&"
		}
		uri += "format=" + url.QueryEscape(*cmd.format)
	}
	fmt.Println(uri)
	err := logyard.Config.AddDrain(name, uri)
	if err != nil {
		log.Fatal(err)
	}
}
