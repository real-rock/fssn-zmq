package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"os"
	"strconv"
	"sync"
)

func genWorker(wg *sync.WaitGroup, id int) {
	defer wg.Done()
	worker, err := goczmq.NewDealer("inproc://backend")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Worker#%d started\n", id)
	for {
		msg, err := worker.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to recieve message: %v\n", err)
			continue
		}
		fmt.Printf("Worker#%d received '%s' from 'client#%s'\n", id, string(msg[1]), msg[2])
		for {
			if err := worker.SendMessage(msg); err == nil {
				break
			}
		}
	}
}

func main() {
	var numServer int
	var wg sync.WaitGroup

	if len(os.Args) < 2 {
		numServer = 5
	} else {
		numServer, _ = strconv.Atoi(os.Args[1])
	}

	proxy := goczmq.NewProxy()
	if err := proxy.SetFrontend(goczmq.Router, "tcp://127.0.0.1:5570"); err != nil {
		log.Fatalln(err)
	}
	if err := proxy.SetBackend(goczmq.Dealer, "inproc://backend"); err != nil {
		log.Fatalln(err)
	}
	defer proxy.Destroy()

	for i := 0; i < numServer; i++ {
		go genWorker(&wg, i)
	}
	wg.Add(numServer)
	wg.Wait()
}
