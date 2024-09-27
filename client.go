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

			fmt.Println("Sender has messages: ", len(messagesToPiggyback))

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
				fmt.Printf("%s timed out\n", nodeId)

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

			// TODO Process piggyback information.
			fmt.Printf("Received ACK %s Response: [%s] \n", nodeId, buffer[:mLen])

			messages, err := DecodeAckMessage(buffer[:mLen])
			fmt.Println("Messages in ACK: ", len(messages))
			if err != nil {
				fmt.Printf("Unable to decode ACK message from node: %s", nodeId)
				continue
			}

			for _, subMessage := range messages {
				switch subMessage.Kind {
				case HELLO:
					ProcessHelloMessage(subMessage)
				case LEAVE:
					ProcessFailOrLeaveMessage(subMessage)
				case FAIL:
					ProcessFailOrLeaveMessage(subMessage)
				default:
					fmt.Printf("Unexpected message kind")
				}
			}

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
