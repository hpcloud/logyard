package apptail

import (
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"os"
	"time"
)

// Instance is the NATS message sent by dea_ng to notify of new instances.
type Instance struct {
	AppGUID  string
	AppName  string
	AppSpace string `json:"space"`
	Type     string
	Index    int
	DockerId string `json:"docker_id"`
	LogFiles map[string]string
}

// Tail begins tailing the files for this instance.
func (instance *Instance) Tail(nodeid string) {
	log.Infof("Tailing %v logs for %v[%v] -- %+v",
		instance.Type, instance.AppName, instance.Index, instance)

	// convert MB to limit in bytes.
	filesize_limit := GetConfig().FileSizeLimit * 1024 * 1024

	if !(filesize_limit > 0) {
		panic("invalid value for `read_limit' in apptail config")
	}

	for name, filename := range instance.LogFiles {
		go func(name string, filename string) {
			pub := logyard.Broker.NewPublisherMust()
			defer pub.Stop()

			fi, err := os.Stat(filename)
			if err != nil {
				log.Errorf("Cannot stat file (%s); %s", filename, err)
				return
			}
			size := fi.Size()
			limit := filesize_limit
			if size > filesize_limit {
				err := fmt.Errorf("Skipping much of a large log file (%s); size (%v bytes) > read_limit (%v bytes)",
					name, size, filesize_limit)
				// Publish special error message.
				instance.publishLine(nodeid, name, pub, &tail.Line{
					Text: err.Error(),
					Time: time.Now(),
					Err:  err})
			} else {
				limit = size
			}

			tail, err := tail.TailFile(filename, tail.Config{
				MaxLineSize: GetConfig().MaxRecordSize,
				MustExist:   true,
				Follow:      true,
				Location:    &tail.SeekInfo{-limit, os.SEEK_END},
				ReOpen:      false,
				Poll:        false,
				LimitRate:   GetConfig().RateLimit})
			if err != nil {
				log.Errorf("Cannot tail file (%s); %s", filename, err)
				return
			}

			for line := range tail.Lines {
				instance.publishLine(nodeid, name, pub, line)
			}

			err = tail.Wait()
			if err != nil {
				log.Error(err)
			}

			log.Infof("Completed tailing %v for %v[%v]", name, instance.AppName, instance.Index)
		}(name, filename)
	}
}

func (instance *Instance) publishLine(
	nodeid string,
	name string, pub *zmqpubsub.Publisher,
	line *tail.Line) {

	msg := &Message{
		Text:          line.Text,
		LogFilename:   name,
		UnixTime:      line.Time.Unix(),
		HumanTime:     ToHerokuTime(line.Time),
		Source:        instance.Type,
		InstanceIndex: instance.Index,
		AppGUID:       instance.AppGUID,
		AppName:       instance.AppName,
		AppSpace:      instance.AppSpace,
		NodeID:        nodeid,
	}

	if line.Err != nil {
		// Mark this as a special error record, as it is
		// coming from tail, not the app.
		msg.Source = "stackato.apptail"
		msg.LogFilename = ""
		log.Warnf("[%s] %s", instance.AppName, line.Text)
	}

	err := msg.Publish(pub, false)
	if err != nil {
		log.Fatal(err)
	}

}
