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

func (d *RedisDrain) Start(config DrainConfig) {
	d.connect()

	defer d.Done()

	c := logyard.NewClient()
	defer c.Close()

	ss, err := c.Recv(config.Filters)

	if err != nil {
		d.Kill(err)
		return
	}

	var redisKey string

	if redisKeyInterface, ok := config.Params["key"]; ok {
		if redisKey, ok = redisKeyInterface.(string); !ok {
			d.Killf("redis key from `params` is of wrong type; expecting string")
			return
		}
	} else {
		d.Killf("a redis key must be specified in `params`")
		return
	}

	// TODO: read from config
	listMaxLen := 1500

	go func() {
		for {
			select {
			case msg := <-ss.Ch:
				d.Lpushcircular(redisKey, msg.Value, listMaxLen)
			case <-d.Dying():
				return
			}
		}
	}()

	d.Kill(ss.Wait())
}

func (d *RedisDrain) Stop() {
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
