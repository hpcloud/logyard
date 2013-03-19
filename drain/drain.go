package drain

type Drain interface {
	Start(*DrainConfig)
	Stop() error
	Wait() error
}

// DrainConstructor is a function that returns a new drain instance
type DrainConstructor func(string) Drain

// DRAINS is a map of drain type (string) to its constructur function
var DRAINS = map[string]DrainConstructor{
	"redis": NewRedisDrain,
	"tcp":   NewIPConnDrain,
	"udp":   NewIPConnDrain,
	"file":  NewFileDrain,
}
