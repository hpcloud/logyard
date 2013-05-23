package commands

import (
	"logyard/util/subcommand"
)

// GetAll returns all subcommands defined in this package.
func GetAll() []subcommand.SubCommand {
	return []subcommand.SubCommand{
		new(recv),
		new(stream),
		new(list),
		new(add),
		new(delete),
		new(status)}
}
