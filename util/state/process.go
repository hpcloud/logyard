package state

// Process is a start-able/stop-able entity not unlike an OS process
// or thread.
type Process interface {
	// Start starts the process, and returns immediately without
	// blocking.
	Start() error
	// Stop stop the process.
	Stop() error
	// WaitRunning waits until the Start'ed process is fully running.
	// Returns false if there was an error starting.
	WaitRunning() bool
	// Wait waits for the process to exit, returning an error if any.
	Wait() error
	// String returns a short printable string representation of the
	// process.
	String() string
	// Logf returns a loggable message pertaining to the given action
	// and this process.
	Logf(msg string, v ...interface{}) string
}
