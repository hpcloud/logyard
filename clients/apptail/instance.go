package apptail

import (
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"logyard/clients/messagecommon"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Instance is the NATS message sent by dea_ng to notify of new instances.
type Instance struct {
	AppGUID  string
	AppName  string
	AppSpace string
	Type     string
	Index    int
	DockerId string `json:"docker_id"`
	RootPath string
	LogFiles map[string]string
}

func (instance *Instance) Identifier() string {
	return fmt.Sprintf("%v[%v:%v]", instance.AppName, instance.Index, instance.DockerId[:ID_LENGTH])
}

// Tail begins tailing the files for this instance.
func (instance *Instance) Tail() {
	log.Infof("Tailing %v logs for %v -- %+v",
		instance.Type, instance.Identifier(), instance)

	stopCh := make(chan bool)
	logfiles := instance.getLogFiles()

	log.Infof("Determined log files: %+v", logfiles)

	for name, filename := range logfiles {
		go instance.tailFile(name, filename, stopCh)
	}

	go func() {
		DockerListener.WaitForContainer(instance.DockerId)
		log.Infof("Container for %v exited", instance.Identifier())
		close(stopCh)
	}()
}

func (instance *Instance) tailFile(name, filename string, stopCh chan bool) {
	var err error

	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()

	limit, err := instance.getReadLimit(pub, name, filename)
	if err != nil {
		log.Warn(err)
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
		log.Warnf("Cannot tail file (%s); %s", filename, err)
		return
	}

FORLOOP:
	for {
		select {
		case line, ok := <-tail.Lines:
			if !ok {
				err = tail.Wait()
				break FORLOOP
			}
			instance.publishLine(pub, name, line)
		case <-stopCh:
			err = tail.Stop()
			break FORLOOP
		}
	}

	if err != nil {
		log.Warn(err)
	}

	log.Infof("Completed tailing %v log for %v", name, instance.Identifier())
}

func (instance *Instance) getLogFiles() map[string]string {
	var logfiles map[string]string

	rawMode := len(instance.LogFiles) > 0
	if rawMode {
		// If the logfiles list was explicitly passed, use it as is.
		logfiles = instance.LogFiles
	} else {
		// Else, use the default list configured in apptail config.
		logfiles = make(map[string]string)
		for key, value := range GetConfig().DefaultLogFiles {
			logfiles[key] = value
		}

		// Lookup app-specific log files
		stackatoYmlPath := filepath.Join(instance.RootPath, "/app/app/stackato.yml")
		if _, err := os.Stat(stackatoYmlPath); err == nil {
			if stackatoYml, err := NewStackatoYml(stackatoYmlPath); err != nil {
				log.Warnf("Unable to access/parse stackato.yml for %v: %v", instance.Identifier(), err)
			} else {
				log.Infof("stackato.yml found with %d logfiles entries for %v", len(stackatoYml.LogFiles), instance.Identifier())
				if len(stackatoYml.LogFiles) > 0 {
					log.Infof("Adding app-specific log files: %+v", stackatoYml.LogFiles)
					for key, value := range stackatoYml.LogFiles {
						logfiles[key] = value
					}
				}
			}
		}
	}

	// Expand paths, and securely ensure they fail within the app root.
	logfilesSecure := make(map[string]string)
	for name, path := range logfiles {
		fullpath := filepath.Join(instance.RootPath, path)
		fullpath, err := filepath.Abs(fullpath)
		if err != nil {
			// TODO: push warnings in this function to the app log stream.
			log.Warnf("Cannot find Abs of %v <join> %v: %v", instance.RootPath, path, err)
			continue
		}
		fullpath, err = filepath.EvalSymlinks(fullpath)
		if err != nil {
			log.Warnf("Cannot eval symlinks in path %v <join> %v: %v", instance.RootPath, path, err)
			continue
		}
		if !strings.HasPrefix(fullpath, instance.RootPath) {
			log.Warnf("Ignoring insecure log path %v (via %v) in instance %+v", fullpath, path, instance)
			continue
		}
		logfilesSecure[name] = fullpath
	}

	return logfilesSecure
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

	if line == nil {
		panic("line is nil")
	}

	msg := &Message{
		LogFilename:   logname,
		Source:        instance.Type,
		InstanceIndex: instance.Index,
		AppGUID:       instance.AppGUID,
		AppName:       instance.AppName,
		AppSpace:      instance.AppSpace,
		MessageCommon: messagecommon.New(line.Text, line.Time, LocalNodeId()),
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
