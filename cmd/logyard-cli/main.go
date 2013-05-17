package main

import (
	"logyard/util/subcommand"
)

func main() {
	subcommand.Parse(
		new(recv),
		new(stream),
		new(list),
		new(add),
		new(delete),
		new(status))
}
