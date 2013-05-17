package apptail

import (
	"github.com/ActiveState/log"
	"stackato/server"
)

type Config struct {
	MaxRecordSize int `json:"max_record_size"`
}

var c *server.GroupConfig

func GetConfig() *Config {
	return c.Config.(*Config)
}

func LoadConfig() {
	server.Init()

	var err error
	c, err = server.NewGroupConfig("apptail", Config{})
	if err != nil {
		log.Fatal(err)
	}
}
