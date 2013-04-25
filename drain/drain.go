package drain

import (
	"fmt"
	"logyard"
)

type DrainType interface {
	Start(*DrainConfig)
	Stop() error
	Wait() error
}

// DrainConstructor is a function that returns a new drain instance
type DrainConstructor func(string) DrainType

// DRAINS is a map of drain type (string) to its constructur function
var DRAINS = map[string]DrainConstructor{
	"redis": NewRedisDrain,
	"tcp":   NewIPConnDrain,
	"udp":   NewIPConnDrain,
	"file":  NewFileDrain,
}

type DrainProcess struct {
	drain DrainType
	cfg   *DrainConfig
}

func NewDrainProcess(name, uri string) (*DrainProcess, error) {
	p := &DrainProcess{}

	cfg, err := ParseDrainUri(name, uri, logyard.GetConfig().DrainFormats)
	if err != nil {
		return nil, fmt.Errorf("[drain:%s] Invalid drain URI (%s): %s", name, uri, err)
	}

	p.cfg = cfg

	if constructor, ok := DRAINS[cfg.Type]; ok && constructor != nil {
		p.drain = constructor(name)
	} else {
		return nil, fmt.Errorf("[drain:%s] Unsupported drain", name)
	}

	return p, nil
}

func (p *DrainProcess) Start() error {
	go p.drain.Start(p.cfg)
	return nil
}

func (p *DrainProcess) Stop() error {
	return p.drain.Stop()
}

func (p *DrainProcess) Wait() error {
	return p.drain.Wait()
}
