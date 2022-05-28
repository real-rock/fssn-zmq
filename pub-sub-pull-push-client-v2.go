package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"math/rand"
	"os"
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
	defer subscriber.Destroy()
	pusher, err := goczmq.NewPush("tcp://localhost:5558")
	if err != nil {
		log.Fatalln(err)
	}
	defer pusher.Destroy()

	var clientID string
	if len(os.Args) < 1 {
		clientID = "Default ID"
	} else {
		clientID = os.Args[1]
	}
	for {
		if subscriber.Pollin() {
			msg, err := subscriber.RecvMessage()
			if err != nil {
				log.Printf("[ERROR] Failed to get message from server: %v\n", err)
			}
			fmt.Printf("%s: receive status => %s\n", clientID, msg[0])
		} else {
			randomNum := getRandRange(1, 100)
			if randomNum < 10 {
				time.Sleep(time.Second)
				msg := "(" + clientID + ":ON)"
				if err := pusher.SendFrame([]byte(msg), 0); err != nil {
					log.Printf("[ERROR] Failed to send message to server: %v\n", err)
				} else {
					fmt.Printf("%s: send status - activated\n", clientID)
				}
			} else if randomNum > 90 {
				time.Sleep(time.Second)
				msg := "(" + clientID + ":OFF)"
				if err := pusher.SendFrame([]byte(msg), 0); err != nil {
					log.Printf("[ERROR] Failed to send message to server: %v\n", err)
				} else {
					fmt.Printf("%s: send status - deactivated\n", clientID)
				}
			}
		}
	}

}
