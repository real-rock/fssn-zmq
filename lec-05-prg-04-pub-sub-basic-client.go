package main

import (
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	var zipFilter string

	if len(os.Args) > 1 {
		zipFilter = os.Args[1]
	} else {
		zipFilter = "10001"
	}

	sock, err := goczmq.NewSub("tcp://localhost:5556", zipFilter)
	if err != nil {
		log.Fatalln(err)
	}
	defer sock.Destroy()

	totalTemp := 0
	var i int
	for i = 0; i < 20; i++ {
		msg, err := sock.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to read message from server: %v\n", err)
			continue
		}
		res := strings.Split(string(msg[0]), " ")
		if temp, err := strconv.Atoi(res[1]); err != nil {
			log.Printf("[ERROR] Invalid temperature: %v\n", err)
		} else {
			totalTemp += temp
			fmt.Printf("Receive temperature for zipcode '%s' was %d F\n", zipFilter, temp)
		}
	}
	fmt.Printf("Average temperature for zipcode '%s' was '%f' F\n",
		zipFilter, float32(totalTemp)/float32(i+1))
}
