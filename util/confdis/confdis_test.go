package confdis

import (
	"testing"
	"time"
)

type SampleConfig struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
	Meta  struct {
		Researcher string `json:"researcher"`
		Grant      int    `json:"grant"`
	} `json:"meta"`
}

func TestSimple(t *testing.T) {
	var config SampleConfig
	c, err := New(
		"localhost:6379",
		"test:confdis:simple",
		&config)
	if err != nil {
		t.Fatal(err)
	}
	config.Name = "primates"
	config.Users = []string{"chimp", "bonobo", "lemur"}
	config.Meta.Researcher = "Jane Goodall"
	config.Meta.Grant = 1200
	c.Save()
}

func TestChangeNotification(t *testing.T) {
	// Seed data, using the first client
	var config SampleConfig
	c, err := New(
		"localhost:6379",
		"test:confdis:notify",
		&config)
	if err != nil {
		t.Fatal(err)
	}
	config.Name = "primates-changes"
	config.Users = []string{"chimp", "bonobo", "lemur"}
	config.Meta.Researcher = "Jane Goodall"
	config.Meta.Grant = 1200
	c.Save()
	// Allow reasonable delay for network/redis latency
	time.Sleep(time.Duration(100 * time.Millisecond))

	// Second client
	var config2 SampleConfig
	if _, err = New(
		"localhost:6379",
		"test:confdis:notify",
		&config2); err != nil {
		t.Fatal(err)
	}
	if config2.Meta.Researcher != "Jane Goodall" {
		t.Fatal("different value")
	}

	// Trigger a change via the first client
	config.Meta.Researcher = "Francine Patterson"
	c.Save()
	// Allow reasonable delay for network/redis latency
	time.Sleep(time.Duration(100 * time.Millisecond))

	// Second client must get notified of that change
	if config2.Meta.Researcher != "Francine Patterson" {
		t.Fatal("did not receive change")
	}
}
