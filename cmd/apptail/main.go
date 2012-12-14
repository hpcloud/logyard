package main

import (
	"github.com/ActiveState/log"
	"github.com/apcera/nats"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"logyard"
	"logyard/stackato/apptail"
	"os"
)

func main() {
	apptail.LoadConfig()

	uid := getUID()

	logyardclient, err := logyard.NewClientGlobal()
	if err != nil {
		log.Fatal(err)
	}
	natsclient := newNatsClient()

	natsclient.Subscribe("logyard."+uid+".newinstance", func(instance *apptail.AppInstance) {
		apptail.AppInstanceStarted(logyardclient, instance)
	})

	natsclient.Publish("logyard."+uid+".start", []byte("{}"))
	log.Infof("Waiting for instances...")

	apptail.MonitorCloudEvents()
}

func newNatsClient() *nats.EncodedConn {
	log.Infof("Connecting to NATS %s\n", apptail.Config.NatsUri)
	nc, err := nats.Connect(apptail.Config.NatsUri)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Connected to NATS %s\n", apptail.Config.NatsUri)
	client, err := nats.NewEncodedConn(nc, "json")
	if err != nil {
		log.Fatal(err)
	}
	return client
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
			log.Fatal(err)
		}
		UID = uid.String()
		if err = ioutil.WriteFile(uidFile, []byte(UID), 0644); err != nil {
			log.Fatal(err)
		}
	} else {
		data, err := ioutil.ReadFile(uidFile)
		if err != nil {
			log.Fatal(err)
		}
		UID = string(data)
	}
	log.Infof("detected logyard UID: %s\n", UID)
	return UID
}
