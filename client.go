package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

func introduce() {
	// Send join message to the introducer.
	// fmt.Printf("%s sending JOIN message\n", listenPort)
	conn, err := net.Dial("udp", "localhost:"+INTRODUCER_PORT)

	if err != nil {
		log.Fatalln("Unable to dial introducer")
	}

	joinMessage := Message{Kind: "JOIN", Data: "Let me in"}

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
	fmt.Println("Dialing connection")

	for _, member := range membershipList {
		connection, err := net.Dial("udp", fmt.Sprintf("%s:%d", member, SERVER_PORT))

		if err != nil {
			log.Fatalf("Couldn't connect to server: %s", err.Error())
		}

		message := Message{Kind: "PING", Data: "hello, it's me"}
		messageEnc, _ := json.Marshal(message)
		connection.Write(messageEnc)
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}

		var response Message
		json.Unmarshal(buffer[:mLen], &response)
		fmt.Println("Received: ", response)

		// TODO should I close this?
		defer connection.Close()

		time.Sleep(2 * time.Second)
	}

}
