package main

import (
	"logyard/cli/commands"
	"logyard/util/subcommand_server"
)

func main() {
	srv := subcommand_server.SubCommandServer{
		commands.GetAll()}
	srv.Start(8891)
}
