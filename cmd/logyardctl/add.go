package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
	"net/url"
	"strings"
)

// Filters is a slice of message filters
type Filters []string

func (f *Filters) String() string {
	return fmt.Sprint(*f)
}

func (f *Filters) Set(value string) error {
	*f = append(*f, value)
	return nil
}

// Options are drain specific options (ssh's -o style)
type Options map[string]string

func (o *Options) String() string {
	return fmt.Sprintf("%+v", map[string]string(*o))
}

func (o *Options) Set(value string) error {
	if value == "" {
		// default: no options
		return nil
	}
	parts := strings.FieldsFunc(value, func(c rune) bool { return c == '=' })
	if len(parts) != 2 {
		return fmt.Errorf("options must be of the `key=value` format")
	}
	key, value := parts[0], parts[1]
	if _, ok := (*o)[key]; ok {
		return fmt.Errorf("duplication option '%s' specified", key)
	}
	(*o)[key] = value
	return nil
}

// Example:
//  .. add -uri redis://core -filter systail.kato -o limit=200 -o key=kato_history kato_history
type add struct {
	uri     string
	filters Filters
	params  Options
}

func (cmd *add) Name() string {
	return "add"
}

func (cmd *add) DefineFlags(fs *flag.FlagSet) {
	fs.StringVar(&cmd.uri, "uri", "", "Drain URI (eg: udp://logs.loggly.com:12345)")
	fs.Var(&cmd.filters, "filter", "Message filter")
	cmd.params = make(map[string]string)
	fs.Var(&cmd.params, "o", "Drain options (eg: -o 'limit=100' or -o 'format={{.Text}}'")
}

func (cmd *add) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("need exactly one positional argument")
	}
	name := args[0]
	uri := cmd.uri
	if uri == "" {
		return fmt.Errorf("need -uri")
	}

	if !strings.Contains(uri, "://") {
		return fmt.Errorf("not an URI: %s", uri)
	}

	query := url.Values{}

	for _, filter := range cmd.filters {
		query.Add("filter", filter)
	}

	for key, value := range cmd.params {
		query.Set(key, value)
	}

	uri += "?" + query.Encode()

	fmt.Printf("Added drain %s: %s\n", name, uri)

	Init()
	err := logyard.Config.AddDrain(name, uri)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
