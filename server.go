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
		var responseEnc []byte

		fmt.Println("Server received message: ", message)

		switch message.Kind {
		case PING:
			var messages Messages
			err = json.Unmarshal([]byte(message.Data), &messages)
			for _, subMessage := range messages {
				switch subMessage.Kind {
				case JOIN:
					fmt.Println("Case: JOIN")
					// TODO this will differ depending on whether you are introducer or not.
					// How do i handle multiple messages?
					responseEnc, err = ProcessJoinMessage(subMessage.Data, address)
					if err != nil {
						log.Fatalf("Failed to process join message")
					}
				case PING:
					responseEnc, _ = ProcessPingMessage(subMessage.Data, address)
				default:
					log.Fatalf("Unexpected message kind")
				}
			}

			// Adding a random sleep to simulate failures.
			// var sleepTime time.Duration = time.Duration(rand.Intn(4)) * time.Second
			// fmt.Println("PING from: ", address, " Sleep for: ", sleepTime)
			// time.Sleep(sleepTime)
		default:
			log.Fatalf("Unexpected message kind")
		}

		fmt.Println("Writing: ", responseEnc)
		server.WriteToUDP(responseEnc, address)
	}
}

func ProcessJoinMessage(request string, addr *net.UDPAddr) ([]byte, error) {
	if isIntroducer {
		// TODO Add corner case checking, what if the introducer gets a looped around message from
		// the past? It should check that the node doesn't already exist.

		fmt.Println("Join message body: ", request)

		ipAddr := addr.IP.String()
		nodeId := ConstructNodeID(ipAddr)

		fmt.Printf("IP: %s NodeID: %s", ipAddr, nodeId)

		membershipList = append(membershipList, nodeId)

		// For the response, add yourself to the list as well.
		membershipListResponse := append(membershipList, NODE_ID)

		conn, err := net.Dial("udp", GetServerEndpoint(ipAddr))
		if err != nil {
			return nil, err
		}

		membershipInfo[nodeId] = MemberInfo{
			connection: &conn,
			host:       ipAddr,
			failed:     false,
		}

		// Informing other nodes to add this node to their lists via piggybacks.
		joinPiggybackMessage := Message{
			Kind: JOIN,
			Data: nodeId,
		}

		piggybacks = append(piggybacks, PiggbackMessage{
			message: joinPiggybackMessage,
			ttl:     len(membershipList),
		})

		responseEnc, err := json.Marshal(membershipListResponse)
		if err != nil {
			return nil, err
		}

		response := Message{
			Kind: JOIN,
			Data: string(responseEnc),
		}

		responseEnc, err = json.Marshal(response)

		return responseEnc, nil
	} else {
		// You should simply add this node to your list, if it does not exist already,
		// or if you ain't it.
		return nil, nil
	}
}

func ProcessPingMessage(request string, addr *net.UDPAddr) ([]byte, error) {
	response := Message{
		Kind: ACK,
		Data: "",
	}

	responseEnc, _ := json.Marshal(response)

	return responseEnc, nil
}
