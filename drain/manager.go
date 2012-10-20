package drain

import (
	"fmt"
	"log"
	"os"
)

// DrainConstructor is a function that returns a new drain instance
type DrainConstructor func(*log.Logger) Drain

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
	log := NewDrainLogger(config)

	if _, ok := manager.running[config.Name]; ok {
		log.Printf("drain already exists")
		return
	}

	var drain Drain

	if constructor, ok := DRAINS[config.Type]; ok && constructor != nil {
		drain = constructor(log)
	} else {
		log.Printf("unsupported drain")
		return
	}

	manager.running[config.Name] = drain

	log.Printf("Starting drain with config: %+v", config)
	drain.Start(config)

	err := drain.Wait()
	if err != nil {
		log.Printf("Exited with error -- %s", err)
	}

	delete(manager.running, config.Name)
}

func NewDrainLogger(c *DrainConfig) *log.Logger {
	l := log.New(os.Stderr, "", log.LstdFlags)
	l.SetPrefix(fmt.Sprintf("-- %s (%s drain) -- ", c.Name, c.Type))
	return l
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
