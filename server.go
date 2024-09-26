package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// func sendJoinMessage(info MemberInfo, nodeId string) {
// 	// TODO implement.
// 	fmt.Println("Sending join to", info.port)
// }

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
	// TODO Add a check here for whether you are the introducer or not.

	// TODO Add corner case checking, what if the introducer gets a looped around message from
	// the past? It should check that the node doesn't already exist.

	fmt.Println("Join message body: ", request)

	// TODO get ip address, construct node id, get existing member list, construct response body.
	ipAddr := addr.IP.String()
	nodeId := ConstructNodeID(ipAddr)

	fmt.Printf("IP: %s NodeID: %s", ipAddr, nodeId)

	// Add new node as well as yourself to the list.
	membershipList = append(membershipList, nodeId, NODE_ID)

	conn, err := net.Dial("udp", GetServerEndpoint(ipAddr))
	if err != nil {
		return nil, err
	}

	membershipInfo[nodeId] = MemberInfo{
		connection: &conn,
		host:       ipAddr,
		failed:     false,
	}

	// TODO inform other nodes to also add this node to their lists.

	responseEnc, err := json.Marshal(membershipList)
	if err != nil {
		return nil, err
	}

	response := Message{
		Kind: JOIN,
		Data: string(responseEnc),
	}

	responseEnc, err = json.Marshal(response)

	return responseEnc, nil
}

func ProcessPingMessage(request string, addr *net.UDPAddr) ([]byte, error) {
	response := Message{
		Kind: ACK,
		Data: "",
	}

	responseEnc, _ := json.Marshal(response)

	return responseEnc, nil
}
