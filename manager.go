package logyard

import (
	"fmt"
	"github.com/ActiveState/doozerconfig"
	"github.com/ActiveState/log"
	"logyard/retry"
	"strings"
	"sync"
	"time"
)

// DrainConstructor is a function that returns a new drain instance
type DrainConstructor func(string) Drain

// DRAINS is a map of drain type (string) to its constructur function
var DRAINS = map[string]DrainConstructor{
	"redis": NewRedisDrain,
	"tcp":   NewIPConnDrain,
	"udp":   NewIPConnDrain,
	"file":  NewFileDrain,
}

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
		log.Infof("[%s] Stopping drain ...\n", drainName)

		// drain.Stop is expected to stop in 1s, but a known bug
		// (#96008) causes certain drains to hang. workaround it using
		// timeouts. 
		done := make(chan error)
		go func() {
			done <- drain.Stop()
		}()
		var err error
		select {
		case err = <-done:
			break
		case <-time.After(5 * time.Second):
			log.Fatalf("Error: expecting drain %s to stop in 1s, "+
				"but it is taking more than 5s; exiting now and "+
				"awaiting supervisord restart.", drainName)
		}

		if err != nil {
			log.Errorf("[%s] Unable to stop drain: %s\n", drainName, err)
		} else {
			delete(manager.running, drainName)
			log.Infof("[%s] Removed drain from memory\n", drainName)
		}
	} else {
		log.Infof("[%s] Drain cannot be stopped (it is not running)", drainName)
	}
}

// StartDrain starts the drain and waits for it exit.
func (manager *DrainManager) StartDrain(name, uri string, retry retry.Retryer) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if _, ok := manager.running[name]; ok {
		log.Errorf("[%s] Cannot start drain (already running)", name)
		return
	}

	config, err := DrainConfigFromUri(name, uri)
	if err != nil {
		log.Errorf("[%s] Invalid drain URI (%s): %s", name, uri, err)
		return
	}

	var drain Drain

	if constructor, ok := DRAINS[config.Type]; ok && constructor != nil {
		drain = constructor(name)
	} else {
		log.Info("[%s] Unsupported drain", name)
		return
	}

	manager.running[config.Name] = drain
	log.Infof("[%s] Starting drain: %+v", name, config)
	go drain.Start(config)

	go func() {
		err = drain.Wait()
		delete(manager.running, name)
		if err != nil {
			// HACK: apptail.* drains should not log WARN or ERROR
			// records. ideally, make this configurable in drain URI
			// arguments (eg: tcp://google.com:12345?warn=false);
			// doing so will require changes to cloud_controller/kato
			// (the ruby code).
			shouldWarn := !strings.HasPrefix(name, "appdrain.")

			proceed := retry.Wait(
				fmt.Sprintf("[%s] Drain exited abruptly: %s", name, err),
				shouldWarn)
			if !proceed {
				return
			}
			if _, ok := Config.Drains[name]; ok {
				manager.StartDrain(name, uri, retry)
			} else {
				log.Infof("[%s] Not restarting because the drain was deleted recently", name)
			}
		}
	}()
}

// chooseRetryer chooses an appropriate retryer for the given drain
// name.
func chooseRetryer(name string) retry.Retryer {
	if strings.HasPrefix(name, "tmp.") {
		// "tmp" drains -- such as 'kato tail' -- need not be retried
		// infinitely.
		return retry.NewFiniteRetryer()
	}
	return retry.NewInfiniteRetryer()
}

func (manager *DrainManager) Run() {
	log.Infof("Found %d drains to start\n", len(Config.Drains))
	for name, uri := range Config.Drains {
		manager.StartDrain(name, uri, chooseRetryer(name))
	}

	// Watch for config changes in doozer
	for change := range Config.Ch {
		switch change.Type {
		case doozerconfig.DELETE:
			manager.StopDrain(change.Key)
		case doozerconfig.SET:
			manager.StopDrain(change.Key)
			manager.StartDrain(
				change.Key, Config.Drains[change.Key], chooseRetryer(change.Key))
		}
	}
}
