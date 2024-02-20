package tcp

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

func FindLowestIPAddress(living_IPs <-chan []string, result chan<- string) {
	var lowestIP net.IP
	var lowestIPStr string // Keep track of the lowest IP string with port

	for ips := range living_IPs {
		currentLowestIP := net.IP(nil)
		currentLowestIPStr := ""

		// Assume no lowest IP at the start of each new list of IPs
		for _, ipStr := range ips {
			parts := strings.Split(ipStr, ":")
			ip := net.ParseIP(parts[0])
			if ip == nil {
				fmt.Printf("Invalid IP address format: %s\n", ipStr)
				continue
			}
			ip = ip.To4() // Ensure we're working with IPv4
			if ip == nil {
				fmt.Printf("Skipping non-IPv4 address: %s\n", ipStr)
				continue
			}
			if currentLowestIP == nil || bytes.Compare(ip, currentLowestIP) < 0 {
				currentLowestIP = ip
				currentLowestIPStr = ipStr // This includes the port part
			}
		}

		// Check if we found a new lowest IP in the current list
		if currentLowestIP != nil && (lowestIP == nil || bytes.Compare(currentLowestIP, lowestIP) < 0 || !ipInList(lowestIP, ips)) {
			lowestIP = currentLowestIP
			lowestIPStr = currentLowestIPStr
			result <- lowestIPStr // Update the result channel with the new lowest IP
		}
	}

	if lowestIP != nil {
		fmt.Println("Final Lowest IP Address:", lowestIPStr)
	} else {
		fmt.Println("No valid IPv4 addresses found.")
		close(result) // No more values will be sent
	}
}

// ipInList checks if the given IP is present in the list of IP addresses.
func ipInList(ip net.IP, ips []string) bool {
	for _, ipStr := range ips {
		parts := strings.Split(ipStr, ":")
		listIP := net.ParseIP(parts[0]).To4()
		if listIP != nil && ip.Equal(listIP) {
			return true
		}
	}
	return false
}

func getPrimaryIP() (string, error) {
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

// Checks if the current node's IP is the lowest among the known active IPs.
func isLowestIP(ip string) bool {
	myIP, err := getPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return false
	}
	return myIP == ip
}
