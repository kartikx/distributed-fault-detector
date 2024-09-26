package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// TODO @kartikr2 This should be indexed by the identifier.
var membershipInfo map[string]MemberInfo

// Stores the identifiers, gets shuffled and round-robinned for pings.
var membershipList = []string{}

func main() {
	num := os.Args[1]

	// TODO this should be populated via introduction.
	switch num {
	case "0":
		membershipList = append(membershipList, "fa24-cs425-6402.cs.illinois.edu", "fa24-cs425-6403.cs.illinois.edu")
	case "1":
		membershipList = append(membershipList, "fa24-cs425-6401.cs.illinois.edu", "fa24-cs425-6403.cs.illinois.edu")
	case "2":
		membershipList = append(membershipList, "fa24-cs425-6401.cs.illinois.edu", "fa24-cs425-6402.cs.illinois.edu")
	}

	// TODO write a logging abstraction to direct all logs into a file.
	localIP, err := GetLocalIP()

	if err != nil {
		log.Fatalf("Unable to get local IP")
	}

	fmt.Println(localIP)

	// TODO @kartikr2 This should be a check on VM name instead.
	// isIntroducer := listenPort == INTRODUCER_PORT
	// if !isIntroducer {
	// introduce(name, listenPort)
	// }

	// go startListener()

	fmt.Println("Sleeping")
	time.Sleep(10 * time.Second)
	fmt.Println("Awake")

	// Dial connection.
	// go startSender()

	// to force waiting.
	ch := make(chan int)
	<-ch
}
