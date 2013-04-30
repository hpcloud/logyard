package test

import (
	"fmt"
	"logyard/util/state"
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
	stm := state.NewStateMachine("Drain", p, r)
	if err := stm.SendAction(state.START); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	st := stm.GetState()
	if st.String() != "RUNNING" {
		t.Fatalf("not running yet; %v", st)
	}
	time.Sleep(150 * time.Millisecond)
	st = stm.GetState()
	if st.String() != "FATAL" {
		t.Fatalf("expecting FATAL; %v", st)
	}
	fmt.Printf(
		"as expected, exited with error:- '%v'\n",
		st.(state.Fatal).Error)
}
