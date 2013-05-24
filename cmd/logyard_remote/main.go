package main

import (
	"github.com/ActiveState/log"
	"logyard/cli/commands"
	"logyard/util/subcommand_server"
)

func main() {
	srv := subcommand_server.SubCommandServer{
		commands.GetAll()}
	log.Fatal(srv.Start("127.0.0.1:8891"))
}
