package main

import (
	"log"
	"os"
)

var membershipInfo map[string]MemberInfo = make(map[string]MemberInfo)

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

		NODE_ID = InitializeMembershipInfoAndList(members, introducer_conn, localIP)

		helloMessage := Message{
			Kind: HELLO,
			Data: NODE_ID,
		}

		AddToPiggybacks(helloMessage, len(membershipInfo))
	} else {
		NODE_ID = ConstructNodeID(INTRODUCER_SERVER_HOST)
	}

	// Dial connection.
	go startSender()

	var b []byte = make([]byte, 1)

	go func() {
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
	}()

	ch := make(chan int)
	<-ch
}
