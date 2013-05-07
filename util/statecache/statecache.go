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

// SetState caches the given state of a process in redis.
func (s *StateCache) SetState(
	name string, state state.State, rev int64) {
	allKey, thisKey := s.getKeys(name)

	log.Infof("[statecache] Caching state of %s", name)
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

// Clear clears the cache associated with the given process and
// current host.
func (s *StateCache) Clear(name string) {
	log.Infof("[statecache] Clearing state of %s", name)
	allKey, thisKey := s.getKeys(name)

	// Note that redis automatically deletes the SET if it will be
	// empty after an SREM.
	reply := s.Client.SRem(allKey, s.Host)
	if err := reply.Err(); err != nil {
		log.Errorf("Unable to clear state cache of %s in redis; %v",
			name, err)
	}

	reply2 := s.Client.Del(thisKey)
	if err := reply2.Err(); err != nil {
		log.Errorf("Unable to clear state cache of %s in redis; %v",
			name, err)
	}
}

// GetState retrieves the cached state for the given process on all
// nodes.
func (s *StateCache) GetState(name string) (map[string]string, error) {
	allKey, _ := s.getKeys(name)
	states := map[string]string{}

	reply := s.Client.SMembers(allKey)
	if err := reply.Err(); err != nil {
		return nil, err
	}
	for _, nodeip := range reply.Val() {
		reply2 := s.Client.Get(s.getKeyFor(name, nodeip))
		if err := reply2.Err(); err != nil {
			return nil, err
		}
		state := reply2.Val()
		states[nodeip] = state
	}
	return states, nil
}

func (s *StateCache) getKeys(name string) (string, string) {
	allKey := s.Prefix + name
	thisKey := allKey + ":" + s.Host
	return allKey, thisKey
}

func (s *StateCache) getKeyFor(name, node string) string {
	return s.Prefix + name + ":" + s.Host
}
