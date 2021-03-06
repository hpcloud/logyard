package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"logyard"
	"logyard/drain"
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
	json    bool
	uri     string
	filters Filters
	params  Options
}

func (cmd *add) Name() string {
	return "add"
}

func (cmd *add) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.json, "json", false, "Output result as JSON")
	fs.StringVar(&cmd.uri, "uri", "", "Drain URI (eg: udp://logs.loggly.com:12345)")
	fs.Var(&cmd.filters, "filter", "Message filter")
	cmd.params = make(map[string]string)
	fs.Var(&cmd.params, "o", "Drain options (eg: -o 'limit=100' or -o 'format={{.Text}}'")
}

func (cmd *add) Run(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("need exactly one positional argument")
	}
	name := args[0]
	uri := cmd.uri

	uri, err := drain.ConstructDrainURI(name, cmd.uri, cmd.filters, cmd.params)
	if err != nil {
		return "", err
	}
	if err = logyard.AddDrain(name, uri); err != nil {
		return "", err
	}

	if cmd.json {
		data, err := json.Marshal(map[string]string{
			"name": name,
			"uri":  uri})
		return string(data), err
	} else {
		return fmt.Sprintf("Added drain %s: %s\n", name, uri), nil
	}
}
