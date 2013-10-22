package docker_events

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"io"
	"net/http"
)

type Event struct {
	Id     string `json:"id"`
	Status string `json:"status"`
	From   string `json:"from"`
	Time   int64  `json:"time"`
}

func Stream() (chan *Event, error) {
	c := http.Client{}
	res, err := c.Get("http://localhost:4243/events")
	if err != nil {
		return nil, err
	}

	ch := make(chan *Event)

	go func() {
		defer res.Body.Close()

		// Read the streaming json from the events endpoint
		// http://docs.docker.io/en/latest/api/docker_remote_api_v1.3/#monitor-docker-s-events
		d := json.NewDecoder(res.Body)
		for {
			var event Event
			if err := d.Decode(&event); err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			ch <- &event
		}

		close(ch)
	}()

	return ch, nil
}
