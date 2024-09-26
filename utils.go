package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid local IP address found")
}

func GetIPFromID(id string) string {
	parts := strings.Split(id, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func ConstructNodeID(ip string) string {
	return fmt.Sprintf("%s@%s", ip, time.Now().Format(time.RFC3339))
}

func GetServerEndpoint(host string) string {
	return fmt.Sprintf("%s:%d", host, SERVER_PORT)
}

// Take a list of messages and encodes them into a PING.
func GetEncodedPingMessage(messages Messages) ([]byte, error) {
	messagesEnc, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	pingMessage := Message{Kind: PING, Data: string(messagesEnc)}

	pingMessageEnc, err := json.Marshal(pingMessage)
	if err != nil {
		return nil, err
	}

	return pingMessageEnc, nil
}

func GetEncodedAckMessage(messages Messages) ([]byte, error) {
	messagesEnc, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	ackMessage := Message{Kind: PING, Data: string(messagesEnc)}

	ackMessageEnc, err := json.Marshal(ackMessage)
	if err != nil {
		return nil, err
	}

	return ackMessageEnc, nil
}

func AddToPiggybacks(message Message, ttl int) {
	// TODO thread safety
	piggybacks = append(piggybacks, PiggbackMessage{message, ttl})
}

// For a given message, returns the sub-messages present inside it.
func GetDecodedSubMessages(messageEnc []byte) (Messages, error) {
	var message Message

	err := json.Unmarshal(messageEnc, &message)
	if err != nil {
		return nil, err
	}

	var messages Messages

	err = json.Unmarshal([]byte(message.Data), &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func PrintMembershipInfo() {
	fmt.Println("====Membership Info===")
	for k, v := range membershipInfo {
		fmt.Println(k, v)
	}
}

func PrintMembershipList() {
	fmt.Println("====Membership List===")
	for _, v := range membershipList {
		fmt.Println(v)
	}
}

func PrintPiggybackMessages() {
	fmt.Println("Printing piggybacks")
	for _, p := range piggybacks {
		fmt.Println(p)
	}
}
