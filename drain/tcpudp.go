package drain

import (
	"launchpad.net/tomb"
	"log"
	"logyard"
	"net"
)

// IPConnDrain is a drain based on net.IPConn
type IPConnDrain struct {
	log *log.Logger
	tomb.Tomb
}

func NewIPConnDrain(log *log.Logger) Drain {
	rd := &IPConnDrain{}
	rd.log = log
	return rd
}

func (d *IPConnDrain) Start(config *DrainConfig) {
	defer d.Done()

	if !(config.Scheme == "udp" || config.Scheme == "tcp") {
		d.Killf("invalid scheme: %s", config.Scheme)
		return
	}

	d.log.Printf("Connecting to %s addr %s\n", config.Scheme, config.Host)
	conn, err := net.Dial(config.Scheme, config.Host)
	if err != nil {
		d.Kill(err)
		return
	}
	defer conn.Close()
	d.log.Printf("Connected to %s addr %s\n", config.Scheme, config.Host)

	c, err := logyard.NewClientGlobal()
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

	go func() {
		for {
			select {
			case msg := <-ss.Ch:
				println(msg.Key, msg.Value)
				data, err := config.FormatJSON(msg.Value)
				if err != nil {
					ss.Stop()
					d.Kill(err)
					return
				}
				_, err = conn.Write(data)
				if err != nil {
					ss.Stop()
					d.Kill(err)
					return
				}
			case <-d.Dying():
				d.log.Println("Dying and stopping stream...")
				err = ss.Stop()
				if err != nil {
					d.log.Printf("Error stopping subscribe stream: %s\n", err)
				}
				return
			}
		}
	}()

	d.Kill(ss.Wait())
}

func (d *IPConnDrain) Stop() error {
	d.log.Println("Stopping...")
	d.Kill(nil)
	return d.Wait()
}
