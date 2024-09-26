package main

import "net"

const (
	PING  MessageType = 0
	ACK   MessageType = 1
	JOIN  MessageType = 2
	LEAVE MessageType = 3
	FAIL  MessageType = 4
	HELLO MessageType = 5
)

type MemberInfo struct {
	connection *net.Conn
	host       string
	failed     bool
}

type MessageType int32

type Message struct {
	Kind MessageType
	// This might be a JSON encoded string, and should be decoded based on Kind.
	Data string
}

type Messages []Message

type PiggbackMessage struct {
	message Message
	ttl     int
}

type PiggybackMessages []PiggbackMessage
