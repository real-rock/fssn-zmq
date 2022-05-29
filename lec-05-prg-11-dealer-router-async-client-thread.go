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

func recvHandler(poller *goczmq.Poller, socket *goczmq.Sock, identity string) {
	for {
		if poller.Wait(1000) == socket {
			msg, err := socket.RecvMessage()
			if err != nil {
				log.Printf("[ERROR] Failed to recieve message: %v\n", err)
			} else {
				fmt.Printf("%s received: %s\n", identity, msg[0])
			}
		}
	}
}

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
		defer poll.Destroy()

		go recvHandler(poll, sock, identity)

		reqs := 0
		for {
			reqs += 1
			msgStr := fmt.Sprintf("request #%d", reqs)
			for {
				if err := sock.SendMessage([][]byte{[]byte(msgStr), []byte(sock.Identity())}); err != nil {
					log.Printf("[ERROR] Failed to send message: %v\n", err)
					continue
				} else {
					fmt.Printf("Req #%d sent..\n", reqs)
					break
				}
			}
			time.Sleep(time.Second)
		}
	}(id)
	wg.Wait()
}
