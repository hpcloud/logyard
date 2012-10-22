package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/srid/tail"
	"log"
	"logyard"
	"net"
	"os"
)

func main() {
	LoadConfig()

	ipaddr, err := localIP()
	if err != nil {
		log.Fatalf("Failed to determine IP addr: %v", err)
	}
	log.Println("Host IP: ", ipaddr)

	c, err := logyard.NewClientGlobal()
	if err != nil {
		log.Fatal(err)
	}
	tailers := []*tail.Tail{}

	for process, logfile := range PROCESSES {
		if logfile == "" {
			logfile = fmt.Sprintf("/s/logs/%s.log", process)
		}
		nodeid := ipaddr.String()

		log.Println("Tailing... ", logfile)
		t, err := tail.TailFile(logfile, tail.Config{
			MaxLineSize: Config.MaxRecordSize,
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
				data, err := json.Marshal(map[string]interface{}{
					"UnixTime": line.UnixTime,
					"Text":     line.Text,
					"Name":     process,
					"NodeID":   nodeid})
				if err != nil {
					tail.Kill(err)
					break
				}
				err = c.Send("systail."+process+"."+nodeid, string(data))
				if err != nil {
					log.Fatal("Failed to send: ", err)
				}
			}
		}(process, t)
	}

	for _, tail := range tailers {
		err := tail.Wait()
		if err != nil {
			log.Printf("error from tail [on %s]: %s", tail.Filename, err)
		}
	}

	// we don't expect any of the tailers to exit with or without
	// error.

	os.Exit(1)
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
