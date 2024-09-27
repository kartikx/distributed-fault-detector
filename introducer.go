package main

import (
	"encoding/json"
	"fmt"
	"net"
)

func IntroduceYourself() ([]string, *net.Conn, error) {
	fmt.Println("Introducing")
	conn, err := net.Dial("udp", GetServerEndpoint(INTRODUCER_SERVER_HOST))
	if err != nil {
		return nil, nil, err
	}

	// TODO Node could create its own ID and pass it along in the Data.
	joinMessage := Message{Kind: JOIN, Data: ""}

	pingMessageEnc, err := GetEncodedPingMessage(Messages{joinMessage})
	if err != nil {
		return nil, nil, err
	}

	conn.Write(pingMessageEnc)

	buffer := make([]byte, 1024)
	// fmt.Println("%s waiting for a response")
	mLen, err := conn.Read(buffer)
	if err != nil {
		return nil, nil, err
	}

	members, err := parseMembersFromResponse(buffer[:mLen])
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Received members: ", members)

	return members, &conn, nil
}

func parseMembersFromResponse(buffer []byte) ([]string, error) {
	// fmt.Println("JOIN Response: ", response)

	messages, err := GetDecodedSubMessages(buffer)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, err
	}

	membersEnc := messages[0].Data

	var members []string
	err = json.Unmarshal([]byte(membersEnc), &members)
	if err != nil {
		return nil, err
	}

	return members, err
}

// Initalizes the Membership Information map for the newly joined node.
// Returns the NODE_ID for this node.
func InitializeMembershipInfoAndList(members []string, introducer_conn *net.Conn, localIP string) string {
	nodeId := ""

	for _, id := range members {
		ip := GetIPFromID(id)

		if ip == INTRODUCER_SERVER_HOST {
			// TODO this should be indexed thread-safe.
			membershipInfo[id] = MemberInfo{
				connection: introducer_conn,
				host:       id,
				failed:     false,
			}
		} else if ip == localIP {
			nodeId = id
		} else {
			conn, err := net.Dial("udp", GetServerEndpoint(ip))

			if err != nil {
				fmt.Println("Failed to estabilish connection with: ", id)
				// TODO what to do here? If it actually failed it should be detected by some other node.
			}

			// TODO this should be indexed thread-safe.
			membershipInfo[id] = MemberInfo{
				connection: &conn,
				host:       id,
				failed:     false,
			}
		}
	}

	return nodeId
}

// Add nodes to membership list and returns a message containing all members.
func IntroduceNodeToGroup(request string, addr *net.UDPAddr) (Message, error) {
	// TODO Add corner case checking, what if the introducer gets a looped around message from
	// the past? It should check that the node doesn't already exist.

	fmt.Println("Join message body: ", request)

	ipAddr := addr.IP.String()
	nodeId := ConstructNodeID(ipAddr)

	fmt.Printf("IP: %s NodeID: %s", ipAddr, nodeId)

	AddNewMemberToMembershipInfo(nodeId)

	members := GetMembers()

	// For the response, add yourself to the list as well.
	membershipListResponse := append(members, NODE_ID)

	membershipListEnc, err := json.Marshal(membershipListResponse)
	if err != nil {
		return Message{}, err
	}

	// TODO is it okay for the kind of this message to be "JOIN"?
	response := Message{
		Kind: JOIN,
		Data: string(membershipListEnc),
	}

	return response, err
}
