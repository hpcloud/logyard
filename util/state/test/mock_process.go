package test

import (
	"fmt"
	"time"
)

type DummyProcess struct {
	name      string
	exitAfter time.Duration
	exitError error
	delayCh   <-chan time.Time
}

func (p *DummyProcess) Start() error {
	if p.delayCh != nil {
		return fmt.Errorf("delayCh is already set")
	}
	p.delayCh = time.After(p.exitAfter)
	return nil
}

func (p *DummyProcess) Stop() error {
	p.delayCh = nil
	return p.exitError
}

func (p *DummyProcess) Wait() error {
	<-p.delayCh
	p.delayCh = nil
	return p.exitError
}

func (p *DummyProcess) String() string {
	return fmt.Sprintf("dummy:%s", p.name)
}

func (p *DummyProcess) Logf(msg string, v ...interface{}) string {
	v = append([]interface{}{p.String()}, v...)
	return fmt.Sprintf("[%s] "+msg, v...)
}
