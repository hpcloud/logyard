// confdis manages JSON based configuration in redis
package confdis

import (
	"encoding/json"
	"github.com/vmihailenco/redis"
	"net"
)

type ConfDis struct {
	rootKey      string
	pubKey       string
	configStruct interface{}
	redis        *redis.Client
}

func New(addr, rootKey string, struc interface{}) *ConfDis {
	c := ConfDis{rootKey, rootKey + ":_changes", struc, nil}
	c.connect(addr)
	return &c
}

// Save saves current config onto redis.
func (c *ConfDis) Save() error {
	if data, err := json.Marshal(c.configStruct); err != nil {
		return err
	} else {
		if r := c.redis.Set(c.rootKey, string(data)); r.Err() != nil {
			return r.Err()
		}
	}
	return nil
}

func (c *ConfDis) connect(addr string) {
	// Bug #97459 -- is the redis client library faking connection for
	// the down server?
	if conn, err := net.Dial("tcp", addr); err != nil {
		panic(err)
	} else {
		conn.Close()
	}

	c.redis = redis.NewTCPClient(addr, "", 0)
}

// reload reloads the config tree from redis
func (c *ConfDis) reload() error {
	if r := c.redis.Get(c.rootKey); r.Err() != nil {
		return r.Err()
	} else {
		data := []byte(r.Val())
		json.Unmarshal(data, c.configStruct)
	}
	return nil
}
