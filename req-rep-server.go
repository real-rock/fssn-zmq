package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"time"
)

func main() {
	sock, err := goczmq.NewRep("tcp://*:5555")
	if err != nil {
		log.Fatalln(err)
	}
	defer sock.Destroy()

	for {
		data, err := sock.RecvMessage()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Printf("Received request: %s\n", data[0])
		time.Sleep(time.Second * 1)

		if err := sock.SendFrame([]byte("World"), 0); err != nil {
			log.Printf("[ERROR] Failed to sent 'World' to client: %v", err)
			continue
		}
	}
}
