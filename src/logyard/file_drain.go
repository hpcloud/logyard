package logyard

import (
	"github.com/ActiveState/log"
	"launchpad.net/tomb"
	"os"
)

// File drain is used to write to local files
type FileDrain struct {
	log *log.Logger
	tomb.Tomb
}

func NewFileDrain(log *log.Logger) Drain {
	rd := &FileDrain{}
	rd.log = log
	return rd
}

func (d *FileDrain) Start(config *DrainConfig) {
	defer d.Done()

	overwrite, err := config.GetParamBool("overwrite", false)
	if err != nil {
		d.Kill(err)
		return
	}

	mode := os.O_WRONLY | os.O_CREATE
	if overwrite {
		mode |= os.O_TRUNC
	} else {
		mode |= os.O_APPEND
	}
	d.log.Infof("Opening %s (overwrite=%v) ...", config.Path, overwrite)
	f, err := os.OpenFile(config.Path, mode, 0600)
	if err != nil {
		d.Kill(err)
		return
	}
	d.log.Infof("Opened %s", config.Path)
	defer f.Close()

	c, err := NewClientGlobal()
	if err != nil {
		d.Kill(err)
		return
	}
	defer c.Close()

	ss, err := c.Recv(config.Filters)
	if err != nil {
		d.Kill(err)
		return
	}
	defer ss.Stop()

	for {
		select {
		case msg := <-ss.Ch:
			data, err := config.FormatJSON(msg.Value)
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
	d.log.Info("Exiting")
}

func (d *FileDrain) Stop() error {
	d.log.Info("Stopping...")
	d.Kill(nil)
	return d.Wait()
}
