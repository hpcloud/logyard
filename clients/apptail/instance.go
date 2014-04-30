package apptail

import (
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"github.com/ActiveState/tail/ratelimiter"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"logyard/clients/apptail/docker"
	"logyard/clients/apptail/event"
	"logyard/clients/apptail/message"
	"logyard/clients/apptail/pubchannel"
	"logyard/clients/apptail/util"
	"logyard/clients/common"
	"logyard/clients/sieve"
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
	pubch    *pubchannel.PubChannel
}

func (instance *Instance) Identifier() string {
	return fmt.Sprintf("%v[%v:%v]", instance.AppName, instance.Index, instance.DockerId[:docker.ID_LENGTH])
}

// Tail begins tailing the files for this instance.
func (instance *Instance) Tail() {
	log.Infof("Tailing %v logs for %v -- %+v",
		instance.Type, instance.Identifier(), instance)

	stopCh := make(chan bool)

	instance.pubch = pubchannel.New("event.timeline", stopCh)

	logfiles := instance.getLogFiles()

	log.Infof("Determined log files: %+v", logfiles)

	for name, filename := range logfiles {
		go instance.tailFile(name, filename, stopCh)
	}

	go func() {
		docker.DockerListener.BlockUntilContainerStops(instance.DockerId)
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
		instance.SendTimelineEvent("WARN -- %v", err)
		return
	}

	capacity := GetConfig().RateLimit
	rate := time.Duration(int64(time.Second) / int64(capacity))
	rateLimiter := ratelimiter.NewLeakyBucket(1000, rate)

	t, err := tail.TailFile(filename, tail.Config{
		MaxLineSize: GetConfig().MaxRecordSize,
		MustExist:   true,
		Follow:      true,
		Location:    &tail.SeekInfo{-limit, os.SEEK_END},
		ReOpen:      false,
		Poll:        false,
		RateLimiter: rateLimiter})
	if err != nil {
		log.Warnf("Cannot tail file (%s); %s", filename, err)
		instance.SendTimelineEvent("ERROR -- Cannot tail file (%s); %s", name, err)
		return
	}

FORLOOP:
	for {
		select {
		case line, ok := <-t.Lines:
			if !ok {
				err = t.Wait()
				break FORLOOP
			}
			instance.publishLine(pub, name, line)
		case <-stopCh:
			err = t.Stop()
			break FORLOOP
		}
	}

	if err != nil {
		log.Warn(err)
		instance.SendTimelineEvent("WARN -- Error tailing file (%s); %s", name, err)
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
		// Use $STACKATO_LOG_FILES
		logfiles = make(map[string]string)
		if env, err := docker.GetDockerAppEnv(instance.RootPath); err != nil {
			log.Errorf("Failed to read docker image env: %v", err)
		} else {
			if s, ok := env["STACKATO_LOG_FILES"]; ok {
				parts := strings.Split(s, ":")
				if len(parts) > 7 {
					parts = parts[len(parts)-7 : len(parts)]
					instance.SendTimelineEvent("WARN -- $STACKATO_LOG_FILES is large; using only last 7 logs: %v", parts)
				}
				for _, f := range parts {
					parts := strings.SplitN(f, "=", 2)
					logfiles[parts[0]] = parts[1]
				}
			} else {
				log.Errorf("Expected env $STACKATO_LOG_FILES not found in docker image")
			}
		}
	}

	// Expand paths, and securely ensure they fall within the app root.
	logfilesSecure := make(map[string]string)
	for name, path := range logfiles {
		var fullpath string

		// Treat relative paths as being relative to $STACKATO_APP_ROOT
		if !filepath.IsAbs(path) {
			stackatoAppRoot := "/home/stackato/"
			fullpath = filepath.Join(instance.RootPath, stackatoAppRoot, path)
		} else {
			fullpath = filepath.Join(instance.RootPath, path)
		}

		fullpath, err := filepath.Abs(fullpath)
		if err != nil {
			log.Warnf("Cannot find Abs of %v <join> %v: %v", instance.RootPath, path, err)
			instance.SendTimelineEvent("WARN -- Failed to find absolute path for %v", path)
			continue
		}
		fullpath, err = filepath.EvalSymlinks(fullpath)
		if err != nil {
			log.Infof("Error reading log file %v: %v", fullpath, err)
			instance.SendTimelineEvent("WARN -- Ignoring missing/inaccessible path %v", path)
			continue
		}
		if !strings.HasPrefix(fullpath, instance.RootPath) {
			log.Warnf("Ignoring insecure log path %v (via %v) in instance %+v", fullpath, path, instance)
			// This user warning is exactly the same as above, lest we provide
			// a backdoor for a malicious user to list the directory tree on
			// the host.
			instance.SendTimelineEvent("WARN -- Ignoring missing/inaccessible path %v", path)
			continue
		}
		logfilesSecure[name] = fullpath
	}

	if len(logfilesSecure) == 0 {
		instance.SendTimelineEvent("ERROR -- No valid log files detected for tailing")
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
		instance.SendTimelineEvent(
			"Skipping much of a large log file (%s); size (%v bytes) > read_limit (%v bytes)",
			logname, size, filesizeLimit)
	} else {
		limit = size
	}
	return limit, nil
}

// publishLine publishes a log line corresponding to this instance.
func (instance *Instance) publishLine(pub *zmqpubsub.Publisher, logname string, line *tail.Line) {
	instance.publishLineAs(pub, instance.Type, logname, line)
}

func (instance *Instance) publishLineAs(pub *zmqpubsub.Publisher, source string, logname string, line *tail.Line) {
	if line == nil {
		panic("line is nil")
	}

	msg := &message.Message{
		LogFilename:   logname,
		Source:        source,
		InstanceIndex: instance.Index,
		AppGUID:       instance.AppGUID,
		AppName:       instance.AppName,
		AppSpace:      instance.AppSpace,
		MessageCommon: common.NewMessageCommon(line.Text, line.Time, util.LocalNodeId()),
	}

	if line.Err != nil {
		// Mark this as a special error record, as it is
		// coming from tail, not the app.
		msg.Source = "stackato[apptail]"
		msg.LogFilename = ""
		log.Warnf("[%s] %s", instance.AppName, line.Text)
	}

	err := msg.Publish(pub, false)
	if err != nil {
		common.Fatal("Unable to publish: %v", err)
	}
}

func (instance *Instance) SendTimelineEvent(format string, v ...interface{}) {
	line := fmt.Sprintf(format, v...)
	tEvent := event.TimelineEvent{event.App{instance.AppGUID, instance.AppSpace, instance.AppName}, instance.Index}
	evt := sieve.Event{
		Type:     "timeline",
		Desc:     line,
		Severity: "INFO",
		Info: map[string]interface{}{
			"app":            tEvent.App,
			"instance_index": tEvent.InstanceIndex,
		},
		Process:       "apptail",
		MessageCommon: common.NewMessageCommon(line, time.Now(), util.LocalNodeId()),
	}
	instance.pubch.Ch <- evt
}
