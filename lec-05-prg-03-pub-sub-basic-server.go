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
	fmt.Println("Publishing updates at weather server...")
	sock, err := goczmq.NewPub("tcp://*:5556")
	if err != nil {
		log.Fatalln(err)
	}
	defer sock.Destroy()

	for {
		zipcode := getRandRange(1, 100000)
		temperature := getRandRange(-80, 135)
		relhumidity := getRandRange(10, 60)

		msg := fmt.Sprintf("%d %d %d", zipcode, temperature, relhumidity)
		for {
			if err := sock.SendFrame([]byte(msg), 0); err == nil {
				break
			}
		}
	}
}
