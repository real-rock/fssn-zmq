package main

import (
	"errors"
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var NameServerNotFoundError = errors.New("no nameserver has been found")

func getRandRange(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min) + min
}

func searchNameserver(ipMask, localIpAddr, portNameserver string) (string, error) {
	sub := goczmq.NewSock(goczmq.Sub)
	defer sub.Destroy()

	for i := 0; i < 255; i++ {
		targetIpAddr := fmt.Sprintf("tcp://%s.%d:%s", ipMask, i, portNameserver)
		if targetIpAddr != localIpAddr {
			if err := sub.Connect(targetIpAddr); err != nil {
				log.Printf("[ERROR] Failed to connect with '%s': %v\n", targetIpAddr, err)
			}
			sub.SetRcvtimeo(2000)
			sub.SetSubscribe("NAMESERVER")
		}
	}
	msg, err := sub.RecvMessage()
	if err != nil {
		log.Printf("[ERROR] Failed to receive message: %v\n", err)
		return "", NameServerNotFoundError
	} else {
		res := strings.Split(string(msg[0]), ":")
		fmt.Printf("%s\n", res)
		if len(res) != 0 && res[0] == "NAMESERVER" {
			return res[1], nil
		}
	}
	return "", NameServerNotFoundError
}

func beaconNameserver(localIpAddr, portNameserver string) error {
	keyboardSig := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(keyboardSig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-keyboardSig
		if sig == os.Interrupt {
			return
		}
		done <- true
	}()

	addr := fmt.Sprintf("tcp://%s:%s", localIpAddr, portNameserver)
	pub, err := goczmq.NewPub(addr)
	if err != nil {
		log.Printf("[ERROR] Failed to connect with '%s': %v", addr, err)
		return err
	}
	defer pub.Destroy()
	fmt.Printf("local p2p name server bind to tcp://%s:%s.\n", localIpAddr, portNameserver)
	for {
		time.Sleep(time.Second)
		msg := fmt.Sprintf("NAMESERVER:%s", localIpAddr)
		if err := pub.SendFrame([]byte(msg), 0); err != nil {
			log.Printf("[ERROR] Failed to send message: %v\n", err)
		}
		if <-done {
			break
		}
	}
	return nil
}

func userMangerNameserver(localIpAddr, portSubscribe string) error {
	var userDb [][]string

	addr := fmt.Sprintf("tcp://%s:%s", localIpAddr, portSubscribe)
	req, err := goczmq.NewReq(addr)
	if err != nil {
		log.Printf("[ERROR] Failed to connect with '%s': %v", addr, err)
		return err
	}
	defer req.Destroy()

	fmt.Printf("local p2p db server activated at tcp://%s:%s.\n", localIpAddr, portSubscribe)
	for {
		userReq, err := req.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to receive request: %v\n", err)
			break
		}
		user := strings.Split(string(userReq[0]), ":")
		userDb = append(userDb, user)
		fmt.Printf("user registration '%s' from '%s'.\n", userReq[1], userReq[0])
		if err := req.SendFrame([]byte("ok"), 0); err != nil {
			log.Printf("[ERROR] Failed to send 'ok': %v", err)
			break
		}
	}
	return nil
}

func relayServerNameserver(localIpAddr, portChatPublisher, portChatCollector string) error {
	pubAddr := fmt.Sprintf("tcp://%s:%s", localIpAddr, portChatPublisher)
	pullAddr := fmt.Sprintf("tcp://%s:%s", localIpAddr, portChatCollector)
	pub, err := goczmq.NewPub(pubAddr)
	if err != nil {
		log.Printf("[ERROR] Failed to connect with '%s': %v", pubAddr, err)
		return err
	}
	defer pub.Destroy()

	collector, err := goczmq.NewPull(pullAddr)
	if err != nil {
		log.Printf("[ERROR] Failed to connect with '%s': %v", pullAddr, err)
		return err
	}
	defer collector.Destroy()

	fmt.Printf("local p2p relay server actiated at tcp://%s:%s & %s", localIpAddr, portChatPublisher, portChatCollector)
	for {
		msg, err := collector.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to receive message: %v\n", err)
			break
		}
		fmt.Printf("p2p-relay:<==>%s", msg[0])
		if err := pub.SendFrame([]byte(fmt.Sprintf("RELAY:%s", msg[0])), 0); err != nil {
			log.Printf("[ERROR] Failed to send message: %v\n", err)
		}
	}
	return nil
}

