package apptail

import (
	"github.com/ActiveState/log"
	"stackato/server"
)

type Config struct {
	MaxRecordSize int   `json:"max_record_size"`
	RateLimit     int64 `json:"rate_limit"`
}

var c *server.Config

func GetConfig() *Config {
	return c.GetConfig().(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewConfig("apptail", Config{})
	if err != nil {
		log.Fatal(err)
	}
}
