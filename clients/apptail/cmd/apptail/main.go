package main

import (
	"github.com/ActiveState/log"
	"github.com/alecthomas/gozmq"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"logyard/clients/apptail"
	"logyard/clients/apptail/docker"
	apptail_event "logyard/clients/apptail/event"
	"logyard/clients/common"
	"os"
	"stackato/server"
)

func main() {
	go common.RegisterTailCleanup()
	major, minor, patch := gozmq.Version()
	log.Infof("Starting apptail (zeromq %d.%d.%d)", major, minor, patch)

	apptail.LoadConfig()
	log.Infof("Config: %+v\n", apptail.GetConfig())

	uid := getUID()

	natsclient := server.NewNatsClient(3)

	natsclient.Subscribe("logyard."+uid+".newinstance", func(instance *apptail.Instance) {
		instance.Tail()
	})

	natsclient.Publish("logyard."+uid+".start", []byte("{}"))
	log.Infof("Waiting for app instances ...")

	go docker.DockerListener.Listen()

	apptail_event.MonitorCloudEvents()
}

// getUID returns the UID of the aggregator running on this node. the UID is
// also shared between the local dea/stager, so that we send/receive messages
// only from the local dea/stagers.
func getUID() string {
	var UID string
	uidFile := "/tmp/logyard.uid"
	if _, err := os.Stat(uidFile); os.IsNotExist(err) {
		uid, err := uuid.NewV4()
		if err != nil {
			common.Fatal("%v", err)
		}
		UID = uid.String()
		if err = ioutil.WriteFile(uidFile, []byte(UID), 0644); err != nil {
			common.Fatal("%v", err)
		}
	} else {
		data, err := ioutil.ReadFile(uidFile)
		if err != nil {
			common.Fatal("%v", err)
		}
		UID = string(data)
	}
	log.Infof("detected logyard UID: %s\n", UID)
	return UID
}
