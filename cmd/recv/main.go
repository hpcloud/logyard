package main

import (
	"fmt"
	"log"
	"logyard"
	"os"
)

func getPrefix() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return ""
}

func main() {
	c := logyard.NewClient()
	ss, err := c.Recv(getPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for msg := range ss.Ch {
		fmt.Println(msg.Key, msg.Value)
	}
	err = ss.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
