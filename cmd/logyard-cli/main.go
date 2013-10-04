package main

import (
	"github.com/ActiveState/logyard/cli/commands"
	"github.com/ActiveState/logyard/util/subcommand"
)

func main() {
	subcommand.Parse(commands.GetAll()...)
}
