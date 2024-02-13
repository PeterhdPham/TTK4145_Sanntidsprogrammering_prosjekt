package pack

import (
	"fmt"
	"net"
	"os"
	"time"
)

func Broadcast_life() {

	fmt.Println("1")

	// Get the local network interface IP address
	iface, err := net.InterfaceByName("Wi-Fi") // Replace "eth0" with your network interface name
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("2")

	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("3")

	// Assuming the first address is the one you want to use
	localAddr, ok := addrs[0].(*net.IPNet)
	if !ok {
		fmt.Println("Error getting IP address")
		os.Exit(1)
	}

	fmt.Println("4")

	// Create a UDP address for broadcasting using the local network IP
	broadcastAddr := fmt.Sprintf("[%s]:9999", localAddr.IP.String())

	fmt.Println(broadcastAddr)

	conn, err := net.Dial("udp", broadcastAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("5")

	// Create a ticker that ticks every 5 seconds
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			// Construct the message with timestamp and sender's IP
			message := fmt.Sprintf("Hello, UDP world! Time: %s, Sender IP: %s", t.Format(time.RFC3339), localAddr.IP.String())
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Broadcast message sent: %s\n", message)
		}
	}
}

func Look_for_life() {

	living_IPs := make([]net.Addr, 0)

	// Define the UDP port on which to listen for messages.
	port := ":9999"

	// Create a UDP socket and listen on the port.
	pc, err := net.ListenPacket("udp6", port) // 'udp6' to listen on IPv6, use 'udp4' to force IPv4, or 'udp' for both
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pc.Close()

	// Create a buffer to store received messages.
	buffer := make([]byte, 2048)

	fmt.Printf("Listening for UDP packets on %s...\n", port)
	for {
		// Read from the UDP socket.
		n, addr, err := pc.ReadFrom(buffer)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if containsAddr(living_IPs, addr) {
			fmt.Print("Adress already in list\n")
		} else {
			living_IPs = append(living_IPs, addr)
			fmt.Print("Adress added to list\n")
		}

		// Handle the received message.
		fmt.Printf("Received message from %s: %s\n", addr.String(), string(buffer[:n]))
	}
}

func containsAddr(slice []net.Addr, addr net.Addr) bool {
	for _, a := range slice {
		// Simple comparison; might need to be more complex depending on your needs
		if a.Network() == addr.Network() && a.String() == addr.String() {
			return true
		}
	}
	return false
}
