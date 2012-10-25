package logyard

import (
	"github.com/fzzbt/radix/redis"
	"launchpad.net/tomb"
	"log"
	"logyard/stackato"
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

	// HACK (stackato-specific): "core" translates to the cc redis
	if config.Host == "core" {
		ccredishost, err := stackato.GetCCRedisUri(Config.Doozer)
		if err != nil {
			d.Killf("cannot determine cc redis uri: %s", err)
			return
		}
		config.Host = ccredishost
	}

	d.connect(config.Host, database)
	defer d.disconnect()
	c, err := NewClientGlobal()
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
			_, err = d.Lpushcircular(key, string(data), limit)
			if err != nil {
				d.Kill(err)
				return
			}
		case <-d.Dying():
			return
		}
	}

	d.log.Println("Exiting")
}

func (d *RedisDrain) Stop() error {
	d.log.Println("Stopping...")
	d.Kill(nil)
	return d.Wait()
}

func (d *RedisDrain) connect(addr string, database int) {
	conf := redis.DefaultConfig()
	conf.Database = database
	conf.Address = addr
	d.log.Printf("Connecting to redis %s[%d] ...", conf.Address, conf.Database)
	d.client = redis.NewClient(conf)
	d.log.Printf("Connected to redis %s", conf.Address)
}

func (d *RedisDrain) disconnect() {
	d.client.Close()
}

// Lpushcircular works like LPUSH, but trims the right most element if length
// exceeds maxlen. Returns the list length before trim.  
func (d *RedisDrain) Lpushcircular(key string, item string, maxlen int) (int, error) {
	reply := d.client.Lpush(key, item)
	if reply.Err != nil {
		return -1, reply.Err
	}

	n, err := reply.Int()
	if err != nil {
		return -1, err
	}

	// Keep the length of the bounded list under check
	if n > maxlen {
		reply = d.client.Ltrim(key, 0, maxlen-1)
		if reply.Err != nil {
			return -1, reply.Err
		}
	}

	return n, nil
}