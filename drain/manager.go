package drain

import (
	"log"
)

// DrainConstructor is a function that returns a new drain instance
type DrainConstructor func() Drain

// DRAINS is a map of drain type (string) to its constructur function
var DRAINS = map[string]DrainConstructor{
	"redis": NewRedisDrain,
	"tcp":   nil, // TODO
	"udp":   nil}

type Drain interface {
	Start(*DrainConfig)
	Stop() error
	Wait() error
}

type DrainManager struct {
	running map[string]Drain // map of drain instance name to drain
}

func NewDrainManager() *DrainManager {
	return &DrainManager{make(map[string]Drain)}
}

// StartDrain starts the drain and waits for it exit.
func (manager *DrainManager) StartDrain(config *DrainConfig) {
	if _, ok := manager.running[config.Name]; ok {
		log.Printf("drain[%s]: drain already exists", config.Name)
		return
	}

	var drain Drain

	if constructor, ok := DRAINS[config.Type]; ok && constructor != nil {
		drain = constructor()
	} else {
		log.Printf("unsupported drain %s for %s", config.Type, config.Name)
		return
	}

	manager.running[config.Name] = drain

	log.Printf("drain[%s]: Starting drain with config: %+v", config.Name, config)
	drain.Start(config)

	err := drain.Wait()
	if err != nil {
		log.Printf("drain[%s]: exited with error -- %s", config.Name, err)
	}

	delete(manager.running, config.Name)
}

func Run() {
	manager := NewDrainManager()

	sampleDrains := map[string]string{
		"apptail":      "redis://core/?filter=apptail&limit=1500",
		"kato_history": "redis://core/?filter=systail.kato&limit=256",
		"systail":      "redis://core/?filter=systail&limit=400",
	}
	for name, uri := range sampleDrains {
		c, err := DrainConfigFromUri(name, uri)
		if err != nil {
			log.Fatal(err)
		}
		go manager.StartDrain(c)
	}
}
