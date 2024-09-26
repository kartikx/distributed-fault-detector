package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

var membershipInfo map[string]MemberInfo

// Stores the identifiers, gets shuffled and round-robinned for pings.
var membershipList = []string{}

var NODE_ID = ""

func main() {
	// Do this at the start, so that the introducer can connect to you.
	go startListener()

	// TODO write a logging abstraction to direct all logs into a file.
	localIP, err := GetLocalIP()

	fmt.Println("IP: ", localIP)

	if err != nil {
		log.Fatalf("Unable to get local IP")
	}

	if localIP != INTRODUCER_SERVER_HOST {
		members, introducer_conn, err := introduce()
		if err != nil {
			log.Fatalf("Unable to join the group: %s", err.Error())
		}

		for _, id := range members {
			ip := GetIPFromID(id)

			if ip == INTRODUCER_SERVER_HOST {
				membershipInfo[id] = MemberInfo{
					connection: introducer_conn,
					host:       id,
					failed:     false,
				}
			} else if ip == localIP {
				NODE_ID = id
			} else {
				conn, err := net.Dial("udp", ip)

				if err != nil {
					// TODO what to do here? If it actually failed it should be detected by some other node.
				}

				membershipInfo[id] = MemberInfo{
					connection: &conn,
					host:       id,
					failed:     false,
				}
			}
		}
	} else {
		NODE_ID = ConstructNodeID(INTRODUCER_SERVER_HOST)
	}

	fmt.Println(membershipInfo)

	// Now process your membershiplist into membershipinfo

	fmt.Println("Sleeping")
	time.Sleep(10 * time.Second)
	fmt.Println("Awake")

	// Dial connection.
	// go startSender()

	// to force waiting.
	ch := make(chan int)
	<-ch
}
