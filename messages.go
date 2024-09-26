package main

import "net"

const (
	PING  MessageType = 0
	JOIN  MessageType = 1
	LEAVE MessageType = 2
	ACK   MessageType = 3
)

type MemberInfo struct {
	connection net.Conn
	server     string
	port       string
}

type MessageType int32

type Message struct {
	Kind MessageType
	// This might be a JSON encoded string, and should be decoded based on Kind.
	Data string
}

type Messages []Message
