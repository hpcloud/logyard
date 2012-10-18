package main

import (
	"errors"
	"fmt"
	"github.com/srid/tail"
	"log"
	"logyard"
	"net"
)

func main() {
	ipaddr, err := localIP()
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to determine local ip addr: %v", err))
	}
	log.Println("Host IP: ", ipaddr)

	c := logyard.NewClient()
	tailers := []*tail.Tail{}

	for _, process := range PROCESSES {
		logfile := fmt.Sprintf("/s/logs/%s.log", process)
		nodeid := ipaddr.String()

		t, err := tail.TailFile(logfile, tail.Config{
			MaxLineSize: 1500, // TODO: read from config
			MustExist:   false,
			Follow:      true,
			// ignore existing content, to support subsequent re-runs of systail
			Location: 0,
			ReOpen:   true,
			Poll:     false})
		if err != nil {
			panic(err)
		}

		tailers = append(tailers, t)

		go func(process string, tail *tail.Tail) {
			for line := range tail.Lines {
				/* TODO json --
				record := SystemProcessLogRecord{line.Text, line.UnixTime, c.Name, c.NodeID}
				data, err := json.Marshal(record)
				if err != nil {
					log.Fatal(err)
				}
				passage.Ch <- data */

				err := c.Send("systail."+nodeid+"."+process, line.Text)
				if err != nil {
					log.Fatal("Failed to send: ", err)
				}
			}
		}(process, t)
	}

	for _, tail := range tailers {
		err := tail.Wait()
		if err != nil {
			log.Println("error from tail [on %s]: %s", tail.Filename, err)
		}
	}
}

func localIP() (net.IP, error) {
	tt, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 { // loopback address 
				continue
			}
			return v4, nil
		}
	}
	return nil, errors.New("cannot find local IP address")
}
