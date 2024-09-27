// Stores functionality for responding to messages.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func startListener() {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(SERVER_HOST),
		Port: SERVER_PORT,
		Zone: "",
	}

	server, err := net.ListenUDP("udp", addr)

	if err != nil {
		log.Fatalf("Couldn't start server: %s", err.Error())
	}

	for {
		buf := make([]byte, 1024)
		mlen, address, err := server.ReadFromUDP(buf)

		if err != nil {
			log.Fatalf("Error accepting: %s", err.Error())
		}

		var message Message
		json.Unmarshal(buf[:mlen], &message)
		var responseMessage Message

		switch message.Kind {
		case PING:
			var messages Messages
			err = json.Unmarshal([]byte(message.Data), &messages)

			if err != nil {
				fmt.Println("Failed to unmarshal PING messages, skipping")
				continue
			}

			// Each PING contains multiple messages within it.
			// TODO How do i handle multiple messages? For example, 2 joins, 1 LEAVE and 1 FAIL?
			// TODO The same piggyback logic should be applied on ACK side.
			for _, subMessage := range messages {
				switch subMessage.Kind {
				case JOIN:
					fmt.Println("submessage JOIN")
					responseMessage, err = ProcessJoinMessage(subMessage, address)
					if err != nil {
						log.Fatalf("Failed to process join message")
					}
				case LEAVE:
					ProcessLeaveMessage(subMessage)
				case HELLO:
					ProcessHelloMessage(subMessage)
				case FAIL:
					ProcessFailMessage(subMessage)
				default:
					log.Fatalf("Unexpected message kind")
				}
			}

			// Adding a random sleep to simulate failures.
			// var sleepTime time.Duration = time.Duration(rand.Intn(10)) * time.Second

			// if sleepTime > 8*time.Second {
			// 	fmt.Println("SIMULATING FAILURE")
			// }

			// time.Sleep(sleepTime)
		default:
			log.Fatalf("Unexpected message kind")
		}

		ackResponse, err := GetEncodedAckMessage(Messages{responseMessage})

		if err != nil {
			fmt.Println("Failed to generate response.")
			continue
		}

		// fmt.Println(ackResponse)
		// fmt.Println("Writing Response: ", string(ackResponse))
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

	_, ok := membershipInfo[nodeId]

	if ok || nodeId == NODE_ID {
		fmt.Printf("Node %s already exists in membership info, Skipping \n", nodeId)
		return nil
	}

	err := AddNewMemberToMembershipInfo(nodeId)
	if err != nil {
		return err
	}

	AddToPiggybacks(message, len(membershipInfo))

	return nil
}

func ProcessFailMessage(message Message) {
	fmt.Println("Received fail message: ", message)

	// If you are the fail, just leave.

	// If you already knew, don't propagate.

	// Else, assign and propagate.
}

func ProcessLeaveMessage(message Message) {
	// If it's you, be very confused.

	// If you already knew, don't propagate.

	// Else, assign and propagate.
}
