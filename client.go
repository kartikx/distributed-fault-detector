// Stores functionality for initiating messages.

package main

import (
	"fmt"
	"time"
)

var piggybacks PiggybackMessages

func startSender() {
	// TODO I feel we should start this after introduction is complete and everything is stable.
	// You can do this via a channel.
	time.Sleep(5 * time.Second)

	for {
		// TODO Make this asynchronous using a goroutine.

		members := GetMembers()
		Shuffle(members)

		for _, nodeId := range members {
			connection := *membershipInfo[nodeId].connection

			if connection == nil {
				fmt.Println("Connection is nil")

				// Perhaps connection is still being made. Sleep for some time.
				time.Sleep(2 * time.Second)
				continue
			}

			var messages Messages

			// TODO This could go in a separate function.
			// TODO This needs to be done on ACKs as well.
			fmt.Println("Piggybacks were: ", piggybacks)

			for index := 0; index < len(piggybacks); index++ {
				if piggybacks[index].ttl > 0 {
					messages = append(messages, piggybacks[index].message)
					piggybacks[index].ttl--
				}

				if piggybacks[index].ttl <= 0 {
					piggybacks = append(piggybacks[:index], piggybacks[index+1:]...)
					index--
				}
			}

			fmt.Println("Piggybacks are: ", piggybacks)

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

				AddToPiggybacks(failedMessage, len(membershipInfo))

				continue
			} else {
				// TODO Ack might have important information, process it.
				fmt.Println("ACK: ", nodeId)
			}

			// TODO remove.
			time.Sleep(PING_INTERVAL * time.Second)
		}
	}
}

func ExitGroup() {

	fmt.Printf("Exiting gracefully %s\n", NODE_ID)

	// Leave message just contains the NODE_ID
	leaveMessage := Message{Kind: LEAVE, Data: NODE_ID}
	leaveMessageEnc, err := json.Marshal(leaveMessage)
	if err != nil {
		fmt.Println("Unable to encode leave message")
		continue
	}

	for _, nodeId := range membershipList {
		
		connection := *membershipInfo[nodeId].connection
		if connection != nil {
			fmt.Printf("Exiting gracefully %s sent to %s\n", NODE_ID, nodeId)
			connection.Write(leaveMessageEnc)
			connection.Close()
		}
		
	}

	// TODO close the log file

	os.Exit()

}
