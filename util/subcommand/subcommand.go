// A simple sub command parser based on the flag package.
package subcommand

import (
	"flag"
)

type SubCommand interface {
	Name() string
	DefineFlags(*flag.FlagSet)
	// Run runs the subcommand with the given arguments and returns
	// the string to be printed and an error if any. If there is an
	// error, the string will (should) always be empty.
	Run([]string) (string, error)
}
