package main

import (
	"logyard/cmd/logyardctl/subcommand"
)

func main() {
	subcommand.Parse(
		new(recv),
		new(list))
}
