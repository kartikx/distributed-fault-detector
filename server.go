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

		switch message.Kind {
		case JOIN:
			responseEnc, err = ProcessJoinMessage(string(buf[:mlen]), addr)
		case PING:
			responseEnc, err = ProcessPingMessage(string(buf[:mlen]), addr)
			// Adding a random sleep to simulate failures.
			// var sleepTime time.Duration = time.Duration(rand.Intn(4)) * time.Second
			// fmt.Println("PING from: ", address, " Sleep for: ", sleepTime)
			// time.Sleep(sleepTime)
		}

		server.WriteToUDP(responseEnc, address)
	}
}

func ProcessJoinMessage(request string, addr *net.UDPAddr) ([]byte, error) {
	fmt.Println("Join message body: ", request)

	// TODO get ip address, construct node id, get existing member list, construct response body.
	ipAddr := addr.IP.String()
	nodeId := ConstructNodeID(ipAddr)

	membershipList = append(membershipList, nodeId)

	conn, err := net.Dial("udp", GetServerEndpoint(ipAddr))

	if err != nil {
		return nil, nil
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
