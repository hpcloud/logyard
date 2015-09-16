package main

import (
	"github.com/hpcloud/log"
	"github.com/hpcloud/stackato-go/server"
	"logyard/cli/commands"
	"logyard/util/subcommand_server"
)

func main() {
	srv := subcommand_server.Server{
		commands.GetAll()}
	server.MarkRunning("logyard_remote")
	log.Fatal(srv.Start("127.0.0.1:8891"))
}
