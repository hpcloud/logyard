package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard/util/lineserver"
)

func main() {
	var useUDP bool
	var port int
	flag.BoolVar(&useUDP, "u", false, "use UDP instead of TCP")
	flag.IntVar(&port, "p", 9090, "port number to bind to")
	flag.Parse()

	var srv *lineserver.LineServer
	var err error
	addr := fmt.Sprintf(":%d", port)
	proto := "tcp"
	if useUDP {
		proto = "udp"
	}

	if useUDP {
		srv, err = lineserver.NewLineServerUDP(addr)
	} else {
		srv, err = lineserver.NewLineServerTCP(addr)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Server running as %s://%s", proto, addr)
	go srv.Start()

	for line := range srv.Ch {
		fmt.Println(line)
	}
}
