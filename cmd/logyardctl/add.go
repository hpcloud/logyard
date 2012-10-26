package main

import (
	"flag"
	"fmt"
	"log"
	"logyard"
	"net/url"
	"strings"
)

// Example:
//  .. add -scheme redis -host core -filter systail.kato -params "limit=200;key=kato_history" kato_history
type add struct {
	scheme *string
	host   *string
	filter *string
	format *string
	params *string
}

func (cmd *add) Name() string {
	return "add"
}

func (cmd *add) DefineFlags(fs *flag.FlagSet) {
	cmd.scheme = fs.String("scheme", "", "Drain scheme (eg: tcp, udp, redis)")
	cmd.host = fs.String("host", "", "Drain hostname/port (eg: logs.loggly.com:12345)")
	cmd.filter = fs.String("filter", "", "Message filters separated by ; (eg: systail;events)")
	cmd.format = fs.String("format", "", "Template to format the json record (eg: {{.Node}}: {{.Text}})")
	cmd.params = fs.String("params", "", "Drain specific params (eg: foo=2;bar=3)")
}

func (cmd *add) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("need exactly one positional argument")
	}
	name := args[0]

	Init()

	uri := fmt.Sprintf("%s://%s/?", *cmd.scheme, *cmd.host)
	query := url.Values{}

	for _, filter := range strings.Split(*cmd.filter, ";") {
		query.Add("filter", filter)
	}

	if *cmd.format != "" {
		query.Set("format", *cmd.format)
	}

	for _, param := range strings.Split(*cmd.params, ";") {
		parts := strings.Split(param, "=")
		key, value := parts[0], parts[1]
		query.Set(key, value)
	}

	uri += query.Encode()

	fmt.Println(uri)
	err := logyard.Config.AddDrain(name, uri)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
