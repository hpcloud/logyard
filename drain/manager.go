package drain

import (
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/util/mapdiff"
	"logyard/util/retry"
	"logyard/util/state"
	"logyard/util/statecache"
	"stackato/server"
	"strings"
	"sync"
	"time"
)

const configKey = "/proc/logyard/config/"

type DrainManager struct {
	mux        sync.Mutex // mutex to protect Start/Stop
	stopCh     chan bool
	stmMap     map[string]*state.StateMachine
	stateCache *statecache.StateCache
}

func NewDrainManager() *DrainManager {
	manager := new(DrainManager)
	manager.stopCh = make(chan bool)
	manager.stmMap = make(map[string]*state.StateMachine)
	manager.stateCache = &statecache.StateCache{
		"logyard:drainstatus:",
		server.LocalIPMust(),
		logyard.NewRedisClientMust(server.Config.CoreIP+":6464", 0)}
	return manager
}

// Stop stops the drain manager including running drains
func (manager *DrainManager) Stop() {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	for name, _ := range manager.stmMap {
		manager.stopDrain(name, true)
	}
}

func (manager *DrainManager) StopDrain(drainName string, clearStateCache bool) {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	manager.stopDrain(drainName, clearStateCache)
}

// stopDrain should only be called inside a mutex.
func (manager *DrainManager) stopDrain(drainName string, clearStateCache bool) {
	if drainStm, ok := manager.stmMap[drainName]; ok {
		if err := drainStm.SendAction(state.STOP); err != nil {
			log.Fatalf("Failed to stop drain %s; %v", drainName, err)
		}
		drainStm.Stop()
		delete(manager.stmMap, drainName)
		if clearStateCache {
			manager.stateCache.Clear(drainName)
		}
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

	var stateChangeFn state.StateChangedFn

	_, exists := manager.stmMap[name]
	if exists {
		// Stop the running drain first.
		manager.stopDrain(name, false)
	} else {
		stateChangeFn = func(state state.State, rev int64) {
			manager.stateCache.SetState(name, state, rev)
		}
	}

	process, err := NewDrainProcess(name, uri)
	if err != nil {
		log.Error(process.Logf("Couldn't create drain: %v", err))
		return
	}
	drainStm := state.NewStateMachine("Drain", process, retry, stateChangeFn)
	manager.stmMap[name] = drainStm

	if err = drainStm.SendAction(state.START); err != nil {
		log.Fatalf("Failed to start drain %s; %v", name, err)
	}
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
	iteration := 0
	drains := logyard.GetConfig().Drains
	log.Infof("Found %d drains to start\n", len(drains))
	for name, uri := range drains {
		manager.StartDrain(name, uri, NewRetryerForDrain(name))
	}

	// Watch for config changes in redis.
	for {
		iteration += 1
		prefix := fmt.Sprintf("CONFIG.%d", iteration)

		select {
		case err := <-logyard.GetConfigChanges():
			if err != nil {
				log.Fatalf("Error re-loading config: %v", err)
			}
			log.Infof(
				"[%s] checking drains after a config change...",
				prefix)
			newDrains := logyard.GetConfig().Drains
			for _, c := range mapdiff.MapDiff(drains, newDrains) {
				if c.Deleted {
					log.Infof("[%s] Drain %s was deleted.", prefix, c.Key)
					manager.StopDrain(c.Key, true)
					delete(drains, c.Key)
				} else {
					log.Infof("[%s] Drain %s was added.", prefix, c.Key)
					manager.StopDrain(c.Key, false)
					manager.StartDrain(
						c.Key,
						c.NewValue,
						NewRetryerForDrain(c.Key))
					drains[c.Key] = c.NewValue
				}
			}
			log.Infof("[%s] Done checking drains.", prefix)
		case <-manager.stopCh:
			break
		}
	}
}
