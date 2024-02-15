package pack

import (
	"fmt"
	"net"
	"os"
	"time"
)

func Broadcast_life() {

	broadcastAddr := "255.255.255.255:9999"

	fmt.Println(broadcastAddr)

	// Dial the UDP connection using the IPv4 broadcast address
	conn, err := net.Dial("udp4", broadcastAddr) // "udp4" to explicitly use IPv4
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("5")

	// Create a ticker that ticks every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Construct the message with timestamp and sender's IP
		message := "Hello" // Simplified message for demonstration
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func Look_for_life(receiver chan<- []string) {

	IP_lifetimes := make(map[string]time.Time, 0)

	// Define the UDP port on which to listen for messages.
	port := ":9999"

	// Create a UDP socket and listen on the port.
	pc, err := net.ListenPacket("udp", port) // 'udp' listens for both udp4 and udp6 connections
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pc.Close()

	// Create a buffer to store received messages.
	buffer := make([]byte, 2048)

	fmt.Printf("Listening for UDP packets on %s...\n", port)
	for {

		err := pc.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			fmt.Println("Failed to set a deadline for the read operation:", err)
			os.Exit(1)
		}

		// Read from the UDP socket.
		_, addr, err := pc.ReadFrom(buffer)

		if err != nil {
			if os.IsTimeout(err) {
				fmt.Println("Read timeout: No messages received for 5 seconds\nAll other nodes assumed dead")
				IP_lifetimes = update_living_IPs(IP_lifetimes, addr)
				receiver <- get_living_IPs(IP_lifetimes)
				continue
			} else {
				fmt.Println("Read error:", err)
				continue
			}
		} else {
			// Handle the received message.
			// fmt.Println("Received message")
			IP_lifetimes = update_living_IPs(IP_lifetimes, addr)
			receiver <- get_living_IPs(IP_lifetimes)
		}
	}
}

func update_living_IPs(IP_lifetimes map[string]time.Time, new_addr net.Addr) map[string]time.Time {

	if new_addr == nil {
		for addr_in_list := range IP_lifetimes {
			IP_lifetimes[addr_in_list] = time.Now()
		}
	} else {
		_, ok := IP_lifetimes[new_addr.String()]
		if !ok {
			fmt.Println("New node discovered: ", new_addr.String())
		}
		IP_lifetimes[new_addr.String()] = time.Now().Add(5 * time.Second)
	}
	return IP_lifetimes
}

func get_living_IPs(m map[string]time.Time) []string {
	living_IPs := []string{}
	for address, death := range m {
		if death.After(time.Now()) {
			living_IPs = append(living_IPs, address)
		}
	}
	return living_IPs
}
