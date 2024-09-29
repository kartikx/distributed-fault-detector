// Contains functionality for accessing and updating piggyback messages.
package main

import (
	"fmt"
	"sync"
)

var piggybacks PiggybackMessages

var piggybacksLock = sync.RWMutex{}

func PrintPiggybackMessages() {
	piggybacksLock.RLock()
	defer piggybacksLock.RUnlock()

	for _, p := range piggybacks {
		fmt.Println(p)
	}
}

func AddPiggybackMessage(message Message) {
	piggybacksLock.Lock()
	defer piggybacksLock.Unlock()

	piggybacks = append(piggybacks, PiggbackMessage{message, PIGGYBACK_TTL})
}

// Returns messages from Piggyback that aren't expired.
func GetUnexpiredPiggybackMessages() Messages {
	var messages Messages

	piggybacksLock.Lock()
	defer piggybacksLock.Unlock()

	for index := 0; index < len(piggybacks); index++ {
		if piggybacks[index].ttl > 0 {
			messages = append(messages, piggybacks[index].message)
			piggybacks[index].ttl--
		}

		if piggybacks[index].ttl <= 0 {
			piggybacks = append(piggybacks[:index], piggybacks[index+1:]...)
			index--
		}
	}

	return messages
}
