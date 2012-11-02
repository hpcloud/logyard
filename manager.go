package logyard

import (
	"fmt"
	"github.com/srid/doozerconfig"
	"log"
	"os"
	"sync"
)

// DrainConstructor is a function that returns a new drain instance
type DrainConstructor func(*log.Logger) Drain

// DRAINS is a map of drain type (string) to its constructur function
var DRAINS = map[string]DrainConstructor{
	"redis": NewRedisDrain,
	"tcp":   NewIPConnDrain,
	"udp":   NewIPConnDrain}

type Drain interface {
	Start(*DrainConfig)
	Stop() error
	Wait() error
}

const configKey = "/proc/logyard/config/"

type DrainManager struct {
	mux       sync.Mutex       // mutex to protect Start/Stop
	running   map[string]Drain // map of drain instance name to drain
	doozerCfg *doozerconfig.DoozerConfig
	doozerRev int64
}

func NewDrainManager() *DrainManager {
	manager := new(DrainManager)
	manager.running = make(map[string]Drain)
	return manager
}

// XXX: use tomb and channels to properly process start/stop events.

// StopDrain starts the drain if it is running
func (manager *DrainManager) StopDrain(drainName string) {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	if drain, ok := manager.running[drainName]; ok {
		log.Printf("Stopping drain %s ...\n", drainName)
		// FIXME: drain.Stop() can wait till 1 second, but in the
		// event it waits more than that, we must use timeouts.
		err := drain.Stop()
		if err != nil {
			log.Printf("Error stopping drain %s: %s\n", drainName, err)
		} else {
			delete(manager.running, drainName)
			log.Printf("Removed drain %s\n", drainName)
		}
	} else {
		log.Printf("Error: drain %s is not running\n", drainName)
	}
}

// StartDrain starts the drain and waits for it exit.
func (manager *DrainManager) StartDrain(name, uri string, retry *Retryer) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if _, ok := manager.running[name]; ok {
		log.Printf("Error: drain %s is already running", name)
		return
	}

	config, err := DrainConfigFromUri(name, uri)
	if err != nil {
		log.Printf("Error parsing drain URI (%s): %s\n", uri, err)
		return
	}

	drainLog := NewDrainLogger(config)
	var drain Drain

	if constructor, ok := DRAINS[config.Type]; ok && constructor != nil {
		drain = constructor(drainLog)
	} else {
		log.Printf("unsupported drain")
		return
	}

	manager.running[config.Name] = drain
	drainLog.Printf("Starting drain with config: %+v", config)
	go drain.Start(config)

	go func() {
		err = drain.Wait()
		delete(manager.running, name)
		if err != nil {
			retry.Wait(fmt.Sprintf(
				"Drain %s exited with error -- %s", name, err))
			if _, ok := Config.Drains[name]; ok {
				manager.StartDrain(name, uri, retry)
			} else {
				log.Printf("Not restarting crashed drain %s, becase it was deleted recently", name)
			}
		}
	}()
}

func NewDrainLogger(c *DrainConfig) *log.Logger {
	l := log.New(os.Stderr, "", log.LstdFlags)
	prefix := c.Name + "--" + c.Type
	l.SetPrefix(fmt.Sprintf("[%25s] ", prefix))
	return l
}

func (manager *DrainManager) Run() {
	log.Printf("Found %d drains to start\n", len(Config.Drains))
	for name, uri := range Config.Drains {
		manager.StartDrain(name, uri, NewRetryer())
	}

	// Watch for config changes in doozer
	for change := range Config.Ch {
		switch change.Type {
		case doozerconfig.DELETE:
			manager.StopDrain(change.Key)
		case doozerconfig.SET:
			manager.StopDrain(change.Key)
			manager.StartDrain(
				change.Key, Config.Drains[change.Key], NewRetryer())
		}
	}
}
