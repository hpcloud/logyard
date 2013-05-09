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
	name   string
	initCh chan bool
	tomb.Tomb
}

func NewIPConnDrain(name string) DrainType {
	var d IPConnDrain
	d.name = name
	d.initCh = make(chan bool)
	return &d
}

func (d *IPConnDrain) Start(config *DrainConfig) {
	defer d.Done()

	if !(config.Scheme == "udp" || config.Scheme == "tcp") {
		d.Killf("Invalid scheme: %s", config.Scheme)
		go d.finishedStarting(false)
		return
	}

	log.Infof("[drain:%s] Attempting to connect to %s://%s ...",
		d.name, config.Scheme, config.Host)
	conn, err := net.DialTimeout(config.Scheme, config.Host, 10*time.Second)
	if err != nil {
		d.Kill(err)
		go d.finishedStarting(false)
		return
	}
	defer conn.Close()
	log.Infof("[drain:%s] Successfully connected to %s://%s.",
		d.name, config.Scheme, config.Host)

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

func (d *IPConnDrain) finishedStarting(success bool) {
	d.initCh <- success
}

func (d *IPConnDrain) WaitRunning() bool {
	return <-d.initCh
}

func (d *IPConnDrain) Stop() error {
	d.Kill(nil)
	return d.Wait()
}
