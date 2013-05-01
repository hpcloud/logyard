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

	st := seq.Test(
		t,
		state.NewStateMachine(
			"Drain",
			&MockProcess{
				"simple",
				time.Duration(100 * time.Millisecond),
				fmt.Errorf("error after 100 milliseconds"),
				nil},
			&NoopRetryer{}))

	fmt.Printf(
		"as expected, exited with error:- '%v'\n",
		st.(state.Fatal).Error)
}
