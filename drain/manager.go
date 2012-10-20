package drain

import (
	"log"
)

var AVAILABLE_DRAINS = map[string]int{"redis": 0, "tcp": 0, "udp": 0}

type Drain interface {
	Start(*DrainConfig)
	Stop()
	Wait() error
}

type DrainManager struct {
	running map[string]Drain // map of drain instance name to drain
}

func NewDrainManager() *DrainManager {
	return &DrainManager{make(map[string]Drain)}
}

// StartDrain starts the drain and waits for it exit.
func (dm *DrainManager) StartDrain(config *DrainConfig) {
	if _, ok := dm.running[config.Name]; ok {
		log.Printf("drain[%s]: drain already exists", config.Name)
		return
	}
	d := createDrain(config.Type)
	dm.running[config.Name] = d

	log.Printf("drain[%s]: Starting drain with config: %+v", config.Name, config)
	d.Start(config)

	err := d.Wait()
	if err != nil {
		log.Printf("drain[%s]: exited with error -- %s", config.Name, err)
	}

	delete(dm.running, config.Name)
}

func Run() {
	dm := NewDrainManager()

	sampleDrains := map[string]string{
		"apptail":      "redis://core/?filter=apptail&limit=1500",
		"kato_history": "redis://core/?filter=systail.kato&limit=256",
		"systail":      "redis://core/?filter=systail&limit=400",
	}
	// app logs
	// TODO: load that 1500 from config.yml:apptail/...
	for name, uri := range sampleDrains {
		c, err := DrainConfigFromUri(name, uri)
		if err != nil {
			log.Fatal(err)
		}
		go dm.StartDrain(c)
	}
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
