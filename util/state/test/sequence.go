package test

import (
	"fmt"
	"logyard/util/state"
	"testing"
	"time"
)

type SeqAction int

type SeqState string

type SeqDelay time.Duration

// Sequence is a test abstraction to perform a variety of actions, and
// check state equality after certain delays, on a state machine.
type Sequence []interface{}

// Test runs the sequence performing the necessary state equality.
func (s Sequence) Test(t *testing.T, m *state.StateMachine) state.State {
	for idx, e := range s {
		switch element := e.(type) {
		case SeqAction:
			if err := m.SendAction(int(element)); err != nil {
				t.Fatal(err)
			}
		case SeqState:
			st := m.GetState()
			if st.String() != string(element) {
				t.Fatalf("[%d/%d] expected %s; but %v",
					idx+1, len(s), element, st)
			}
		case SeqDelay:
			time.Sleep(time.Duration(element))
		}
	}
	// Return the final state which might not be the same as the
	// resultant state of the final step in the sequence.
	st := m.GetState()

	switch st2 := st.(type) {
	case state.Fatal:
		fmt.Printf(
			"Final sequence state: FATAL <%v>\n",
			st2.Error)
	default:
		fmt.Printf("Final sequence state: %s\n", st2)
	}
	return st
}
