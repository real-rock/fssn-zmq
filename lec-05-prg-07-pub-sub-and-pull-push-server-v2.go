package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
)

func main() {
	publisher, err := goczmq.NewPub("tcp://*:5557")
	if err != nil {
		log.Fatalln(err)
	}
	defer publisher.Destroy()

	collector, err := goczmq.NewPull("tcp://*:5558")
	if err != nil {
		log.Fatalln(err)
	}
	defer collector.Destroy()

	for {
		msg, err := collector.RecvMessage()
		if err != nil {
			continue
		}
		fmt.Printf("server: publishing update => %s\n", msg[0])
		if err := publisher.SendFrame(msg[0], 0); err != nil {
			log.Printf("[ERROR] Failed to send message to client: %v\n", err)
		}
	}
}
