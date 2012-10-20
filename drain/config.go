package drain

import (
	"fmt"
	"net/url"
	"strconv"
)

type DrainConfig struct {
	Name    string            // name of this particular drain instance
	Type    string            // drain type
	Filters []string          // the messages a drain is interested in
	Params  map[string]string // params specific to a drain
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

	params := url.Query()

	if filters, ok := params["filter"]; ok {
		params.Del("filter")
		config.Filters = filters
	}

	config.Params = make(map[string]string)
	for k, v := range params {
		// NOTE: multi value params are not supported.
		config.Params[k] = v[0]
	}

	return &config, nil
}
