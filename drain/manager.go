package drain

import (
	"fmt"
	"github.com/srid/doozerconfig"
	"log"
	"logyard"
	"logyard/stackato"
	"os"
)

// DrainConstructor is a function that returns a new drain instance
type DrainConstructor func(*log.Logger, *logyard.Client) Drain

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
	logyardClient *logyard.Client
	running       map[string]Drain // map of drain instance name to drain
}

func NewDrainManager() *DrainManager {
	client, err := logyard.NewClientGlobal()
	if err != nil {
		log.Fatal(err)
	}
	return &DrainManager{
		client,
		make(map[string]Drain)}
}

// XXX: using tomb and channels to properly process start/stop events.

// StopDrain starts the drain if it is running
func (manager *DrainManager) StopDrain(drainName string) {
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
func (manager *DrainManager) StartDrain(config *DrainConfig) {
	log := NewDrainLogger(config)

	if _, ok := manager.running[config.Name]; ok {
		// XXX: don't use drain logger here.
		log.Printf("Error: drain already exists")
		return
	}

	var drain Drain

	if constructor, ok := DRAINS[config.Type]; ok && constructor != nil {
		drain = constructor(log, manager.logyardClient)
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
	l.SetPrefix(fmt.Sprintf("-- %s/%s: ", c.Name, c.Type))
	return l
}

var Config struct {
	Drains map[string]string `doozer:"drains"`
}

func (manager *DrainManager) LoadConfig() {
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
				c, err := DrainConfigFromUri(change.Key, Config.Drains[change.Key])
				if err != nil {
					log.Printf("Error: cannot start drain %s=%s\n",
						change.Key, change.Value)
				} else {
					go manager.StartDrain(c)
				}
			}
		}
	})
}

func Run() {
	manager := NewDrainManager()
	manager.LoadConfig()

	for name, uri := range Config.Drains {
		log.Println(Config.Drains)
		c, err := DrainConfigFromUri(name, uri)
		if err != nil {
			log.Fatal(err)
		}
		go manager.StartDrain(c)
	}
}
