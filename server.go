// Stores functionality for responding to messages.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func startServer(clientServerChan chan int) {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(SERVER_HOST),
		Port: SERVER_PORT,
		Zone: "",
	}

	server, err := net.ListenUDP("udp", addr)

	if err != nil {
		log.Fatalf("Couldn't start server: %s", err.Error())
	}

	clientServerChan <- 1

	for {
		buf := make([]byte, 1024)
		mlen, address, err := server.ReadFromUDP(buf)

		if err != nil {
			log.Fatalf("Error accepting: %s", err.Error())
		}

		var message Message
		json.Unmarshal(buf[:mlen], &message)

		// TODO @kartikr2 Not piggybackking on ACK for now. It was causing issues in aligning membership lists.
		var messagesToPiggyback = GetUnexpiredPiggybackMessages()

		// fmt.Println("Server has messages: ", len(messagesToPiggyback))

		switch message.Kind {
		case PING:
			PrintMessage("Incoming", message, "")
			var messages Messages
			err = json.Unmarshal([]byte(message.Data), &messages)

			if err != nil {
				fmt.Println("Failed to unmarshal PING messages, skipping")
				continue
			}

			// Each PING contains multiple messages within it.
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
					log.Fatalf("Unexpected submessage kind in PING")
				}
			}
		case JOIN:
			PrintMessage("Incoming", message, "")
			responseMessage, err := ProcessJoinMessage(message, address)
			if err != nil {
				log.Fatalln("Failed to process join message", message)
			}
			// Don't piggyback anything, just return the join response.
			messagesToPiggyback = Messages{responseMessage}
		case LEAVE:
			ProcessFailOrLeaveMessage(message)
		default:
			log.Fatalln("Unexpected message kind: ", message)
		}

		ackResponse, err := EncodeAckMessage(messagesToPiggyback)
		if err != nil {
			fmt.Println("Failed to generate response.")
			continue
		}

		var ackMessage Message
		err = json.Unmarshal(ackResponse, &ackMessage)
		if err != nil {
			fmt.Println("Unable to decode outgoing ACK message")
			continue
		}
		PrintMessage("outgoing", message, "")

		server.WriteToUDP(ackResponse, address)
	}
}

// request contains the encoded Data of the JOIN message.
// addr is the address of the host that sent this PING.
func ProcessJoinMessage(message Message, addr *net.UDPAddr) (Message, error) {
	if isIntroducer {
		joinResponse, err := IntroduceNodeToGroup(message.Data, addr)
		return joinResponse, err
	} else {
		return Message{}, fmt.Errorf("Unexpected JOIN message received for non Introducer node")
	}
}

func ProcessHelloMessage(message Message) error {
	PrintMessage("incoming", message, "")

	// For the hello message, nodeId is expected to be the node Id.
	nodeId := message.Data

	_, ok := GetMemberInfo(nodeId)

	if ok || nodeId == NODE_ID {
		fmt.Printf("Node %s already exists in membership info, Skipping \n", nodeId)
		return nil
	}

	err := AddNewMemberToMembershipInfo(nodeId)
	if err != nil {
		return err
	}

	AddPiggybackMessage(message)

	return nil
}

func ProcessFailOrLeaveMessage(message Message) error {
	PrintMessage("incoming", message, "")

	// For the fail message, Data is expected to be the node Id.
	nodeId := message.Data

	// If it's you, be very confused.
	if nodeId == NODE_ID {
		fmt.Println("Received self failure message.")
		os.Exit(0)
	}

	_, ok := GetMemberInfo(nodeId)

	if ok { // node exists in membership info, remove and disseminate
		fmt.Printf("Node %s exists in membership info, removing \n", nodeId)

		DeleteMember(nodeId)

		// disseminating info that the node left
		AddPiggybackMessage(message)

		return nil
	}

	return nil
}

func ProcessSuspectMessage(message Message) error {
	PrintMessage("incoming", message, "")

	if !inSuspectMode {
		fmt.Printf("Received a SUSPECT message when not in suspect mode")
		return fmt.Errorf("SUSPECT message but not in suspect mode")
	}

	// SUSPECT message will be of type incarnation@IP@timestamp
	parts := strings.Split(message.Data, "@")
	message_incarnation, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Printf("Unable to get incarnation number from a SUSPECT message", message_incarnation)
		return nil
	}
	nodeId := fmt.Sprintf("%s@%s", parts[1], parts[2])

	if nodeId == NODE_ID {

		// If the self SUSPECT message is for an old self, ignore since the node already disseminated ALIVE
		if message_incarnation < INCARNATION {
			fmt.Printf("Received a SUSPECT message for an old self")
			return nil
		}

		// If a node finds out that it is being suspected, it will increment incarnation and disseminate an ALIVE
		INCARNATION += 1
		aliveMessage := Message{Kind: ALIVE, Data: string(INCARNATION) + "@" + nodeId}
		AddPiggybackMessage(aliveMessage)

		return nil
	} else {

		// Check the message incarnation
		membershipInfoEntry, _ := GetMemberInfo(nodeId)
		if (membershipInfoEntry.incarnation > message_incarnation) || membershipInfoEntry.suspected {
			// You have a more recent incarnation or are suspecting, ignore the new SUSPECT message
			return nil
		}

		MarkMemberSuspected(nodeId)

		// Disseminate the SUSPECT message that you got
		AddPiggybackMessage(message)

		// Wait for suspect timeout
		time.Sleep(time.Second * SUSPECT_TIMEOUT)

		// Check incarnation number in membership info, then disseminate FAIL
		membershipInfoEntry, _ = GetMemberInfo(nodeId)
		if membershipInfoEntry.incarnation > message_incarnation {
			// You have a more recent incarnation, perhaps the suspected node sent an ALIVE
			return nil
		} else {
			failMessage := Message{Kind: FAIL, Data: nodeId}
			ProcessFailOrLeaveMessage(failMessage)
		}
	}

	return nil
}

func ProcessAliveMessage(message Message) error {
	PrintMessage("incoming", message, "")

	if !inSuspectMode {
		fmt.Printf("Received an ALIVE message when not in suspect mode")
		return fmt.Errorf("ALIVE message but not in suspect mode")
	}

	// ALIVE message will be of type incarnation@IP@timestamp
	parts := strings.Split(message.Data, "@")
	message_incarnation, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Printf("Unable to get incarnation number from a SUSPECT message")
		return nil
	}
	nodeId := fmt.Sprintf("%s@%s", parts[1], parts[2])

	if nodeId == NODE_ID {
		// If it is a self ALIVE message, ignore
		return nil
	} else {

		// Check the message incarnation
		membershipInfoEntry, _ := GetMemberInfo(nodeId)
		if membershipInfoEntry.incarnation > message_incarnation {
			// You have a more recent incarnation, ignore the new ALIVE message
			return nil
		}

		// Disseminate the ALIVE message that you got
		AddPiggybackMessage(message)

		// Update the incarnation number and not suspected
		UpdateMemberIncarnation(nodeId, message_incarnation)
	}

	return nil
}

func ProcessSuspectModeMessage(message Message) error {
	PrintMessage("incoming", message, "")

	suspect_mode, err := strconv.ParseBool(message.Data)
	if err != nil {
		fmt.Printf("Was not able to parse SUSPECT_MODE message")
		return fmt.Errorf("Was not able to parse SUSPECT_MODE message")
	}

	if suspect_mode != inSuspectMode {
		inSuspectMode = suspect_mode
		AddPiggybackMessage(message)
		return nil
	}
	return nil
}
