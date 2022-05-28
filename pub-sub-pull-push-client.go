package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"math/rand"
	"time"
)

func getRandRange(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min) + min
}

func main() {
	subscriber, err := goczmq.NewSub("tcp://localhost:5557", "")
	if err != nil {
		log.Fatalln(err)
	}
	pusher, err := goczmq.NewPush("tcp://localhost:5558")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		if subscriber.Pollin() {
			msg, err := subscriber.RecvMessage()
			if err != nil {
				log.Printf("[ERROR] Failed to get message from server: %v\n", err)
			} else {
				fmt.Printf("I: received message %s\n", msg[0])
			}
		} else {
			randNum := getRandRange(1, 100)
			if randNum < 10 {
				if err := pusher.SendFrame([]byte(fmt.Sprintf("%d", randNum)), 0); err != nil {
					log.Printf("[ERROR] Failed to send message to server: %v\n", err)
				} else {
					fmt.Printf("I: sending message %d\n", randNum)
				}
			}
		}
	}
}
