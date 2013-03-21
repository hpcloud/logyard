package main

import (
	"fmt"
	"logyard/zeroutine"
	"math/rand"
	"time"
)

var Broker zeroutine.Zeroutine

func init() {
	Broker.PubAddr = "tcp://127.0.0.1:4000"
	Broker.SubAddr = "tcp://127.0.0.1:4001"
	Broker.BufferSize = 100
}

func RunPublisher(name string) {
	pub, err := Broker.NewPublisher()
	if err != nil {
		panic(err)
	}
	defer pub.Stop()
	
	count := 1
	for {
		err := pub.Publish(name, fmt.Sprintf("%d", count))
		if err != nil {
			panic(err)
		}
		count += 1
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

func EchoSubscribe() {
	sub := Broker.Subscribe("")
	fmt.Println("Monitoring subscription..")
	for msg := range sub.Ch {
		fmt.Printf("%s => %s\n", msg.Key, msg.Value)
	}
	fmt.Println("End of subscription")
}

func main() {
	// Subscribe
	fmt.Println("Starting subscriber")
	go EchoSubscribe()

	// Setup sample publishers
	fmt.Println("Setting up publishers")
	go RunPublisher("srid")
	go RunPublisher("suraj")
	go RunPublisher("kalai")
	go RunPublisher("jill")

	// Run the broker
	fmt.Printf("Running the broker: %+v\n", Broker)
	panic(Broker.Run())
}
