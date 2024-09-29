package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

var NODE_ID = ""
var LOCAL_IP = ""
var INCARNATION = 0

var inSuspectMode = false
var isIntroducer = false

func main() {
	// Synchronizes start of client and server.
	clientServerChan := make(chan int, 5)

	// Listener is started even before introduction so that the
	// introducer can make a connection to this node.
	go startServer(clientServerChan)

	// TODO @sdevata2 write a logging abstraction to direct all logs into a file.
	LOCAL_IP, err := GetLocalIP()
	if err != nil {
		log.Fatalf("Unable to get local IP")
	}

	if LOCAL_IP == INTRODUCER_SERVER_HOST {
		isIntroducer = true
	}

	if !isIntroducer {
		members, introducer_conn, err := IntroduceYourself()
		if err != nil {
			log.Fatalf("Unable to join the group: %s", err.Error())
		}

		NODE_ID = InitializeMembershipInfoAndList(members, introducer_conn, LOCAL_IP)

		helloMessage := Message{
			Kind: HELLO,
			Data: NODE_ID,
		}

		AddPiggybackMessage(helloMessage)
	} else {
		NODE_ID = ConstructNodeID(INTRODUCER_SERVER_HOST)
	}

	clientServerChan <- 1

	// Dial connection.
	go startClient(clientServerChan)

	var b []byte = make([]byte, 1)

	os_signals := make(chan os.Signal, 1)
	signal.Notify(os_signals, os.Interrupt)
	go func() {
		for sig := range os_signals {
			// sig is a ^C, handle it
			fmt.Println("Application got an OS interrupt:", sig, "at", time.Now().Format(time.RFC3339))
			os.Exit(0)
		}
	}()

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
		case 's':
			fmt.Printf("ID: %s\n", NODE_ID)
			fmt.Printf("Incarnation: %d\n", INCARNATION)
			fmt.Printf("InSuspectMode: %t\n", inSuspectMode)
		case 'd':
			StartSuspecting()
		case 'n':
			StopSuspecting()
		}
	}
}
