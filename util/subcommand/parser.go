package subcommand

import (
	"flag"
	"fmt"
	"os"
)

func Parse(commands ...SubCommand) {
	scp := make(map[string]*SubCommandFlagSet, len(commands))
	for _, cmd := range commands {
		scp[cmd.Name()] = NewSubCommandFlagSet(cmd)
	}

	// Setup the usage string.
	oldUsage := flag.Usage
	flag.Usage = func() {
		oldUsage()
		for name, fs := range scp {
			fmt.Fprintf(os.Stderr, "\n# %s %s", os.Args[0], name)
			fs.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmdname := flag.Arg(0)
	if fs, ok := scp[cmdname]; ok {
		args := flag.Args()[1:]
		if output, err := fs.ParseAndRun(args); err != nil {
			fmt.Fprintf(os.Stderr, "command error: %s\n", err)
			fs.PrintDefaults()
			os.Exit(1)
		} else {
			if fs.FlagSet.Lookup("json").Value.String() == "true" {
				fmt.Print(output)
			} else {
				fmt.Printf(output)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "error: %s is not a valid command\n", cmdname)
		flag.Usage()
		os.Exit(1)
	}
}
