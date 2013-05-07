package main

import (
	"logyard"
	"logyard/util/subcommand"
)

func Init(name string) {
	logyard.Init("logyard-cli:"+name, false)
}

func main() {
	subcommand.Parse(
		new(recv),
		new(stream),
		new(list),
		new(add),
		new(delete),
		new(status))
}
