package main

import (
	"errors"
	"fmt"
	"github.com/zeromq/goczmq"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
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

func beaconNameserver(wg *sync.WaitGroup, localIpAddr, portNameserver string) {
	pub := createSocket(goczmq.Pub, localIpAddr, portNameserver)
	defer pub.Destroy()

	fmt.Printf("local p2p name server activated at tcp://%s:%s.\n", localIpAddr, portNameserver)
	wg.Done()
	for {
		time.Sleep(time.Second)
		msg := fmt.Sprintf("NAMESERVER:%s", localIpAddr)
		if err := pub.SendFrame([]byte(msg), 0); err != nil {
			log.Fatalf("[ERROR] Failed to send message: %v\n", err)
		}
	}
}

func userMangerNameserver(wg *sync.WaitGroup, localIpAddr, portSubscribe string) {
	var userDb [][]string

	req := createSocket(goczmq.Rep, localIpAddr, portSubscribe)
	defer req.Destroy()

	fmt.Printf("local p2p db server activated at tcp://%s:%s.\n", localIpAddr, portSubscribe)
	wg.Done()
	for {
		userReq, err := req.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to receive request: %v\n", err)
			continue
		}
		user := strings.Split(string(userReq[0]), ":")
		userDb = append(userDb, user)
		fmt.Printf("user registration '%s' from '%s'.\n", user[1], user[0])
		if err := req.SendFrame([]byte("ok"), 0); err != nil {
			log.Fatalf("[ERROR] Failed to send 'ok': %v", err)
		}
	}
}

func relayServerNameserver(wg *sync.WaitGroup, localIpAddr, portChatPublisher, portChatCollector string) {
	chatPub := createSocket(goczmq.Pub, localIpAddr, portChatPublisher)
	defer chatPub.Destroy()

	chatColl := createSocket(goczmq.Pull, localIpAddr, portChatCollector)
	defer chatColl.Destroy()

	fmt.Printf("local p2p relay server activated at tcp://%s:%s & %s.\n", localIpAddr, portChatPublisher, portChatCollector)
	wg.Done()
	for {
		msg, err := chatColl.RecvMessage()
		if err != nil {
			log.Printf("[ERROR] Failed to receive message: %v\n", err)
			break
		}
		fmt.Printf("p2p-relay:<==> %s\n", msg[0])
		if err := chatPub.SendFrame([]byte(fmt.Sprintf("RELAY:%s", msg[0])), 0); err != nil {
			log.Fatalf("[ERROR] Failed to send message: %v\n", err)
		}
	}
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

func createSocket(sockType int, addr string, port string) *goczmq.Sock {
	var serverish bool

	fullAddr := fmt.Sprintf("tcp://%s:%s", addr, port)
	sock := goczmq.NewSock(sockType)
	if sockType == goczmq.Req || sockType == goczmq.Push {
		serverish = false
	} else {
		serverish = true
	}
	if err := sock.Attach(fullAddr, serverish); err != nil {
		log.Fatalln(err)
	}
	return sock
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

	wg := sync.WaitGroup{}

	userName := os.Args[1]
	ipAddr := getLocalIpAddr()
	getIpMask := func() string {
		addr := strings.Split(ipAddr, ".")
		return addr[0] + "." + addr[1] + "." + addr[2]
	}
	ipMask := getIpMask()
	fmt.Println("searching for p2p server.")
	nameServerIpAddr, err := searchNameserver(ipMask, ipAddr, portNameserver)
	if err == nil {
		ipAddrP2pServer = nameServerIpAddr
		fmt.Printf("p2p server found at %s, and p2p client mode is activated\n", ipAddrP2pServer)
	} else if errors.Is(err, NameServerNotFoundError) {
		wg.Add(3)
		fmt.Println("p2p server is not found, and p2p server mode is activated.")

		ipAddrP2pServer = ipAddr

		go beaconNameserver(&wg, ipAddr, portNameserver)

		go userMangerNameserver(&wg, ipAddr, portSubscribe)

		go relayServerNameserver(&wg, ipAddr, portChatPublisher, portChatCollector)
	} else {
		log.Fatalln(err)
	}
	wg.Wait()
	fmt.Println("starting user registration procedure.")

	clientSock := createSocket(goczmq.Req, ipAddrP2pServer, portSubscribe)
	defer clientSock.Destroy()
	log.Printf("Hello")

	msg := fmt.Sprintf("%s:%s", ipAddr, userName)
	for {
		if err := clientSock.SendFrame([]byte(msg), 0); err == nil {
			break
		}
	}
	if msg, err := clientSock.RecvMessage(); err != nil {
		fmt.Println("user registration to p2p server failed.")
	} else if string(msg[0]) == "ok" {
		fmt.Println("user registration to p2p server completed.")
	} else {
		fmt.Println("user registration to p2p server failed.")
	}
	fmt.Println("starting message transfer procedure.")

	p2pRx, err := goczmq.NewSub(fmt.Sprintf("tcp://%s:%s", ipAddr, portChatPublisher), "RELAY")
	if err != nil {
		log.Fatalln(err)
	}
	defer p2pRx.Destroy()

	p2pTx, err := goczmq.NewPush(fmt.Sprintf("tcp://%s:%s", ipAddr, portChatCollector))
	if err != nil {
		log.Fatalln(err)
	}
	defer p2pTx.Destroy()

	fmt.Println("starting autonomous message transmit and receive scenario.")
	for {
		if p2pRx.Pollin() {
			totalMsg, err := p2pRx.RecvMessage()
			if err != nil {
				log.Printf("[ERROR] Failed to receive message: %v\n", err)
			}
			msgs := strings.Split(string(totalMsg[0]), ":")
			fmt.Printf("p2p-recv::<<== %s:%s\n", msgs[1], msgs[2])
		} else {
			r := getRandRange(1, 100)
			if r < 10 {
				time.Sleep(3 * time.Second)
				msg := "(" + userName + "," + ipAddr + ":ON)"
				if err := p2pTx.SendFrame([]byte(msg), 0); err != nil {
					log.Printf("[ERROR] Failed to send message: %v\n", err)
				} else {
					fmt.Printf("p2p-send::==>> %s\n", msg)
				}
			} else if r > 90 {
				time.Sleep(3 * time.Second)
				msg := "(" + userName + "," + ipAddr + ":OFF)"
				if err := p2pTx.SendFrame([]byte(msg), 0); err != nil {
					log.Printf("[ERROR] Failed to send message: %v\n", err)
				} else {
					fmt.Printf("p2p-send::==>> %s\n", msg)
				}
			}
		}
	}
}
