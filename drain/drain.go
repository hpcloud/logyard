package drain

import (
	"log"
)

type DrainConfig struct {
	Name    string                 // name of this particular drain instance
	Type    string                 // drain type
	Filters []string               // the messages a drain is interested in
	Params  map[string]interface{} // params specific to a drain
}

type Drain interface {
	Start(DrainConfig)
	Stop()
	Wait() error
}

type DrainManager struct {
	running map[string]Drain // map of drain instance name to drain
}

func NewDrainManager() *DrainManager {
	return &DrainManager{make(map[string]Drain)}
}

func (dm *DrainManager) StartDrain(config DrainConfig) {
	if _, ok := dm.running[config.Name]; ok {
		log.Printf("drain[%s]: drain already exists", config.Name)
		return
	}
	d := createDrain(config.Type)
	dm.running[config.Name] = d

	log.Printf("drain[%s]: starting drain", config.Name)
	d.Start(config)

	err := d.Wait()
	if err != nil {
		log.Printf("drain[%s]: exited with error -- %s", config.Name, err)
	}
}

func Run() {
	dm := NewDrainManager()
	// sample drain: system logs
	go dm.StartDrain(DrainConfig{
		Name:    "redis-system-logs",
		Type:    "redis",
		Filters: []string{"systail"},
		Params:  map[string]interface{}{"key": "log.system"}})

	go dm.StartDrain(DrainConfig{
		Name:    "redis-kato-invokations",
		Type:    "redis",
		Filters: []string{"systail.kato"},
		Params:  map[string]interface{}{"key": "log.kato"}})
}

func createDrain(drainType string) Drain {
	// TODO: use reflection 
	if drainType == "redis" {
		return NewRedisDrain()
	} else {
		panic("unknown drain type " + drainType)
	}
	return nil
}
