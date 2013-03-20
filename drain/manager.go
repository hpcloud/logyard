package drain

import (
	"fmt"
	"github.com/ActiveState/doozerconfig"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/retry"
	"strings"
	"sync"
	"time"
)

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
		log.Infof("[drain:%s] Stopping drain ...\n", drainName)

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
			log.Errorf("[drain:%s] Unable to stop drain: %s\n", drainName, err)
		} else {
			delete(manager.running, drainName)
			log.Infof("[drain:%s] Removed drain from memory\n", drainName)
		}
	} else {
		log.Infof("[drain:%s] Drain cannot be stopped (it is not running)", drainName)
	}
}

// StartDrain starts the drain and waits for it exit.
func (manager *DrainManager) StartDrain(name, uri string, retry retry.Retryer) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	if _, ok := manager.running[name]; ok {
		log.Errorf("[drain:%s] Cannot start drain (already running)", name)
		return
	}

	cfg, err := DrainConfigFromUri(name, uri)
	if err != nil {
		log.Errorf("[drain:%s] Invalid drain URI (%s): %s", name, uri, err)
		return
	}

	var drain Drain

	if constructor, ok := DRAINS[cfg.Type]; ok && constructor != nil {
		drain = constructor(name)
	} else {
		log.Info("[drain:%s] Unsupported drain", name)
		return
	}

	manager.running[cfg.Name] = drain
	log.Infof("[drain:%s] Starting drain: %s", name, uri)
	go drain.Start(cfg)

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
				fmt.Sprintf("[drain:%s] Drain exited abruptly -- %s", name, err),
				shouldWarn)
			if !proceed {
				return
			}
			if _, ok := logyard.Config.Drains[name]; ok {
				manager.StartDrain(name, uri, retry)
			} else {
				log.Infof("[drain:%s] Not restarting because the drain was deleted recently", name)
			}
		}
	}()
}

// NewRetryerForDrain chooses 
func NewRetryerForDrain(name string) retry.Retryer {
	var retryLimit time.Duration
	var err error
	for prefix, duration := range logyard.Config.RetryLimits {
		if strings.HasPrefix(name, prefix) {
			if retryLimit, err = time.ParseDuration(duration); err != nil {
				log.Error("[drain:%s] Invalid duration (%s) for drain prefix %s "+
					"-- %s -- using default value (infinite)",
					name, duration, prefix, err)
				retryLimit = time.Duration(0)
			}
			if retryLimit <= retry.RESET_AFTER {
				log.Error("[drain:%s] Invalid retry limit (%v); must be >%v. "+
					"Using default value (infinite)",
					name, retryLimit, retry.RESET_AFTER)
				retryLimit = time.Duration(0)
			}
			break
		}
	}
	log.Infof("[drain:%s] Choosing retry limit %v", name, retryLimit)
	return retry.NewProgressiveRetryer(retryLimit)
}

func (manager *DrainManager) Run() {
	log.Infof("Found %d drains to start\n", len(logyard.Config.Drains))
	for name, uri := range logyard.Config.Drains {
		manager.StartDrain(name, uri, NewRetryerForDrain(name))
	}

	// Watch for config changes in doozer
	for change := range logyard.Config.Ch {
		switch change.Type {
		case doozerconfig.DELETE:
			manager.StopDrain(change.Key)
		case doozerconfig.SET:
			manager.StopDrain(change.Key)
			manager.StartDrain(
				change.Key, logyard.Config.Drains[change.Key], NewRetryerForDrain(change.Key))
		}
	}
}
