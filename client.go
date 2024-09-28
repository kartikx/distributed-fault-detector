// Stores functionality for initiating messages.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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

			var pingMessage Message
			err = json.Unmarshal(pingMessageEnc, &pingMessage)
			if err != nil {
				fmt.Println("Unable to decode outgoing PING message")
				continue
			}
			printMessage("outgoing", pingMessage, nodeId)

			connection.Write(pingMessageEnc)

			buffer := make([]byte, 1024)

			// TODO would this work would even if I were to re-use the connection?
			connection.SetReadDeadline(time.Now().Add(TIMEOUT_DETECTION_MILLISECONDS * time.Millisecond))
			mLen, err := connection.Read(buffer)

			if err != nil {
				fmt.Printf("%s timed out\n", nodeId)

				// In suspicion, you would want to suspect it first.
				if inSuspectMode {
					// Create a SUSPECT message to process and disseminate
					member, _ := GetMemberInfo(nodeId)
					suspectMessage := Message{Kind: SUSPECT, Data: strconv.Itoa(member.incarnation) + "@" + nodeId}
					printMessage("outgoing", suspectMessage, suspectMessage.Data)
					go ProcessSuspectMessage(suspectMessage)

					continue
				} else { // Otherwise, just mark the node as failed

					DeleteMember(nodeId)

					// Start propagating FAIL message.
					failedMessage := Message{
						Kind: FAIL,
						Data: nodeId,
					}

					AddPiggybackMessage(failedMessage, len(membershipInfo))

					continue
				}
			}
			// TODO simulate drops on receiver end.

			messages, err := DecodeAckMessage(buffer[:mLen])
			// fmt.Println("Messages in ACK: ", len(messages))
			if err != nil {
				fmt.Printf("Unable to decode ACK message from node: %s", nodeId)
				continue
			}

			var ackMessage Message
			err = json.Unmarshal(buffer[:mLen], &ackMessage)
			if err != nil {
				continue
			}
			printMessage("incoming", ackMessage, nodeId)

			for _, subMessage := range messages {
				switch subMessage.Kind {
				case HELLO:
					ProcessHelloMessage(subMessage)
				case LEAVE:
					ProcessFailOrLeaveMessage(subMessage)
				case FAIL:
					ProcessFailOrLeaveMessage(subMessage)
				case SUSPECT:
					go ProcessSuspectMessage(subMessage)
				case ALIVE:
					ProcessAliveMessage(subMessage)
				case SUSPECT_MODE:
					ProcessSuspectModeMessage(subMessage)
				default:
					log.Fatalf("Unexpected submessage kind in ACK")
				}
			}

			// TODO remove.
			time.Sleep(PING_INTERVAL_MILLISECONDS * time.Millisecond)
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

			var leaveMessage Message
			err = json.Unmarshal(leaveMessageEnc, &leaveMessage)
			if err != nil {
				continue
			}
			printMessage("outgoing", leaveMessage, nodeId)

			connection.Write(leaveMessageEnc)
			connection.Close()
		}
	}

	// TODO close the log file

	os.Exit(0)

}

func StartSuspecting() {
	suspectMessage := Message{Kind: SUSPECT_MODE, Data: "true"}
	AddPiggybackMessage(suspectMessage, len(membershipInfo))
}

func StopSuspecting() {
	suspectMessage := Message{Kind: SUSPECT_MODE, Data: "false"}
	AddPiggybackMessage(suspectMessage, len(membershipInfo))
}
