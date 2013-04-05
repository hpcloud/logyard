package confdis

import (
	"fmt"
	"testing"
)

type SampleConfig struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

func TestSimple(t *testing.T) {
	var config SampleConfig
	c := New(
		"localhost:6379",
		"test:sample:config",
		&config)
	c.reload()
	fmt.Println(config)
	config.Name = "primates"
	config.Users = []string{"chimp", "bonobo", "lemur"}
	c.Save()
	c.reload()
	fmt.Println(config)
}
