package drain

import (
	"github.com/fzzbt/radix/redis"
	"launchpad.net/tomb"
	"log"
	"logyard"
)

type RedisDrain struct {
	client *redis.Client
	tomb.Tomb
}

func NewRedisDrain() Drain {
	rd := &RedisDrain{}
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

	d.connect()

	c := logyard.NewClient()
	defer c.Close()

	ss, err := c.Recv(config.Filters)
	if err != nil {
		d.Kill(err)
		return
	}

	go func() {
		for {
			select {
			case msg := <-ss.Ch:
				key := msg.Key
				if redisKey != "" {
					key = redisKey
				}
				// println(key, msg.Value, limit)
				d.Lpushcircular(key, msg.Value, limit)
			case <-d.Dying():
				return
			}
		}
	}()

	d.Kill(ss.Wait())
}

func (d *RedisDrain) Stop() error {
	d.Kill(nil)
	return d.Wait()
}

func (d *RedisDrain) connect() {
	conf := redis.DefaultConfig()
	conf.Database = 0               // same database used by CC 
	conf.Address = "localhost:5454" // TODO: read from doozer
	log.Printf("Connecting to redis %s\n", conf.Address)
	d.client = redis.NewClient(conf)
}

// Lpushcircular works like LPUSH, but trims the right most element if length
// exceeds maxlen. Returns the list length before trim.  
// XXX: should this function return `reply` and/or `reply.Err`?
func (d *RedisDrain) Lpushcircular(key string, item string, maxlen int) int {
	reply := d.client.Lpush(key, item)
	if reply.Err != nil {
		panic(reply.Err)
	}

	n, err := reply.Int()
	if err != nil {
		panic(err)
	}

	// keep the length of our 'circular' list under check
	if n > maxlen {
		reply = d.client.Ltrim(key, 0, maxlen-1)
		if reply.Err != nil {
			panic(reply.Err)
		}
	}

	return n
}
