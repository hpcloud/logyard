package drain

import (
	"fmt"
	"github.com/srid/doozerconfig"
	"log"
	"logyard/stackato"
	"os"
	"sync"
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
	mux     sync.Mutex       // mutex to protect Start/Stop
	running map[string]Drain // map of drain instance name to drain
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
		log.Printf("Stopping drain %s\n", drainName)
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
func (manager *DrainManager) StartDrain(name, uri string) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if _, ok := manager.running[name]; ok {
		log.Printf("Error: drain %s is already running", name)
		return
	}

	config, err := DrainConfigFromUri(name, uri)
	if err != nil {
		log.Printf("Error parsing drain URI: %s\n", err)
		return
	}

	dlog := NewDrainLogger(config)
	var drain Drain

	if constructor, ok := DRAINS[config.Type]; ok && constructor != nil {
		drain = constructor(dlog)
	} else {
		log.Printf("unsupported drain")
		return
	}

	manager.running[config.Name] = drain

	// Start and wait for the drain to exit.
	go func() {
		dlog.Printf("Starting drain with config: %+v", config)
		drain.Start(config)

		err := drain.Wait()
		if err != nil {
			log.Printf("Drain %s exited with error -- %s", name, err)
		}
		delete(manager.running, name)
	}()
}

func NewDrainLogger(c *DrainConfig) *log.Logger {
	l := log.New(os.Stderr, "", log.LstdFlags)
	prefix := c.Name + "--" + c.Type
	l.SetPrefix(fmt.Sprintf("[%25s] ", prefix))
	return l
}

var Config struct {
	Drains map[string]string `doozer:"drains"`
}

func (manager *DrainManager) MonitorDrainConfig() {
	conn, headRev, err := stackato.NewDoozerClient("logyard")
	if err != nil {
		log.Fatal(err)
	}

	Config.Drains = make(map[string]string)

	key := "/proc/logyard/config/"

	doozerCfg := doozerconfig.New(conn, headRev, &Config, key)
	err = doozerCfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d drains to start\n", len(Config.Drains))
	for name, uri := range Config.Drains {
		go manager.StartDrain(name, uri)
	}

	// Watch for config changes in doozer
	go doozerCfg.Monitor(key+"drains/*", func(change *doozerconfig.Change, err error) {
		if err != nil {
			log.Println("Error processing config change in doozer: %s", err)
			return
		}
		log.Printf("Config changed in doozer; %+v\n", change)
		if change.FieldName == "Drains" {
			switch change.Type {
			case doozerconfig.DELETE:
				manager.StopDrain(change.Key)
			case doozerconfig.SET:
				manager.StopDrain(change.Key)
				go manager.StartDrain(change.Key, Config.Drains[change.Key])
			}
		}
	})
}

func Run() {
	manager := NewDrainManager()
	manager.MonitorDrainConfig()
}
