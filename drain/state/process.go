package state

// Process is a start-able/stop-able entity not unlike an OS process
// or thread.
type Process interface {
	// Start starts the process.
	Start() error
	// Stop stop the process.
	Stop() error
	// Wait waits for the process to exit, returning an error if any.
	Wait() error
	// String returns a short printable string representation of the
	// process.
	String() string
	// Logf returns a loggable message pertaining to the given action
	// and this process.
	Logf(msg string, v ...interface{}) string
}
