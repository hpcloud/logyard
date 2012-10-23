package drain

import (
	"testing"
)

func TestSimple(_t *testing.T) {
	t := &DrainConfigTest{_t}
	t.Verify("loggly", "tcp://logs.loggly.com:12345/", DrainConfig{
		Name:    "loggly",
		Type:    "tcp",
		Scheme:  "tcp",
		Host:    "logs.loggly.com:12345",
		Format:  nil,
		Filters: []string{""},
		Params:  nil})
}

func TestFilters(_t *testing.T) {
	t := &DrainConfigTest{_t}
	t.Verify("loggly", "tcp://logs.loggly.com:12345/?filter=x&filter=y", DrainConfig{
		Name:    "loggly",
		Type:    "tcp",
		Scheme:  "tcp",
		Host:    "logs.loggly.com:12345",
		Format:  nil,
		Filters: []string{"x", "y"},
		Params:  nil})
}

func TestParams(_t *testing.T) {
	t := &DrainConfigTest{_t}
	t.Verify("loggly", "tcp://logs.loggly.com:12345/?filter=x&a=foo&b=bar", DrainConfig{
		Name:    "loggly",
		Type:    "tcp",
		Scheme:  "tcp",
		Host:    "logs.loggly.com:12345",
		Format:  nil,
		Filters: []string{"x"},
		Params:  map[string]string{"a": "foo", "b": "bar"}})
}

// TODO: test Format

// Test library

type DrainConfigTest struct {
	*testing.T
}

func (t *DrainConfigTest) Verify(name string, uri string, config DrainConfig) {
	c, err := DrainConfigFromUri(name, uri)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != config.Name {
		t.Fatalf("Name didn't match")
	}
	if c.Type != config.Type {
		t.Fatalf("Type didn't match")
	}
	if c.Scheme != config.Scheme {
		t.Fatalf("Scheme didn't match")
	}
	if c.Host != config.Host {
		t.Fatalf("Host didn't match")
	}
	if c.Filters == nil {
		t.Fatalf("Filters can't be nil")
	}
	// filter slice
	if len(c.Filters) != len(config.Filters) {
		t.Fatalf("Filters len didn't match")
	}
	for idx, f := range c.Filters {
		if f != config.Filters[idx] {
			t.Fatalf("A filter didn't match")
		}
	}
	// params map
	if len(c.Params) != len(config.Params) {
		t.Fatalf("Params len didn't match")
	}
	for key, value := range c.Params {
		if value != config.Params[key] {
			t.Fatal("A param didn't match")
		}
	}
}

