package main

import "net"

// const (
// 	JOIN  MessageType = 0
// 	LEAVE MessageType = 1
// 	PING  MessageType = 2
// )

type MemberInfo struct {
	connection net.Conn
	server     string
	port       string
}

// type MessageType int32

type Message struct {
	Kind string
	Data string
}
