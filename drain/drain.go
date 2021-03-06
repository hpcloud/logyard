package drain

import (
	"fmt"
	"logyard"
)

type DrainType interface {
	// Start starts the drain, and returns immediately without
	// blocking.
	Start(*DrainConfig)
	Stop() error
	Wait() error
	WaitRunning() bool
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
	drain       DrainType
	name        string
	cfg         *DrainConfig
	constructor DrainConstructor
}

func NewDrainProcess(name, uri string) (*DrainProcess, error) {
	p := &DrainProcess{}

	cfg, err := ParseDrainUri(name, uri, logyard.GetConfig().DrainFormats)
	if err != nil {
		return nil, fmt.Errorf("[drain:%s] Invalid drain URI (%s): %s", name, uri, err)
	}

	p.name = name
	p.cfg = cfg

	if constructor, ok := DRAINS[cfg.Type]; ok && constructor != nil {
		p.constructor = constructor
	} else {
		return nil, fmt.Errorf("[drain:%s] Unsupported drain", name)
	}

	return p, nil
}

func (p *DrainProcess) Start() error {
	// Drains can only be started once, due to use throw-away tombs,
	// so we create them fresh.
	p.drain = p.constructor(p.name)
	go p.drain.Start(p.cfg)
	return nil
}

// WaitRunning waits until the drain is running
func (p *DrainProcess) WaitRunning() bool {
	return p.drain.WaitRunning()
}

func (p *DrainProcess) Stop() error {
	return p.drain.Stop()
}

func (p *DrainProcess) Wait() error {
	return p.drain.Wait()
}

func (p *DrainProcess) String() string {
	return fmt.Sprintf("drain:%s", p.name)
}

func (p *DrainProcess) Logf(msg string, v ...interface{}) string {
	v = append([]interface{}{p.String()}, v...)
	return fmt.Sprintf("[%s] "+msg, v...)
}
