package stackato

import (
	"fmt"
	"github.com/ActiveState/doozer"
	"github.com/srid/doozerconfig"
)

// GetAppLogStoreRedisUri returns the URI of applog store redis instance
func GetAppLogStoreRedisUri(conn *doozer.Conn) (string, error) {
	var Config struct {
		Host string `doozer:"host"`
		Port int64  `doozer:"port"`
	}
	var rev int64
	var err error
	if rev, err = conn.Rev(); err == nil {
		cfg := doozerconfig.New(conn, rev, &Config, "/proc/cloud_controller/config/redis/")
		if err = cfg.Load(); err == nil {
			// Note: port number is hardcoded as it is not available in config.
			return fmt.Sprintf("%s:%d", Config.Host, 6464), nil
		}
	}
	return "", err
}
