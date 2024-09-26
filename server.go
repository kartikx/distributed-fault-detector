package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

// TODO requestPort should be an identifier
func returnMembers(selfPort, requestPort string) []string {
	var members []string

	// process its members (taking failed / suspicious nodes into account)
	for k, v := range membershipInfo {
		// TODO @kartikr2 Should I take failed / sus nodes into account here?
		members = append(members, k)

		// inform every other node that a new node has joined the system.
		sendJoinMessage(v, requestPort)
	}

	// add itself to the members
	members = append(members, selfPort)

	// add the newnode to the membership list
	addNewNodeToMembershipList(requestPort)

	// return list.
	return members
}

func getMembers() ([]string, net.Conn, error) {
	// dial connection to introducer.
	connection, err := net.Dial("udp", INTRODUCER_SERVER)

	if err != nil {
		log.Fatalf("Couldn't connect to introducer: %s", err.Error())
	}

	// send it a JOIN message.

	// read from the response.
	return []string{"8001"}, connection, nil
}

// TODO requestPort should instead be IP/Port/Identifier
func addNewNodeToMembershipList(requestPort string) {
	// TODO You'll be opening a lot of connections in this
	// program, remember to close them eventually.
	connection, err := net.Dial("udp", SERVER_HOST+":"+requestPort)

	if err != nil {
		log.Fatalf("Couldn't connect to server: %s", err.Error())
	}

	membershipInfo[requestPort] = MemberInfo{
		server:     SERVER_HOST,
		port:       requestPort,
		connection: connection,
	}

	fmt.Printf("Added %s to membership list", requestPort)
}

func sendJoinMessage(info MemberInfo, nodeId string) {
	// TODO implement.
	fmt.Println("Sending join to", info.port)
}

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
		var response Message

		switch message.Kind {
		case JOIN:
			response.Kind = JOIN
			response.Data = "JOIN RESPONSE"
		case PING:
			response.Kind = ACK
			response.Data = ""
			// Adding a random sleep to simulate failures.
			var sleepTime time.Duration = time.Duration(rand.Intn(4)) * time.Second
			fmt.Println("PING from: ", address, " Sleep for: ", sleepTime)
			time.Sleep(sleepTime)
		}

		responseEnc, _ := json.Marshal(response)
		server.WriteToUDP(responseEnc, address)
	}
}
