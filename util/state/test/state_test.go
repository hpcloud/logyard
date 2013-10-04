package test

import (
	"fmt"
	"github.com/ActiveState/logyard/util/state"
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
			&NoopRetryer{},
			nil))
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
			&NoopRetryer{},
			nil))
}

func TestRetry(t *testing.T) {
	seq := Sequence([]interface{}{
		SeqAction(state.START),
		SeqDelay(20 * time.Millisecond),
		SeqState("RUNNING|RETRYING|STARTING"),
		SeqDelay(100 * time.Millisecond),
		SeqState("FATAL"),
	})

	seq.Test(
		t,
		state.NewStateMachine(
			"DummyProcess",
			&MockProcess{
				"retry",
				time.Duration(10 * time.Millisecond),
				fmt.Errorf("exiting after 10ms"),
				nil},
			&ThriceRetryer{},
			nil))
}
