package drain

import (
	"launchpad.net/tomb"
	"log"
	"logyard"
	"net"
)

type UdpDrain struct {
	log *log.Logger
	tomb.Tomb
}

func NewUdpDrain(log *log.Logger) Drain {
	rd := &UdpDrain{}
	rd.log = log
	return rd
}

func (d *UdpDrain) Start(config *DrainConfig) {
	defer d.Done()

	d.log.Println("Connecting to UDP addr ", config.Host)
	conn, err := net.Dial("udp", config.Host)
	if err != nil {
		d.Kill(err)
		return
	}
	defer conn.Close()
	d.log.Println("Connected to UDP addr ", config.Host)

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

func (d *UdpDrain) Stop() error {
	d.log.Println("Stopping...")
	d.Kill(nil)
	return d.Wait()
}
