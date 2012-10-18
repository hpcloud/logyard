package main

import (
	"bufio"
	"log"
	"logyard"
	"os"
	"strings"
)

func main() {
	c := logyard.NewClient()
	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if line == "" {
			continue
		}
		line = line[:len(line)-1]
		parts := strings.SplitN(line, " ", 2)
		key, value := parts[0], parts[1]
		err = c.Send(key, value)
		if err != nil {
			log.Fatal("Failed to send: ", err)
		}
	}
}
