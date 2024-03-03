package tcp

import (
	"net"
)

func GetPrimaryIP() (string, error) {
	var primaryIP string
	conn, err := net.Dial("udp", "8.8.8.8:80") // Using an external server to determine the preferred outbound IP.
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	primaryIP = localAddr.IP.String()
	return primaryIP, nil
}
