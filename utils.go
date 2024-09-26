package main

import (
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
