package logyard

import (
	"github.com/ActiveState/log"
	"github.com/vmihailenco/redis"
	"net"
)

// NewRedisClient connects to redis after ensuring that the server is
// indeed running.
func NewRedisClient(addr string, database int64) (*redis.Client, error) {
	// Bug #97459 -- is the redis client library faking connection for
	// the down server?
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn.Close()

	return redis.NewTCPClient(addr, "", database), nil
}

func NewRedisClientMust(addr string, database int64) *redis.Client {
	client, err := NewRedisClient(addr, database)
	if err != nil {
		log.Fatalf("Unable to connect to redis; %v", err)
	}
	return client
}
