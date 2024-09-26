package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// TODO implement.
func introduce() ([]string, *net.Conn, error) {
	conn, err := net.Dial("udp", GetServerEndpoint(INTRODUCER_SERVER_HOST))
	if err != nil {
		return nil, nil, err
	}

	// It could pass its IP in?
	joinMessage := Message{Kind: JOIN, Data: ""}

	// Create helper for encoding/decoding + error checks.
	joinMessageEnc, err := json.Marshal(Messages{joinMessage})
	if err != nil {
		return nil, nil, err
	}

	// I could construct a helper for this.
	pingMessage := Message{Kind: PING, Data: string(joinMessageEnc)}

	messageEnc, err := json.Marshal(pingMessage)
	if err != nil {
		return nil, nil, err
	}

	conn.Write(messageEnc)

	buffer := make([]byte, 1024)
	// fmt.Println("%s waiting for a response")
	mLen, err := conn.Read(buffer)
	if err != nil {
		return nil, nil, err
	}

	var response Message
	err = json.Unmarshal(buffer[:mLen], &response)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Received: ", response)

	// TODO Could use a struct for this.
	var members []string
	err = json.Unmarshal([]byte(response.Data), &members)
	if err != nil {
		return nil, nil, err
	}

	return members, &conn, nil
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
