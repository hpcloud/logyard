package main

import (
	"flag"
	"fmt"
	"log"
	"logyard"
)

type options struct {
	hideprefix *bool
	filter     *string
}

var Options = options{
	flag.Bool("hideprefix", false, "hide message prefix"),
	flag.String("filter", "", "filter by message key pattern")}

func main() {
	flag.Parse()

	c := logyard.NewClient()
	ss, err := c.Recv(*Options.filter)
	if err != nil {
		log.Fatal(err)
	}
	for msg := range ss.Ch {
		if *Options.hideprefix {
			fmt.Println(msg.Value)
		} else {
			fmt.Println(msg.Key, msg.Value)
		}
	}
	err = ss.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
