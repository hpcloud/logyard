package retry

type Retryer interface {
	// Wait appropriately waits until next try (exponential backoff delay)
	Wait(msg string, shouldWarn bool) bool
}
