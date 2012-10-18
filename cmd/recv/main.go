package main

import (
	"fmt"
	"log"
	"logyard"
)

func main() {
	c := logyard.NewClient()
	ss, err := c.Recv("")
	if err != nil {
		log.Fatal(err)
	}
	for msg := range ss.Ch {
		fmt.Println("->", msg.Key, msg.Value)
	}
	err = ss.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
