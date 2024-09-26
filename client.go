// Stores functionality for initiating messages.

package main

import (
	"fmt"
	"time"
)

var piggybacks PiggybackMessages

func startSender() {
	for {
		// TODO Make this asynchronous using a goroutine.
		for _, nodeId := range membershipList {
			// TODO error check?
			connection := *membershipInfo[nodeId].connection

			if connection == nil {
				// Perhaps connection is still being made. Sleep for some time.
				time.Sleep(2 * time.Second)
				continue
			}

			var messages Messages

			for _, piggyback := range piggybacks {
				// TODO check TTL of message.
				messages = append(messages, piggyback.message)
			}

			pingMessageEnc, err := GetEncodedPingMessage(messages)

			if err != nil {
				fmt.Println("Unable to encode ping message")
				continue
			}

			fmt.Printf("PING %s [%s]\n", nodeId, pingMessageEnc)

			connection.Write(pingMessageEnc)

			buffer := make([]byte, 1024)

			// TODO would this work would even if I were to re-use the connection?
			connection.SetReadDeadline(time.Now().Add(TIMEOUT_DETECTION_SECONDS * time.Second))
			_, err = connection.Read(buffer)

			if err != nil {
				fmt.Println("Add failed message for: ", nodeId)

				// Start propagating FAIL message.
				failedMessage := Message{
					Kind: FAIL,
					Data: nodeId,
				}

				AddToPiggybacks(failedMessage, 1)

				continue
			} else {
				// TODO Ack might have important information, process it.
				fmt.Println("ACK: ", nodeId)

				// TODO should I close this?
				// defer connection.Close()
			}

			time.Sleep(PING_INTERVAL * time.Second)
		}

		// TODO shuffle list here.
	}
}
