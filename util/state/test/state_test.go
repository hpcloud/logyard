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
	seq := Sequence([]interface{}{
		SeqAction(state.START),
		SeqDelay(20 * time.Millisecond),
		SeqState("RUNNING"),
		SeqDelay(150 * time.Millisecond),
		SeqState("FATAL"),
	})

	seq.Test(
		t,
		state.NewStateMachine(
			"DummyProcess",
			&MockProcess{
				"simple",
				time.Duration(100 * time.Millisecond),
				fmt.Errorf("error after 100 milliseconds"),
				nil},
			&NoopRetryer{}))
}

// Test stopping of a running process.
func TestStop(t *testing.T) {
	seq := Sequence([]interface{}{
		SeqAction(state.START),
		SeqDelay(20 * time.Millisecond),
		SeqState("RUNNING"),
		SeqAction(state.STOP),
		SeqDelay(20 * time.Millisecond),
		SeqState("STOPPED"),
	})

	seq.Test(
		t,
		state.NewStateMachine(
			"DummyProcess",
			&MockProcess{
				"stop",
				time.Duration(0),
				nil,
				nil},
			&NoopRetryer{}))
}

