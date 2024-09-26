package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// TODO implement.
func introduce() {
	// Send join message to the introducer.
	// fmt.Printf("%s sending JOIN message\n", listenPort)
	conn, err := net.Dial("udp", "localhost:"+INTRODUCER_PORT)

	if err != nil {
		log.Fatalln("Unable to dial introducer")
	}

	joinMessage := Message{Kind: JOIN, Data: "Let me in"}

	messageEnc, _ := json.Marshal(joinMessage)

	conn.Write(messageEnc)

	buffer := make([]byte, 1024)
	fmt.Println("%s waiting for a response")
	mLen, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	var response Message
	json.Unmarshal(buffer[:mLen], &response)
	fmt.Println("Received: ", response)
	// members, conn, err := getMembers()

	// iterate over the membership list.

	// construct a membership map by trying to connect to the members.
	// note: things can fail here.
}

func startSender() {
	for {
		// TODO Make this asynchronous using a goroutine.
		for _, member := range membershipList {
			// TODO this should be dialed already based on introduce method.
			connection, err := net.Dial("udp", fmt.Sprintf("%s:%d", member, SERVER_PORT))

			if err != nil {
				log.Fatalf("Couldn't connect to server: %s", err.Error())
			}

			fmt.Println("PING ", member)

			message := Message{Kind: PING, Data: ""}
			messageEnc, _ := json.Marshal(message)

			connection.Write(messageEnc)

			buffer := make([]byte, 1024)
			// TODO add timeout to this.

			// TODO would this would even if I were to re-use the connection?
			connection.SetReadDeadline(time.Now().Add(TIMEOUT_DETECTION_SECONDS * time.Second))
			_, err = connection.Read(buffer)

			if err != nil {
				// TODO Start propagating FAIL message.
				fmt.Println("Error reading:", err.Error())
			}

			// No need to read ACK. Empty response is good enough.

			fmt.Println("ACK: ", member)

			// TODO should I close this?
			// defer connection.Close()

			time.Sleep(2 * time.Second)
		}

		// TODO shuffle list here.
	}

}
