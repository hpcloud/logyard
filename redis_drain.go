package logyard

import (
	"fmt"
	"github.com/ActiveState/log"
	"github.com/vmihailenco/redis"
	"launchpad.net/tomb"
	"net"
	"stackato/server"
	"strings"
)

type RedisDrain struct {
	client *redis.Client
	log    *log.Logger
	tomb.Tomb
}

func NewRedisDrain(log *log.Logger) Drain {
	rd := &RedisDrain{}
	rd.log = log
	return rd
}

func (d *RedisDrain) Start(config *DrainConfig) {
	defer d.Done()

	// store messages under `redisKey` (redis key). if it is empty,
	// store them under that message's key.
	redisKey := config.GetParam("key", "")

	limit, err := config.GetParamInt("limit", 1500)
	if err != nil {
		d.Killf("limit key from `params` is not a number -- %s", err)
		return
	}

	database, err := config.GetParamInt("database", 0)
	if err != nil {
		d.Killf("invalid database specified: %s", err)
		return
	}

	// HACK (stackato-specific): "core" translates to the applog redis on core node
	if config.Host == "stackato-core" {
		config.Host = server.Config.CoreIP
	} else if strings.HasPrefix(config.Host, "stackato-core:") {
		config.Host = fmt.Sprintf("%s:%s",
			server.Config.CoreIP, config.Host[len("stackato-core:"):])
	}

	if err = d.connect(config.Host, int64(database)); err != nil {
		d.Kill(err)
		return
	}
	defer d.disconnect()
	c, err := NewClientGlobal(false)
	if err != nil {
		d.Kill(err)
		return
	}
	defer c.Close()

	ss, err := c.Recv(config.Filters)
	if err != nil {
		d.Kill(err)
		return
	}
	defer ss.Stop()

	for {
		select {
		case msg := <-ss.Ch:
			key := msg.Key
			if redisKey != "" {
				key = redisKey
			}
			data, err := config.FormatJSON(msg.Value)
			if err != nil {
				d.Kill(err)
				return
			}
			_, err = d.Lpushcircular(key, string(data), int64(limit))
			if err != nil {
				d.Kill(err)
				return
			}
		case <-d.Dying():
			return
		}
	}
}

func (d *RedisDrain) Stop() error {
	d.log.Info("Stopping...")
	d.Kill(nil)
	return d.Wait()
}

func (d *RedisDrain) connect(addr string, database int64) error {
	d.log.Infof("Connecting to redis %s[%d] ...", addr, database)

	// Bug #97459 -- is the redis client library faking connection for
	// the down server?
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	conn.Close()

	d.client = redis.NewTCPClient(addr, "", database)
	d.log.Infof("Connected to redis %s[%d]", addr, database)
	return nil
}

func (d *RedisDrain) disconnect() {
	d.client.Close()
}

// Lpushcircular works like LPUSH, but trims the right most element if length
// exceeds maxlen. Returns the list length before trim.  
func (d *RedisDrain) Lpushcircular(
	key string, item string, maxlen int64) (int64, error) {
	reply := d.client.LPush(key, item)
	if err := reply.Err(); err != nil {
		return -1, err
	}

	n := reply.Val()

	// Keep the length of the bounded list under check
	if n > maxlen {
		reply := d.client.LTrim(key, 0, maxlen-1)
		if err := reply.Err(); err != nil {
			return -1, err
		}
	}

	return n, nil
}
