package main

import (
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"os"
)

func cleanup() {
	log.Info("cleanup: closing open inotify watchers")
	tail.Cleanup()
}

func fatal(format string, v ...interface{}) {
	log.Fatal0(format, v...)
	cleanup()
	os.Exit(1)
}
