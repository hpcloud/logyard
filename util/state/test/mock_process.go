package test

import (
	"fmt"
	"time"
)

type MockProcess struct {
	name      string
	exitAfter time.Duration
	exitError error
	delayCh   <-chan time.Time
}

func (p *MockProcess) Start() error {
	if p.delayCh != nil {
		return fmt.Errorf("delayCh is already set")
	}
	if p.exitAfter > time.Duration(0) {
		p.delayCh = time.After(p.exitAfter)
	}
	return nil
}

func (p *MockProcess) Stop() error {
	p.delayCh = nil
	return p.exitError
}

func (p *MockProcess) Wait() error {
	if p.exitAfter > time.Duration(0) {
		<-p.delayCh
		p.delayCh = nil
	} else {
		// Block forever
		<-make(chan bool)
	}
	return p.exitError
}

func (p *MockProcess) String() string {
	return fmt.Sprintf("dummy:%s", p.name)
}

func (p *MockProcess) Logf(msg string, v ...interface{}) string {
	v = append([]interface{}{p.String()}, v...)
	return fmt.Sprintf("[%s] "+msg, v...)
}
