package drain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"logyard/util/pubsub"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

type DrainConfig struct {
	Name    string
	Type    string
	Scheme  string
	Host    string // host+port part of the uri (optional in some drains)
	Path    string
	Filters []string           // Filter messages by these keys.
	Format  *template.Template // Format message json using Go's
	// template library; if
	// format==raw, send the raw
	// stream: "<key> <msg>"
	Params    map[string]string // Params specific to that drain type.
	rawFormat bool
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

func (c *DrainConfig) GetParamBool(key string, def bool) (bool, error) {
	data := c.GetParam(key, "")
	if data == "" {
		return def, nil
	}
	var val bool
	var err error
	if val, err = strconv.ParseBool(data); err != nil {
		return false, err
	}
	return val, nil
}

// FormatJSON formats the given message and returns it with a newline
func (c *DrainConfig) FormatJSON(msg pubsub.Message) ([]byte, error) {
	if c.Format == nil {
		if c.rawFormat {
			// <key> <json>
			return []byte(fmt.Sprintf("%s %s\n", msg.Key, msg.Value)), nil
		} else {
			// <json>
			return []byte(msg.Value + "\n"), nil
		}
	}
	record := make(map[string]interface{})
	err := json.Unmarshal([]byte(msg.Value), &record)
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
func DrainConfigFromUri(name string, uri string, namedFormats map[string]string) (*DrainConfig, error) {
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
	config.Path = url.Path

	// Go doesn't correctly parse file:// uris with empty <host>.
	// http://tools.ietf.org/html/rfc1738
	if url.Scheme == "file" {
		if strings.HasPrefix(url.Path, "//") {
			config.Path = url.Path[2:]
		}
	}

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

		if format[0] == "raw" {
			config.rawFormat = true
		} else {
			if value, ok := namedFormats[format[0]]; ok {
				format[0] = value
			}

			tmpl, err := template.New(name).Parse(format[0])
			if err != nil {
				return nil, err
			}
			config.Format = tmpl
		}
	}

	// assign the rest of the params
	config.Params = make(map[string]string)
	for k, v := range params {
		// NOTE: multi value params are not supported.
		config.Params[k] = v[0]
	}

	return &config, nil
}
