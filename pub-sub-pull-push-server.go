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
	pull, err := goczmq.NewPull("tcp://*:5558")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		msg, err := pull.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to get message from client: %v\n", err)
		} else {
			fmt.Printf("I: publishing update %s\n", msg[0])
			if err := publisher.SendMessage(msg); err != nil {
				log.Printf("[ERROR] Failed to send message to client: %v\n", err)
			}
		}
	}
}
