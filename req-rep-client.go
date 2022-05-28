package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
)

func main() {
	sock, err := goczmq.NewReq("tcp://localhost:5555")
	if err != nil {
		log.Fatalln(err)
	}
	defer sock.Destroy()

	for i := 0; i < 10; i++ {
		fmt.Printf("Sending request %d...\n", i)
		if err := sock.SendFrame([]byte("Hello"), 0); err != nil {
			log.Printf("[ERROR] Failed to send 'Hello' to server: %v\n", err)
			continue
		}

		msg, err := sock.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to recieve message from server: %v\n", err)
			continue
		}
		fmt.Printf("Received reply %d [ %s ]\n", i, msg[0])
	}
}
