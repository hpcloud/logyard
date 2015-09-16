package drain

import (
	"github.com/hpcloud/zmqpubsub"
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

func TestFormat(_t *testing.T) {
	t := &DrainConfigTest{_t}
	formatEncoded := "%7B%7B.name%7D%7D%40%7B%7B.node_id%7D%7D%3A+%7B%7B.text%7D%7D"
	cfg, err := ParseDrainUri(
		"loggly", "tcp://logs.loggly.com:123/?format="+formatEncoded,
		make(map[string]string))
	if err != nil {
		t.Fatal(err)
	}
	data, err := cfg.FormatJSON(
		zmqpubsub.Message{"samplekey", `{"name":"dea", "node_id":"192", "text":"started app"}`})
	if err != nil {
		t.Fatal(err)
	}
	expected := "dea@192: started app\n"
	if string(data) != expected {
		t.Fatalf("FormatJSON returned unexpected value: `%s` -- expecting `%s`",
			string(data), expected)
	}
}

func TestRawFormat(_t *testing.T) {
	t := &DrainConfigTest{_t}
	t.Verify("loggly", "tcp://logs.loggly.com:12345/?format=raw", DrainConfig{
		Name:      "loggly",
		Type:      "tcp",
		Scheme:    "tcp",
		Host:      "logs.loggly.com:12345",
		Format:    nil,
		rawFormat: true,
		Filters:   []string{""},
		Params:    nil})
}

func TestURIConstruction(_t *testing.T) {
	t := &DrainConfigTest{_t}
	uri, err := ConstructDrainURI(
		"loggly",
		"tcp://logs.loggly.com:12345",
		[]string{"systail"},
		map[string]string{"format": "raw"})
	if err != nil {
		t.Fatal(err)
	}
	t.Verify("loggly", uri, DrainConfig{
		Name:      "loggly",
		Type:      "tcp",
		Scheme:    "tcp",
		Host:      "logs.loggly.com:12345",
		Format:    nil,
		rawFormat: true,
		Filters:   []string{"systail"},
		Params:    nil})
}

// Test library

type DrainConfigTest struct {
	*testing.T
}

func (t *DrainConfigTest) Verify(name string, uri string, config DrainConfig) {
	c, err := ParseDrainUri(name, uri, make(map[string]string))
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
	if c.rawFormat != config.rawFormat {
		t.Fatal("rawFormat didn't match")
	}
}
