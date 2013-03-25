package cli

import (
	"fmt"
	"logyard"
	"net/url"
	"strings"
)

// AddDrain adds a drain. URI should not contain a query fragment,
// which will be constructed from the `filters` and `params`
// arguments.
func AddDrain(name, uri string, filters []string, params map[string]string) (string, error) {
	if uri == "" {
		return "", fmt.Errorf("URI cannot be empty")
	}

	if !strings.Contains(uri, "://") {
		return "", fmt.Errorf("Not an URI: %s", uri)
	}

	// Build the query string
	query := url.Values{}
	for _, filter := range filters {
		query.Add("filter", filter)
	}
	for key, value := range params {
		if key == "filter" {
			return "", fmt.Errorf("params cannot have a key called 'filter'")
		}
		query.Set(key, value)
	}

	uri += "?" + query.Encode()

	err := logyard.Config.AddDrain(name, uri)
	if err != nil {
		return "", err
	}
	return uri, err
}
