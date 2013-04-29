package drain

import (
	"github.com/ActiveState/log"
	"logyard"
	"logyard/drain/state"
	"logyard/util/mapdiff"
	"logyard/util/retry"
	"strings"
	"sync"
	"time"
)

const configKey = "/proc/logyard/config/"

type DrainManager struct {
	mux    sync.Mutex // mutex to protect Start/Stop
	stopCh chan bool

	stmMap map[string]*state.StateMachine
}

func NewDrainManager() *DrainManager {
	manager := new(DrainManager)
	manager.stopCh = make(chan bool)
	manager.stmMap = make(map[string]*state.StateMachine)
	return manager
}

// Stop stops the drain manager including running drains
func (manager *DrainManager) Stop() {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	for name, _ := range manager.stmMap {
		manager.stopDrain(name)
	}
	go func() {

	}()
}

func (manager *DrainManager) StopDrain(drainName string) {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	manager.stopDrain(drainName)
}

func (manager *DrainManager) stopDrain(drainName string) {
	if drainStm, ok := manager.stmMap[drainName]; ok {
		drainStm.ActionCh <- state.STOP
		drainStm.Stop()
		delete(manager.stmMap, drainName)
	}
	// Sending on stopCh could block if DrainManager.Run().select
	// {...} is blocking on a mutex (via Start/Stop). Ideally, get rid
	// of mutexes and use channels.
	go func() {
		manager.stopCh <- true
	}()
}

func (manager *DrainManager) StartDrain(name, uri string, retry retry.Retryer) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	_, ok := manager.stmMap[name]
	if ok {
		// Stop the running drain first.
		manager.stopDrain(name)
	}
	process, err := NewDrainProcess(name, uri)
	if err != nil {
		log.Error(process.Logf("Couldn't start drain: %v", err))
		return
	}
	drainStm := state.NewStateMachine(process, retry)
	manager.stmMap[name] = drainStm

	log.Infof(process.Logf("Going to <- state.START .."))
	drainStm.ActionCh <- state.START
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
	drains := logyard.GetConfig().Drains
	log.Infof("Found %d drains to start\n", len(drains))
	for name, uri := range drains {
		manager.StartDrain(name, uri, NewRetryerForDrain(name))
	}

	// Watch for config changes in redis.
	for {
		select {
		case err := <-logyard.GetConfigChanges():
			if err != nil {
				log.Fatalf("Error re-loading config: %v", err)
			}
			log.Info("Config changed; checking drains...")
			newDrains := logyard.GetConfig().Drains
			for _, c := range mapdiff.MapDiff(drains, newDrains) {
				if c.Deleted {
					log.Infof("[Config change] Drain %s was deleted.", c.Key)
					manager.StopDrain(c.Key)
					delete(drains, c.Key)
				} else {
					log.Infof("[Config change] Drain %s was added.", c.Key)
					manager.StopDrain(c.Key)
					manager.StartDrain(
						c.Key,
						c.NewValue,
						NewRetryerForDrain(c.Key))
					drains[c.Key] = c.NewValue
				}
			}
			log.Info("[Config change] Done checking drains.")
		case <-manager.stopCh:
			break
		}
	}
}
