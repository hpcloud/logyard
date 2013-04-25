package state

// Process is a start-able/stop-able entity not unlike an OS process
// or thread.
type Process interface {
	Start() error
	Stop() error
	Wait() error
	String() string
}
