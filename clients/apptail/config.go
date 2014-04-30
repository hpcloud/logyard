package apptail

import (
	"logyard/clients/common"
	"stackato/server"
)

type Config struct {
	MaxRecordSize int    `json:"max_record_size"`
	RateLimit     uint16 `json:"rate_limit"`
	FileSizeLimit int64  `json:"read_limit"`
}

var c *server.Config

func GetConfig() *Config {
	return c.GetConfig().(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewConfig("apptail", Config{})
	if err != nil {
		common.Fatal("Unable to load apptail config; %v", err)
	}
}
