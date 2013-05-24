package subcommand

import (
	"flag"
)

type SubCommandFlagSet struct {
	cmd SubCommand
	*flag.FlagSet
}

func NewSubCommandFlagSet(cmd SubCommand) *SubCommandFlagSet {
	name := cmd.Name()
	fs := &SubCommandFlagSet{
		cmd, flag.NewFlagSet(name, flag.ExitOnError)}
	cmd.DefineFlags(fs.FlagSet)
	return fs
}

func (fs *SubCommandFlagSet) ParseAndRun(args []string) (string, error) {
	fs.Parse(args)
	return fs.cmd.Run(fs.Args())
}
