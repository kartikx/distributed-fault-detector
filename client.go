// Stores functionality for initiating messages.

package main

import (
	"fmt"
	"os"
	"time"
)

func startClient(clientServerChan chan int) {
	// Ensures that sending starts after listener has started and introduction is complete.
	_, _ = <-clientServerChan, <-clientServerChan

	for {
		members := GetMembers()
		Shuffle(members)

		for _, nodeId := range members {
			connection := GetNodeConnection(nodeId)

			if connection == nil {
				fmt.Printf("Node %s connection is nil, it might have been removed from the group\n", nodeId)
				continue
			}

			var messagesToPiggyback Messages = GetUnexpiredPiggybackMessages()

			pingMessageEnc, err := EncodePingMessage(messagesToPiggyback)

			if err != nil {
				fmt.Println("Unable to encode ping message")
				continue
			}

			fmt.Printf("PING %s [%s]\n", nodeId, pingMessageEnc)

			connection.Write(pingMessageEnc)

			buffer := make([]byte, 1024)

			// TODO would this work would even if I were to re-use the connection?
			connection.SetReadDeadline(time.Now().Add(TIMEOUT_DETECTION_SECONDS * time.Second))
			mLen, err := connection.Read(buffer)

			if err != nil {
				fmt.Println("%s timed out", nodeId)

				// TODO in suspicion, you would want to suspect it first.
				DeleteMember(nodeId)

				// Start propagating FAIL message.
				failedMessage := Message{
					Kind: FAIL,
					Data: nodeId,
				}

				AddPiggybackMessage(failedMessage, len(membershipInfo))

				continue
			}
			// TODO simulate drops on receiver end.

			// TODO Ack might have piggybacked information to be processed.

			fmt.Printf("Received ACK %s Response: [%s]", nodeId, buffer[:mLen])

			// TODO remove.
			time.Sleep(PING_INTERVAL * time.Second)
		}
	}
}

func ExitGroup() {

	fmt.Printf("Exiting gracefully %s\n", NODE_ID)

	// Leave message just contains the NODE_ID
	leaveMessageEnc, err := GetEncodedLeaveMessage(NODE_ID)

	if err != nil {
		fmt.Println("Unable to encode leave message")
		return
	}

	members := GetMembers()
	for _, nodeId := range members {
		connection := GetNodeConnection(nodeId)
		if connection != nil {
			fmt.Printf("Exiting gracefully %s sent to %s\n", NODE_ID, nodeId)
			connection.Write(leaveMessageEnc)
			connection.Close()
		}
	}

	// TODO close the log file

	os.Exit(0)

}
