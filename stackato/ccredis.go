package stackato

import (
	"fmt"
	"github.com/ActiveState/doozer"
	"github.com/srid/doozerconfig"
)

// GetCCRedisUri returns the redis-server URI of the stackato core node.
func GetCCRedisUri(conn *doozer.Conn) (string, error) {
	var Config struct {
		Host string `doozer:"host"`
		Port int64  `doozer:"port"`
	}
	var rev int64
	var err error
	if rev, err = conn.Rev(); err == nil {
		cfg := doozerconfig.New(conn, rev, &Config, "/proc/cloud_controller/config/redis/")
		if err = cfg.Load(); err == nil {
			return fmt.Sprintf("%s:%d", Config.Host, Config.Port), nil
		}
	}
	return "", err
}
