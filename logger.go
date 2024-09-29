// Implements logging functionality.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var log_file, _ = os.OpenFile("./machine.log", os.O_WRONLY|os.O_CREATE, 0666)
var log_file_writer = bufio.NewWriter(log_file)

func PrintMessage(direction string, message Message, nodeId string) {
	fmt.Fprintf(log_file_writer, "---------\nPrinting %s message\n%s\n", direction, time.Now())
	switch message.Kind {
	case PING:
		fmt.Fprintf(log_file_writer, "PING message (%s):\n", nodeId)

		var messages Messages
		err := json.Unmarshal([]byte(message.Data), &messages)
		if err != nil {
			fmt.Fprintf(log_file_writer, "Failed to unmarshal PING submessages")
			return
		}

		fmt.Fprintf(log_file_writer, "Submessages vvvvv\n")
		for _, subMessage := range messages {
			PrintMessage(direction, subMessage, nodeId)
		}
		fmt.Fprintf(log_file_writer, "Submessages ^^^^^\n")
	case ACK:
		fmt.Fprintf(log_file_writer, "ACK message (%s):\n", nodeId)

		var messages Messages
		err := json.Unmarshal([]byte(message.Data), &messages)
		if err != nil {
			fmt.Fprintf(log_file_writer, "Failed to unmarshal PING submessages")
			return
		}

		fmt.Fprintf(log_file_writer, "Submessages vvvvv\n")
		for _, subMessage := range messages {
			PrintMessage(direction, subMessage, nodeId)
		}
		fmt.Fprintf(log_file_writer, "Submessages ^^^^^\n")

	case JOIN:
		fmt.Fprintf(log_file_writer, "JOIN message with %s\n", message.Data)

	case LEAVE:
		fmt.Fprintf(log_file_writer, "LEAVE message with %s\n", message.Data)

	case FAIL:
		fmt.Fprintf(log_file_writer, "FAIL message with %s\n", message.Data)

	case HELLO:
		fmt.Fprintf(log_file_writer, "HELLO message with %s\n", message.Data)

	case SUSPECT:
		fmt.Fprintf(log_file_writer, "SUSPECT message with %s\n", message.Data)

	case ALIVE:
		fmt.Fprintf(log_file_writer, "ALIVE message with %s\n", message.Data)

	case SUSPECT_MODE:
		fmt.Fprintf(log_file_writer, "SUSPECT_MODE message with %s\n", message.Data)

	default:
		fmt.Fprintf(log_file_writer, "********Trying to print unknown message type**********")
	}
	fmt.Fprintf(log_file_writer, "---------\n")
}

func LogMessage(message string) {
	fmt.Fprintf(log_file_writer, "[%s] %s\n", time.Now().Format(time.TimeOnly), message)
}
