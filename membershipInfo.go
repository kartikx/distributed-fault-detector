// Stores functions for the map that stores membership information.
package main

import (
	"fmt"
	"net"
	"sync"
)

// TODO Rename to use capital M after merge.
var membershipInfo map[string]MemberInfo = make(map[string]MemberInfo)

var membershipInfoLock = sync.RWMutex{}

func AddNewMemberToMembershipInfo(nodeId string) error {
	ipAddr := GetIPFromID(nodeId)

	conn, err := net.Dial("udp", GetServerEndpoint(ipAddr))
	if err != nil {
		return err
	}

	membershipInfoLock.Lock()
	defer membershipInfoLock.Unlock()

	membershipInfo[nodeId] = MemberInfo{
		connection: &conn,
		host:       ipAddr,
		failed:     false,
	}

	return nil
}

// Returns the members in the group. Doesn't return failed members.
func GetMembers() []string {
	members := []string{}

	membershipInfoLock.RLock()
	defer membershipInfoLock.RUnlock()

	for k, v := range membershipInfo {
		if !v.failed {
			members = append(members, k)
		}
	}
	return members
}

func PrintMembershipInfo() {
	fmt.Println("====Membership Info===")

	membershipInfoLock.RLock()
	defer membershipInfoLock.RUnlock()

	for k, v := range membershipInfo {
		fmt.Println(k, v)
	}
}

func GetNodeConnection(nodeId string) net.Conn {
	membershipInfoLock.RLock()
	defer membershipInfoLock.RUnlock()

	conn := membershipInfo[nodeId].connection

	if conn == nil {
		return nil
	}

	return *conn
}

func AddToMembershipInfo(nodeId string, member *MemberInfo) {
	membershipInfoLock.Lock()
	defer membershipInfoLock.Unlock()

	membershipInfo[nodeId] = *member
}

func GetMemberInfo(nodeId string) (MemberInfo, bool) {
	membershipInfoLock.RLock()
	defer membershipInfoLock.RUnlock()

	member, ok := membershipInfo[nodeId]

	return member, ok
}
