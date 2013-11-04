package apptail

import (
	"github.com/ActiveState/log"
	"logyard/clients/docker_events"
	"sync"
)

const ID_LENGTH = 12

type dockerListener struct {
	waiters map[string]chan bool
	mux     sync.Mutex
}

var DockerListener *dockerListener

func init() {
	DockerListener = new(dockerListener)
	DockerListener.waiters = make(map[string]chan bool)
}

func (l *dockerListener) WaitForContainer(id string) {
	var total int
	ch := make(chan bool)
	id = id[:ID_LENGTH]

	if len(id) != ID_LENGTH {
		log.Fatalf("Invalid docker ID length: %v", len(id))
	}

	// Add a wait channel
	func() {
		l.mux.Lock()
		defer l.mux.Unlock()
		if _, ok := l.waiters[id]; ok {
			panic("already added")
		}
		l.waiters[id] = ch
		total = len(l.waiters)
	}()

	// Wait
	log.Infof("Waiting for container %v to exit (total waiters: %d)", id, total)
	<-ch
}

func (l *dockerListener) Listen() {
	for evt := range docker_events.Stream() {
		if len(evt.Id) != ID_LENGTH {
			log.Fatalf("Invalid docker ID length: %v", len(evt.Id))
		}

		// Notify container stop events by closing the appropriate ch.
		if !(evt.Status == "die" || evt.Status == "kill") {
			continue
		}
		l.mux.Lock()
		if ch, ok := l.waiters[evt.Id]; ok {
			close(ch)
			delete(l.waiters, evt.Id)
		}
		l.mux.Unlock()
	}
}
