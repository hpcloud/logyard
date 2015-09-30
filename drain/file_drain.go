package drain

import (
	"logyard"
	"os"

	"github.com/hpcloud/log"
	"gopkg.in/tomb.v1"
)

// File drain is used to write to local files
type FileDrain struct {
	name   string
	initCh chan bool
	tomb.Tomb
}

func NewFileDrain(name string) DrainType {
	var d FileDrain
	d.name = name
	d.initCh = make(chan bool)
	return &d
}

func (d *FileDrain) Start(config *DrainConfig) {
	defer d.Done()

	overwrite, err := config.GetParamBool("overwrite", false)
	if err != nil {
		d.Kill(err)
		go d.finishedStarting(false)
		return
	}

	mode := os.O_WRONLY | os.O_CREATE
	if overwrite {
		mode |= os.O_TRUNC
	} else {
		mode |= os.O_APPEND
	}
	log.Infof("[drain:%s] Attempting to open %s (overwrite=%v) ...",
		d.name, config.Path, overwrite)
	f, err := os.OpenFile(config.Path, mode, 0600)
	if err != nil {
		d.Kill(err)
		go d.finishedStarting(false)
		return
	}
	log.Infof("[drain:%s] Successfully opened %s.", d.name, config.Path)
	defer f.Close()

	sub := logyard.Broker.Subscribe(config.Filters...)
	defer sub.Stop()

	go d.finishedStarting(true)

	for {
		select {
		case msg := <-sub.Ch:
			data, err := config.FormatJSON(msg)
			if err != nil {
				d.Kill(err)
				return
			}
			_, err = f.Write(data)
			if err != nil {
				d.Kill(err)
				return
			}
		case <-d.Dying():
			return
		}
	}
}

func (d *FileDrain) finishedStarting(success bool) {
	d.initCh <- success
}

func (d *FileDrain) WaitRunning() bool {
	return <-d.initCh
}

func (d *FileDrain) Stop() error {
	d.Kill(nil)
	return d.Wait()
}
