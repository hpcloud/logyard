package main

import (
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"os"
	"os/signal"
	"syscall"
)

func Cleanup() {
	log.Info("Cleaning inotify watchers before exiting.")
	tail.Cleanup()
}

func HandleInterrupts() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for sig := range c {
		log.Infof("Captured signal %v; exiting", sig)
		Cleanup()
		os.Exit(1)
	}
}
