package main

import (
	"fmt"
	"log"
	"net"
)

var membershipInfo map[string]MemberInfo = make(map[string]MemberInfo)

// Stores the identifiers, gets shuffled and round-robinned for pings.
var membershipList = []string{}

var NODE_ID = ""

var isIntroducer = false

func main() {
	// Do this at the start, so that the introducer can connect to you.
	go startListener()

	// TODO write a logging abstraction to direct all logs into a file.
	localIP, err := GetLocalIP()

	fmt.Println("IP: ", localIP)

	if err != nil {
		log.Fatalf("Unable to get local IP")
	}

	if localIP == INTRODUCER_SERVER_HOST {
		isIntroducer = true
	}

	if !isIntroducer {
		members, introducer_conn, err := introduce()
		fmt.Println("Received members: ", members)
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

	fmt.Println("Printing membership info table")
	for k, _ := range membershipInfo {
		fmt.Printf("Node Id: %s\n", k)
	}

	// Now process your membershiplist into membershipinfo

	// Dial connection.
	// go startSender()

	// to force waiting.
	ch := make(chan int)
	<-ch
}
