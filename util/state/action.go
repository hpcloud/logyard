package state

// Actions on state
const (
	START = iota
	STOP
)

func getActionString(action int) string {
	switch action {
	case START:
		return "START"
	case STOP:
		return "STOP"
	}
	panic("unreachable")
}
