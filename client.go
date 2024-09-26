// Stores functionality for initiating messages.

package main

import (
	"encoding/json"
	"fmt"
	"time"
)

var piggybacks PiggybackMessages

func startSender() {
	for {
		// TODO Make this asynchronous using a goroutine.
		for _, member := range membershipList {
			// TODO error check?
			connection := *membershipInfo[member].connection

			fmt.Println("PING ", member)

			var messages Messages

			for _, piggyback := range piggybacks {
				messages = append(messages, piggyback.message)
			}

			messagesEnc, _ := json.Marshal(messages)

			message := Message{Kind: PING, Data: string(messagesEnc)}
			messageEnc, _ := json.Marshal(message)

			connection.Write(messageEnc)

			buffer := make([]byte, 1024)

			// TODO would this work would even if I were to re-use the connection?
			connection.SetReadDeadline(time.Now().Add(TIMEOUT_DETECTION_SECONDS * time.Second))
			_, err := connection.Read(buffer)

			if err != nil {
				fmt.Println("Add failed message for: ", member)

				// Start propagating FAIL message.
				failedMessage := Message{
					Kind: FAIL,
					Data: member,
				}

				// TODO create helper method that appends to piggyback in a thread-safe way.
				piggybacks = append(piggybacks, PiggbackMessage{message: failedMessage, ttl: 1})

				continue
			}

			// No need to read ACK. Empty response is good enough.

			fmt.Println("ACK: ", member)

			// TODO should I close this?
			// defer connection.Close()

			time.Sleep(4 * time.Second)
		}

		// TODO shuffle list here.
	}
}
