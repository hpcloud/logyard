package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Event struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

type Config struct {
	Hostname string
}

type NetworkSettings struct {
	IpAddress   string
	PortMapping map[string]map[string]string
}

type Container struct {
	Id              string
	Image           string
	Config          *Config
	NetworkSettings *NetworkSettings
}

func inspectContainer(id string, c http.Client) *Container {
	// Use the container id to fetch the container json from the Remote API
	// http://docs.docker.io/en/latest/api/docker_remote_api_v1.4/#inspect-a-container
	res, err := c.Get("http://localhost:4243/containers/" + id + "/json")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		d := json.NewDecoder(res.Body)

		var container Container
		if err = d.Decode(&container); err != nil {
			log.Fatal(err)
		}
		return &container
	}
	return nil
}

func notify(container *Container) {
	settings := container.NetworkSettings

	if settings != nil && settings.PortMapping != nil {
		// I only care about Tcp ports but you can also view Udp mappings
		if ports, ok := settings.PortMapping["Tcp"]; ok {

			log.Printf("Ip address allocated for: %s", container.Id)

			// Log the public and private port mappings
			for privatePort, publicPort := range ports {
				// I am just writing to stdout but you can use this information to update hipache, redis, etc...
				log.Printf("%s -> %s", privatePort, publicPort)
			}
		}
	}
}

func main() {
	c := http.Client{}
	res, err := c.Get("http://localhost:4243/events")
	if err != nil {
		log.Fatal(err)
	}
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
		if event.Status == "start" {
			// We only want to inspect the container if it has started
			if container := inspectContainer(event.Id, c); container != nil {
				notify(container)
			}
		}
	}
}
