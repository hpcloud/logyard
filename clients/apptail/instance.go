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
func (instance *Instance) Tail() {
	log.Infof("Tailing %v logs for %v[%v] -- %+v",
		instance.Type, instance.AppName, instance.Index, instance)

	for name, filename := range instance.LogFiles {
		go instance.tailFile(name, filename)
	}
}

func (instance *Instance) tailFile(name, filename string) {
	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()

	limit, err := instance.getReadLimit(pub, name, filename)
	if err != nil {
		log.Error(err)
		return
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
		instance.publishLine(pub, name, line)
	}

	err = tail.Wait()
	if err != nil {
		log.Error(err)
	}

	log.Infof("Completed tailing %v for %v[%v]", name, instance.AppName, instance.Index)
}

func (instance *Instance) getReadLimit(
	pub *zmqpubsub.Publisher,
	logname string,
	filename string) (int64, error) {
	// convert MB to limit in bytes.
	filesizeLimit := GetConfig().FileSizeLimit * 1024 * 1024
	if !(filesizeLimit > 0) {
		panic("invalid value for `read_limit' in apptail config")
	}

	fi, err := os.Stat(filename)
	if err != nil {
		return -1, fmt.Errorf("Cannot stat file (%s); %s", filename, err)
	}
	size := fi.Size()
	limit := filesizeLimit
	if size > filesizeLimit {
		err := fmt.Errorf("Skipping much of a large log file (%s); size (%v bytes) > read_limit (%v bytes)",
			logname, size, filesizeLimit)
		// Publish special error message.
		instance.publishLine(pub, logname, &tail.Line{
			Text: err.Error(),
			Time: time.Now(),
			Err:  err})
	} else {
		limit = size
	}
	return limit, nil
}

// publishLine zmq-publishes a log line corresponding to this instance
func (instance *Instance) publishLine(
	pub *zmqpubsub.Publisher,
	logname string,
	line *tail.Line) {

	msg := &Message{
		Text:          line.Text,
		LogFilename:   logname,
		UnixTime:      line.Time.Unix(),
		HumanTime:     ToHerokuTime(line.Time),
		Source:        instance.Type,
		InstanceIndex: instance.Index,
		AppGUID:       instance.AppGUID,
		AppName:       instance.AppName,
		AppSpace:      instance.AppSpace,
		NodeID:        LocalNodeId(),
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
