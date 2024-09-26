package main

import (
	"fmt"
	"log"
)

var membershipInfo map[string]MemberInfo = make(map[string]MemberInfo)

// Stores the identifiers, gets shuffled and round-robinned for pings.
var membershipList = []string{}

var NODE_ID = ""

var isIntroducer = false

func main() {
	// Listener is started even before introduction so that the
	// introducer can make a connection.
	go startListener()

	// TODO write a logging abstraction to direct all logs into a file.
	localIP, err := GetLocalIP()
	if err != nil {
		log.Fatalf("Unable to get local IP")
	}

	if localIP == INTRODUCER_SERVER_HOST {
		isIntroducer = true
	}

	if !isIntroducer {
		members, introducer_conn, err := IntroduceYourself()
		if err != nil {
			log.Fatalf("Unable to join the group: %s", err.Error())
		}

		NODE_ID = InitializeMembershipInfo(members, introducer_conn, localIP)

		helloMessage := Message{
			Kind: HELLO,
			Data: NODE_ID,
		}

		AddToPiggybacks(helloMessage, len(membershipList))
	} else {
		NODE_ID = ConstructNodeID(INTRODUCER_SERVER_HOST)
	}

	fmt.Println("Printing membership info table")
	for nodeId := range membershipInfo {
		fmt.Printf("Node Id: %s\n", nodeId)
	}

	fmt.Println("Printing piggybacks")
	for _, p := range piggybacks {
		fmt.Println(p)
	}

	// Dial connection.
	// go startSender()

	// to force waiting.
	ch := make(chan int)
	<-ch
}
