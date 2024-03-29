package utils

import (
	"fmt"
	"net"

	"github.com/google/uuid"
)

func GetIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}

	return "", fmt.Errorf("unable to determine ip address")
}

func GenerateUniqueServerId() string {
	serverId := uuid.NewString()
	return serverId
}
