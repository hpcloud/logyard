package main

import (
	"github.com/apcera/nats"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"logyard"
	"logyard/log2"
	"os"
)

func main() {
	LoadConfig()

	uid := getUID()

	logyardclient, err := logyard.NewClientGlobal()
	if err != nil {
		log2.Fatal(err)
	}
	natsclient := newNatsClient()

	natsclient.Subscribe("logyard."+uid+".newinstance", func(instance *AppInstance) {
		AppInstanceStarted(logyardclient, instance)
	})

	natsclient.Publish("logyard."+uid+".start", []byte("{}"))
	log2.Infof("Waiting for instances...")

	MonitorCloudEvents()
}

func newNatsClient() *nats.EncodedConn {
	log2.Infof("Connecting to NATS %s\n", Config.NatsUri)
	nc, err := nats.Connect(Config.NatsUri)
	if err != nil {
		log2.Fatal(err)
	}
	log2.Infof("Connected to NATS %s\n", Config.NatsUri)
	client, err := nats.NewEncodedConn(nc, "json")
	if err != nil {
		log2.Fatal(err)
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
			log2.Fatal(err)
		}
		UID = uid.String()
		if err = ioutil.WriteFile(uidFile, []byte(UID), 0644); err != nil {
			log2.Fatal(err)
		}
	} else {
		data, err := ioutil.ReadFile(uidFile)
		if err != nil {
			log2.Fatal(err)
		}
		UID = string(data)
	}
	log2.Infof("detected logyard UID: %s\n", UID)
	return UID
}
