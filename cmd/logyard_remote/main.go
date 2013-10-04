package main

import (
	"github.com/ActiveState/log"
	"github.com/ActiveState/logyard/cli/commands"
	"github.com/ActiveState/logyard/util/subcommand_server"
)

func main() {
	srv := subcommand_server.Server{
		commands.GetAll()}
	log.Fatal(srv.Start("127.0.0.1:8891"))
}
