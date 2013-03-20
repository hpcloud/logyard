package drain

import (
	"github.com/ActiveState/log"
	"launchpad.net/tomb"
	"logyard"
	"os"
)

// File drain is used to write to local files
type FileDrain struct {
	name string
	tomb.Tomb
}

func NewFileDrain(name string) Drain {
	rd := &FileDrain{}
	rd.name = name
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
	log.Infof("[drain:%s] Attempting to open %s (overwrite=%v) ...",
		d.name, config.Path, overwrite)
	f, err := os.OpenFile(config.Path, mode, 0600)
	if err != nil {
		d.Kill(err)
		return
	}
	log.Infof("[drain:%s] Successfully opened %s.", d.name, config.Path)
	defer f.Close()

	sub := logyard.Broker.Subscribe(config.Filters...)
	defer sub.Stop()

	for {
		select {
		case msg := <-sub.Ch:
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
}

func (d *FileDrain) Stop() error {
	d.Kill(nil)
	return d.Wait()
}
