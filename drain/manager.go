package drain

import (
	"fmt"
	"github.com/ActiveState/doozerconfig"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/util/retry"
	"strings"
	"sync"
	"time"
)

const configKey = "/proc/logyard/config/"

type DrainManager struct {
	mux       sync.Mutex       // mutex to protect Start/Stop
	running   map[string]Drain // map of drain instance name to drain
	stopCh    chan bool
	doozerCfg *doozerconfig.DoozerConfig
	doozerRev int64
}

func NewDrainManager() *DrainManager {
	manager := new(DrainManager)
	manager.running = make(map[string]Drain)
	manager.stopCh = make(chan bool)
	return manager
}

// Stop stops the drain manager including running drains.
func (manager *DrainManager) Stop() {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	for name, _ := range manager.running {
		manager.stopDrain(name)
	}
	// Sending on stopCh could block if DrainManager:Run().select{ ...
	// } is blocking on a mutex (via Start/Stop). A better solution
	// would be to get rid of mutexes and use channels. Until that
	// happens, we asynchronously send on stopCh. Blocking is fine,
	// because the process will be killed eventually.
	go func() {
		manager.stopCh <- true
	}()
}

// XXX: use tomb and channels to properly process start/stop events.

// StopDrain starts the drain if it is running
func (manager *DrainManager) StopDrain(drainName string) {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	manager.stopDrain(drainName)
}

// StopDrain starts the drain if it is running
func (manager *DrainManager) stopDrain(drainName string) {
	if drain, ok := manager.running[drainName]; ok {
		log.Infof("[drain:%s] Stopping drain ...\n", drainName)

		// timeout faulty drains (unlikely) from blocking the rest of
		// the logyard.
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
		log.Infof("[drain:%s] Cannot start drain (already running)", name)
		return
	}

	fmt.Printf("%+v\n", logyard.GetConfig().DrainFormats)
	cfg, err := ParseDrainUri(name, uri, logyard.GetConfig().DrainFormats)
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
			proceed := retry.Wait(
				fmt.Sprintf("[drain:%s] Drain exited abruptly -- %s", name, err))
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
	for prefix, duration := range logyard.GetConfig().RetryLimits {
		if strings.HasPrefix(name, prefix) {
			if retryLimit, err = time.ParseDuration(duration); err != nil {
				log.Error("[drain:%s] Invalid duration (%s) for drain prefix %s "+
					"-- %s -- using default value (infinite)",
					name, duration, prefix, err)
				retryLimit = time.Duration(0)
			}
			if retryLimit <= retry.RESET_AFTER {
				log.Error("[drain:%s] Invalid retry limit (%v) -- must be >%v -- "+
					"using default value (infinite)",
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
	for {
		select {
		case change := <-logyard.Config.DrainChanges:
			switch change.Type {
			case doozerconfig.DELETE:
				manager.StopDrain(change.Key)
			case doozerconfig.SET:
				manager.StopDrain(change.Key)
				manager.StartDrain(
					change.Key, logyard.Config.Drains[change.Key], NewRetryerForDrain(change.Key))
			}
		case <-manager.stopCh:
			break
		}
	}
}
