package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	var id int
	if len(os.Args) < 2 {
		id = 0
	} else {
		id, _ = strconv.Atoi(os.Args[1])
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(id int) {
		defer wg.Done()
		sock, err := goczmq.NewDealer("tcp://localhost:5570")
		if err != nil {
			log.Fatalln(err)
		}
		defer sock.Destroy()

		identity := fmt.Sprintf("%d", id)
		sock.SetIdentity(identity)
		fmt.Printf("Client %s started\n", identity)
		poll, err := goczmq.NewPoller(sock)
		if err != nil {
			log.Fatalln(err)
		}
		reqs := 0
		for {
			reqs += 1
			fmt.Printf("Req #%d sent..\n", reqs)
			msgStr := fmt.Sprintf("request #%d", reqs)
			for {
				if err := sock.SendMessage([][]byte{[]byte(msgStr), []byte(sock.Identity())}); err != nil {
					//log.Printf("[ERROR] Failed to send message: %v\n", err)
					continue
				} else {
					break
				}
			}
			time.Sleep(time.Second)
			if poll.Wait(1000) == sock {
				msg, err := sock.RecvMessage()
				if err != nil {
					continue
					//log.Printf("[ERROR] Failed to receive message: %v\n", err)
				} else {
					fmt.Printf("%s received: %s\n", identity, msg[0])
				}
			}
		}
	}(id)
	wg.Wait()
}
