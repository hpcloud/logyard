package main

import (
	"encoding/json"
	"fmt"
	"github.com/srid/tail"
	"log"
	"logyard"
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
				MaxLineSize: Config.MaxRecordSize,
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
