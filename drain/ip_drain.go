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

func NewIPConnDrain(name string) Drain {
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

	var conn net.Conn
	dialer := NewNetDialer(config.Scheme, config.Host, 10*time.Second)

	select {
	case conn = <-dialer.Ch:
		if dialer.Error != nil {
			d.Kill(dialer.Error)
			return
		}
	case <-d.Dying():
		// Close the connection returned in future by the dialer.
		log.Infof("[drain:%s] Stop request; deferring close of connection",
			d.name)
		go func() {
			conn = <-dialer.Ch
			if dialer.Error == nil {
				conn.Close()
			}
		}()
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
