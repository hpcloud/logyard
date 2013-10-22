package apptail

import (
	"github.com/ActiveState/log"
	"stackato/server"
	"sync"
)

var once sync.Once
var nodeid string

// LocalNodeId returns the node ID of the local node.
func LocalNodeId() string {
	once.Do(func() {
		var err error
		nodeid, err = server.LocalIP()
		if err != nil {
			log.Fatalf("Failed to determine IP addr: %v", err)
		}
		log.Info("Local Node ID: ", nodeid)
	})
	return nodeid
}
