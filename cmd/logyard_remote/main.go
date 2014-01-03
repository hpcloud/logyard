package main

import (
	"github.com/ActiveState/log"
	"logyard/cli/commands"
	"logyard/util/subcommand_server"
	"stackato/server"
)

func main() {
	srv := subcommand_server.Server{
		commands.GetAll()}
	server.MarkRunning("logyard_remote")
	log.Fatal(srv.Start("127.0.0.1:8891"))
}
