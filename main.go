package main

import (
	"log"
	"os"
)

var NODE_ID = ""

var isIntroducer = false

func main() {
	// Synchronizes start of client and server.
	clientServerChan := make(chan int, 5)

	// Listener is started even before introduction so that the
	// introducer can make a connection to this node.
	go startServer(clientServerChan)

	// TODO @sdevata2 write a logging abstraction to direct all logs into a file.
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

		NODE_ID = InitializeMembershipInfoAndList(members, introducer_conn, localIP)

		helloMessage := Message{
			Kind: HELLO,
			Data: NODE_ID,
		}

		AddToPiggybacks(helloMessage, len(membershipInfo))
	} else {
		NODE_ID = ConstructNodeID(INTRODUCER_SERVER_HOST)
	}

	clientServerChan <- 1

	// Dial connection.
	go startClient(clientServerChan)

	var b []byte = make([]byte, 1)

	// TODO make this more elaborate and in-line with demo expectations.
	for {
		os.Stdin.Read(b)

		switch b[0] {
		case 'm':
			PrintMembershipInfo()
		case 'p':
			PrintPiggybackMessages()
		case 'e':
			ExitGroup()
		}
	}
}
