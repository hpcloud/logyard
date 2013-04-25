package drain

import (
	"github.com/ActiveState/log"
	"launchpad.net/tomb"
	"logyard"
	"net"
	"time"
)

// IPConnDrain is a drain based on net.IPConn
type IPConnDrain struct {
	name string
	tomb.Tomb
}

func NewIPConnDrain(name string) DrainType {
	rd := &IPConnDrain{name, tomb.Tomb{}}
	return rd
}

func (d *IPConnDrain) Start(config *DrainConfig) {
	defer d.Done()

	if !(config.Scheme == "udp" || config.Scheme == "tcp") {
		d.Killf("Invalid scheme: %s", config.Scheme)
		return
	}

	log.Infof("[drain:%s] Attempting to connect to %s://%s ...",
		d.name, config.Scheme, config.Host)
	conn, err := net.DialTimeout(config.Scheme, config.Host, 10*time.Second)
	if err != nil {
		d.Kill(err)
		return
	}
	defer conn.Close()
	log.Infof("[drain:%s] Successfully connected to %s://%s.",
		d.name, config.Scheme, config.Host)

	sub := logyard.Broker.Subscribe(config.Filters...)
	defer sub.Stop()

	for {
		select {
		case msg := <-sub.Ch:
			data, err := config.FormatJSON(msg)
			if err != nil {
				d.Kill(err)
				return
			}
			_, err = conn.Write(data)
			if err != nil {
				d.Kill(err)
				return
			}
		case <-d.Dying():
			return
		}
	}
}

func (d *IPConnDrain) Stop() error {
	d.Kill(nil)
	return d.Wait()
}
