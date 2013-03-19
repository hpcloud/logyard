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
	log.Infof("[%s] Opening %s (overwrite=%v) ...", d.name, config.Path, overwrite)
	f, err := os.OpenFile(config.Path, mode, 0600)
	if err != nil {
		d.Kill(err)
		return
	}
	log.Infof("[%s] Opened %s", d.name, config.Path)
	defer f.Close()

	c, err := logyard.NewClientGlobal(false)
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
}

func (d *FileDrain) Stop() error {
	d.Kill(nil)
	return d.Wait()
}
