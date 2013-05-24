// A simple sub command parser based on the flag package.
package subcommand

import (
	"flag"
	"fmt"
	"os"
)

type SubCommand interface {
	Name() string
	DefineFlags(*flag.FlagSet)
	// Run runs the subcommand with the given arguments and returns
	// the string to be printed and an error if any. If there is an
	// error, the string will (should) always be empty.
	Run([]string) (string, error)
}

type SubCommandParser struct {
	cmd SubCommand
	fs  *flag.FlagSet
}

func Parse(commands ...SubCommand) {
	scp := make(map[string]*SubCommandParser, len(commands))
	for _, cmd := range commands {
		name := cmd.Name()
		scp[name] = &SubCommandParser{cmd, flag.NewFlagSet(name, flag.ExitOnError)}
		cmd.DefineFlags(scp[name].fs)
	}

	oldUsage := flag.Usage
	flag.Usage = func() {
		oldUsage()
		for name, sc := range scp {
			fmt.Fprintf(os.Stderr, "\n# %s %s", os.Args[0], name)
			sc.fs.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmdname := flag.Arg(0)
	if sc, ok := scp[cmdname]; ok {
		sc.fs.Parse(flag.Args()[1:])
		if output, err := sc.cmd.Run(sc.fs.Args()); err != nil {
			fmt.Fprintf(os.Stderr, "command error: %s\n", err)
			sc.fs.PrintDefaults()
			os.Exit(1)
		} else {
			fmt.Printf(output)
		}
	} else {
		fmt.Fprintf(os.Stderr, "error: %s is not a valid command\n", cmdname)
		flag.Usage()
		os.Exit(1)
	}
}
