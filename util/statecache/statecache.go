// statecache provides a simple way to cache the current status in
// redis.
package statecache

import (
	"github.com/ActiveState/log"
	"github.com/vmihailenco/redis"
	"logyard/util/state"
)

type StateCache struct {
	Prefix string // redis key prefix to use for cache entries
	Host   string // host identifier, to identifier state values from the current host
	Client *redis.Client
}

func (s *StateCache) SetState(
	name string, state state.State, rev int64) {
	allKey := s.Prefix + name
	thisKey := allKey + ":" + s.Host

	reply := s.Client.SAdd(allKey, s.Host)
	if err := reply.Err(); err != nil {
		log.Errorf("Unable to cache state of %s in redis; %v",
			name, err)
		return
	}
	reply2 := s.Client.Set(thisKey, state.String())
	if err := reply2.Err(); err != nil {
		log.Errorf("Unable to cache state of %s in redis; %v",
			name, err)
	}
}
