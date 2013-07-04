package main

import (
	"logyard/cli/commands"
	"logyard/util/subcommand"
)

func main() {
	subcommand.Parse(commands.GetAll()...)
}
