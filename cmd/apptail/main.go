package main

import (
	"encoding/json"
	"fmt"
	"github.com/apcera/nats"
	"github.com/nu7hatch/gouuid"
	"github.com/srid/tail"
	"io/ioutil"
	"log"
	"logyard"
	"os"
	"path/filepath"
)

// AppInstance is the NATS message sent by dea/stager to notify of new
// instances.
type AppInstance struct {
	AppID    int
	AppName  string
	Type     string
	Index    int
	LogFiles []string
}

// AppInstanceStarted is invoked when dea/stager starts an application
// instance.
func AppInstanceStarted(c *logyard.Client, instance *AppInstance) {
	log.Printf("New instance was started: %v\n", instance)
	key := fmt.Sprintf("apptail.%d", instance.AppID)
	for _, filename := range instance.LogFiles {
		go func(filename string) {
			tail, err := tail.TailFile(filename, tail.Config{
				MaxLineSize: 1500, // TODO logyard.Config.MaxRecordSize,
				MustExist:   true,
				Follow:      true,
				Location:    -1,
				ReOpen:      false,
				Poll:        true})
			if err != nil {
				log.Printf("Error: cannot tail file (%s); %s\n", filename, err)
				return
			}
			for line := range tail.Lines {
				data, err := json.Marshal(map[string]interface{}{
					"Text":          line.Text,
					"LogFilename":   filepath.Base(filename),
					"UnixTime":      line.UnixTime,
					"InstanceIndex": instance.Index,
					"InstanceType":  instance.Type})
				if err != nil {
					log.Fatal(err)
				}
				err = c.Send(key, string(data))
				if err != nil {
					log.Fatal("Failed to send: ", err)
				}
			}
			err = tail.Wait()
			if err != nil {
				log.Println(err)
			}
		}(filename)
	}
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
