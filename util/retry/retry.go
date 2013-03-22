package retry

type Retryer interface {
	// Wait appropriately waits until next try
	Wait(msg string) bool
}
