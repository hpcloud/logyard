package state

import (
	"fmt"
	"github.com/ActiveState/log"
	"testing"
	"time"
)

// Test the simplest case: start with no retrying, catch expected
// error exit.
func TestSimple(t *testing.T) {
	r := &NoopRetryer{}
	p := &DummyProcess{
		"simple",
		time.Duration(100 * time.Millisecond),
		fmt.Errorf("error after 100 milliseconds"), nil}
	stm := NewStateMachine("Drain", p, r)
	if err := stm.SendAction(START); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if stm.state.String() != "RUNNING" {
		t.Fatalf("not running yet; %v", stm.state)
	}
	time.Sleep(150 * time.Millisecond)
	if stm.state.String() != "FATAL" {
		t.Fatalf("expecting FATAL; %v", stm.state)
	}
	fmt.Printf(
		"as expected, exited with error:- '%v'\n",
		stm.state.(Fatal).Error)
}

// Test library.

type NoopRetryer struct{}

func (retry *NoopRetryer) Wait(msg string) bool {
	log.Errorf("%s -- never retrying.", msg)
	return false
}

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
