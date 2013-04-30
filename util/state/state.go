package state

type State interface {
	// Transition transitions to another state based on the given action.
	Transition(action int, rev int64) State
	String() string
}