func getLocalIpAddr() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer func() {
		if err := conn.Close(); err != nil {
			return
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage is 'go run p2p-dechat.go _user-name_'.")
		return
	} else {
		fmt.Println("starting p2p chatting program")
	}
	ipAddrP2pServer := ""
	portNameserver := "9001"
	portChatPublisher := "9002"
	portChatCollector := "9003"
	portSubscribe := "9004"

	userName := os.Args[1]
	ipAddr := getLocalIpAddr()
	getIpMask := func() string {
		addr := strings.Split(ipAddr, ".")
		return addr[len(addr)-1]
	}
	ipMask := getIpMask()
	fmt.Println("searching for p2p server.")
	nameServerIpAddr, err := searchNameserver(ipMask, ipAddr, portNameserver)
	if err == nil {
		ipAddrP2pServer = nameServerIpAddr
		fmt.Printf("p2p server found at %s, and p2p client mode is activated", ipAddrP2pServer)
		fmt.Println("starting user registration procedure.")

		clientSock, err := goczmq.NewReq(fmt.Sprintf("tcp://%s:%s", ipAddrP2pServer, portSubscribe))
		if err != nil {
			log.Fatalln(err)
		}
		defer clientSock.Destroy()
		msg := fmt.Sprintf("%s:%s", ipAddr, userName)
		for {
			if err := clientSock.SendFrame([]byte(msg), 0); err == nil {
				break
			}
		}
		if msg, err := clientSock.RecvMessage(); err != nil {
			fmt.Printf("user registration to p2p server failed.")
		} else if string(msg[0]) == "ok" {
			fmt.Printf("user registration to p2p server completed.")
		} else {
			fmt.Printf("user registration to p2p server failed.")
		}
		fmt.Println("starting message transfer procedure.")
		p2pRx, err := goczmq.NewSub(fmt.Sprintf("tcp://%s:%s", ipAddr, portChatPublisher), "relay")
		if err != nil {
			log.Fatalln(err)
		}
		defer p2pRx.Destroy()

		p2pTx, err := goczmq.NewPush(fmt.Sprintf("tcp://%s:%s", ipAddr, portChatPublisher))
		if err != nil {
			log.Fatalln(err)
		}
		defer p2pTx.Destroy()

		fmt.Println("starting autonomous message transmit and receive scenario.")
		for {
			if p2pRx.Pollin() {
				totalMsg, err := p2pRx.RecvMessage()
				if err != nil {
					log.Printf("[ERROR] Failed to receive message: %v", err)
				}
				msgs := strings.Split(string(totalMsg[0]), ":")
				fmt.Printf("p2p-recv::<<== %s:%s", msgs[1], msgs[2])
			} else {
				r := getRandRange(1, 100)
				if r < 10 {
					time.Sleep(3 * time.Second)
					msg := "(" + userName + "," + ipAddr + ":ON)"
					if err := p2pTx.SendFrame([]byte(msg), 0); err != nil {
						log.Printf("[ERROR] Failed to send message: %v", err)
					} else {
						fmt.Printf("p2p-send::==>> %s", msg)
					}
				}
			}
		}

	} else if errors.Is(err, NameServerNotFoundError) {
		ipAddrP2pServer = ipAddr
		fmt.Println("p2p server is not found, and p2p server mode is activated.")
		go beaconNameserver(ipAddr, portNameserver)
		fmt.Println("p2p beacon server is activated.")
		go userMangerNameserver(ipAddr, portSubscribe)
		fmt.Println("p2p subscriber database server is activated.")
		go relayServerNameserver(ipAddr, portChatPublisher, portChatCollector)
		fmt.Println("p2p message relay server is activated.")
	} else {
		log.Fatalln(err)
	}
}
