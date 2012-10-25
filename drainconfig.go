package logyard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"text/template"
)

type DrainConfig struct {
	Name    string // name of this particular drain instance
	Type    string // drain type
	Scheme  string
	Host    string             // host+port part of the uri (optional in some drains)
	Format  *template.Template // format message json using go's tempate library
	Filters []string           // the messages a drain is interested in
	Params  map[string]string  // params specific to a drain
}

// GetParam returns the corresponding param; else the default value (def)
func (c *DrainConfig) GetParam(key string, def string) string {
	if val, ok := c.Params[key]; ok {
		return val
	}
	return def
}

func (c *DrainConfig) GetParamInt(key string, def int) (int, error) {
	data := c.GetParam(key, "")
	if data == "" {
		return def, nil
	}
	var val int
	var err error
	if val, err = strconv.Atoi(data); err != nil {
		return 0, err
	}
	return val, nil
}

// FormatJSON formats the given message and returns it with a newline
func (c *DrainConfig) FormatJSON(data string) ([]byte, error) {
	if c.Format == nil {
		return []byte(data + "\n"), nil
	}
	record := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = c.Format.Execute(&buf, record)
	if err != nil {
		return nil, err
	}
	return append(buf.Bytes(), byte('\n')), nil
}

// DrainConfigFromUri constructs a DrainConfig from a drain URI spec.
// Examples:
//  - "redis://core/?max_records=1500&filter=apptail"
//  - "udp://logs.papertrailapp.com:35234/?filter=systail&filter=events"
func DrainConfigFromUri(name string, uri string) (*DrainConfig, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	config := DrainConfig{Name: name, Type: url.Scheme}
	if _, ok := DRAINS[config.Type]; !ok {
		return nil, fmt.Errorf("unknown drain type: %s", uri)
	}

	config.Scheme = url.Scheme
	config.Host = url.Host

	params := url.Query()

	// parse filters
	if filters, ok := params["filter"]; ok {
		params.Del("filter")
		config.Filters = filters
	} else {
		// default filter: all 
		config.Filters = []string{""}
	}

	if len(config.Filters) == 0 {
		panic("filters can't be empty")
	}

	// parse format
	if format, ok := params["format"]; ok {
		params.Del("format")
		tmpl, err := template.New(name).Parse(format[0])
		if err != nil {
			return nil, err
		}
		config.Format = tmpl
	}

	// assign rest of the params
	config.Params = make(map[string]string)
	for k, v := range params {
		// NOTE: multi value params are not supported.
		config.Params[k] = v[0]
	}

	return &config, nil
}
