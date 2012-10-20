package drain

import (
	"fmt"
	"net/url"
)

type DrainConfig struct {
	Name    string                 // name of this particular drain instance
	Type    string                 // drain type
	Filters []string               // the messages a drain is interested in
	Params  map[string]interface{} // params specific to a drain
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
	if _, ok := AVAILABLE_DRAINS[config.Type]; !ok {
		return nil, fmt.Errorf("unknown drain type: %s", uri)
	}

	params := url.Query()

	if filters, ok := params["filter"]; ok {
		params.Del("filter")
		config.Filters = filters
	}

	config.Params = make(map[string]interface{})
	for k, v := range params {
		if len(v) == 1 {
			config.Params[k] = v[0]
		} else {
			// XXX: we have not came across this case (non-filter
			// multi-value keys), buf if we do, we must revist this
			// code. a multi-value key with a single provided value
			// would end up not treated as a list (as it would pass
			// the preceding `if` condition).
			config.Params[k] = v
		}
	}

	return &config, nil
}
