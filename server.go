// Stores functionality for responding to messages.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
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

		var messagesToPiggyback = GetUnexpiredPiggybackMessages()

		switch message.Kind {
		case PING:
			fmt.Println("Received PING: ", message)
			var messages Messages
			err = json.Unmarshal([]byte(message.Data), &messages)

			if err != nil {
				fmt.Println("Failed to unmarshal PING messages, skipping")
				continue
			}

			// Each PING contains multiple messages within it.
			// TODO The same processing logic should be applied on ACK side.
			for _, subMessage := range messages {
				switch subMessage.Kind {
				case HELLO:
					ProcessHelloMessage(subMessage)
				case LEAVE:
					ProcessFailOrLeaveMessage(subMessage)
				case FAIL:
					ProcessFailOrLeaveMessage(subMessage)
				default:
					log.Fatalf("Unexpected message kind")
				}
			}
		case JOIN:
			fmt.Println("Received JOIN")
			responseMessage, err := ProcessJoinMessage(message, address)
			if err != nil {
				log.Fatalf("Failed to process join message")
			}
			// Don't piggyback anything, just return the join response.
			messagesToPiggyback = Messages{responseMessage}
		case LEAVE:
			ProcessFailOrLeaveMessage(message)
		default:
			log.Fatalf("Unexpected message kind")
		}

		fmt.Println("Count of messages in ACK: ", len(messagesToPiggyback))
		ackResponse, err := EncodeAckMessage(messagesToPiggyback)
		if err != nil {
			fmt.Println("Failed to generate response.")
			continue
		}

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
		// You should simply add this node to your list, if it does not exist already,
		// or if you ain't it.
		return Message{}, fmt.Errorf("Unexpected JOIN message received for non Introducer node")
	}
}

func getPingResponse([]byte, error) string {
	response := Message{
		Kind: ACK,
		Data: "",
	}

	responseEnc, _ := json.Marshal(response)

	return string(responseEnc)
}

func ProcessHelloMessage(message Message) error {
	fmt.Println("Processing Hello Message: ", message)

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

	AddPiggybackMessage(message, len(membershipInfo))

	return nil
}

func ProcessFailOrLeaveMessage(message Message) error {

	fmt.Println("Processing Fail/Leave Message: ", message)

	// For the fail message, Data is expected to be the node Id.
	nodeId := message.Data

	// If it's you, be very confused.
	if nodeId == NODE_ID {
		os.Exit(0)
	}

	_, ok := GetMemberInfo(nodeId)

	if ok { // node exists in membership info, remove and disseminate
		fmt.Printf("Node %s exists in membership info, removing \n", nodeId)

		DeleteMember(nodeId)

		// disseminating info that the node left
		AddPiggybackMessage(message, len(membershipInfo))

		return nil
	}

	return nil
}
