package drain

import (
	"github.com/hpcloud/log"
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

	var conn net.Conn
	dialer := NewNetDialer(config.Scheme, config.Host, 10*time.Second)

	select {
	case conn = <-dialer.Ch:
		if dialer.Error != nil {
			d.Kill(dialer.Error)
			go d.finishedStarting(false)
			return
		}
	case <-d.Dying():
		// Close the connection returned in future by the dialer.
		log.Infof("[drain:%s] Stop request; deferring close of connection",
			d.name)
		go dialer.WaitAndClose()

		// [bug 105165].
		// did not attain running state. No kill however as
		// getting aborted by the user is not really an
		// error. The state machine will treat it as a regular
		// non-error exit.
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
