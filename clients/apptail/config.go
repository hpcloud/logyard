package apptail

import (
	"github.com/ActiveState/log"
	"stackato/server"
)

type Config struct {
	MaxRecordSize   int               `json:"max_record_size"`
	RateLimit       int64             `json:"rate_limit"`
	FileSizeLimit   int64             `json:"read_limit"`
	DefaultLogFiles map[string]string `json:"default_log_files"`
}

var c *server.Config

func GetConfig() *Config {
	return c.GetConfig().(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewConfig("apptail", Config{})
	if err != nil {
		log.Fatalf("Unable to load apptail config; %v", err)
	}
}
