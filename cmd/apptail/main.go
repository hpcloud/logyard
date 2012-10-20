package main

import (
	"github.com/apcera/nats"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"log"
	"logyard"
	"os"
)

func main() {
	uid := getUID()

	logyardclient := logyard.NewClient()
	natsclient := newNatsClient()

	natsclient.Subscribe("logyard."+uid+".newinstance", func(instance *AppInstance) {
		AppInstanceStarted(logyardclient, instance)
	})

	natsclient.Publish("logyard."+uid+".start", []byte("{}"))

	log.Printf("Waiting for instances...")
	<-make(chan int) // block forever
}

func newNatsClient() *nats.EncodedConn {
	// TODO: use doozer
	natsUri := "nats://127.0.0.1:4222/"
	log.Printf("Connecting to NATS %s \n", natsUri)
	nc, err := nats.Connect(natsUri)
	if err != nil {
		log.Fatal(err)
	}
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
			panic(err)
		}
		UID = uid.String()
		if err = ioutil.WriteFile(uidFile, []byte(UID), 0644); err != nil {
			panic(err)
		}
	} else {
		data, err := ioutil.ReadFile(uidFile)
		if err != nil {
			panic(err)
		}
		UID = string(data)
	}
	log.Printf("detected logyard UID: %s\n", UID)
	return UID
}
