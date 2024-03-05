package udp

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

const PORT = "9999" // Port used to broadcast and listen to "I'm alive"-messages

const BROADCAST_ADDR = "255.255.255.255:" + PORT // Address to broadcast "I'm alive"-msg
const BROADCAST_PERIOD = 100 * time.Millisecond  // Time to wait before broadcasting new msg
const LISTEN_ADDR = "0.0.0.0:" + PORT            // Address to listen for "I'm alive"-msg
const LISTEN_TIMEOUT = 5 * time.Second           // Time to listen before giving up
const NODE_LIFE = time.Second                    // Time added to node-lifetime when msg is received

func BroadcastLife() {

	// Dial the UDP connection using the IPv4 broadcast address
	conn, err := net.Dial("udp4", BROADCAST_ADDR) // "udp4" to explicitly use IPv4
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Create a ticker that ticks every 2 seconds
	ticker := time.NewTicker(BROADCAST_PERIOD)
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

func LookForLife(livingIPsChan chan<- []string) {

	IPLifetimes := make(map[string]time.Time)

	// Create a UDP socket and listen on the port.
	pc, err := net.ListenPacket("udp", LISTEN_ADDR) // 'udp' listens for both udp4 and udp6 connections
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pc.Close()

	// Create a buffer to store received messages.
	buffer := make([]byte, 2048)

	fmt.Printf("Listening for UDP packets on %s...\n", PORT)
	for {

		err := pc.SetReadDeadline(time.Now().Add(LISTEN_TIMEOUT))
		if err != nil {
			fmt.Println("Failed to set a deadline for the read operation:", err)
			os.Exit(1)
		}

		// Read from the UDP socket.
		_, addr, err := pc.ReadFrom(buffer)

		if err != nil {
			if os.IsTimeout(err) {
				fmt.Println("Read timeout: No messages received for 5 seconds\nAll other nodes assumed dead")
				IPLifetimes = updateLivingIPs(IPLifetimes, addr)
				livingIPsChan <- getLivingIPs(IPLifetimes)
				continue
			} else {
				fmt.Println("Read error:", err)
				continue
			}
		} else {
			// Handle the received message.
			IPLifetimes = updateLivingIPs(IPLifetimes, addr)
			livingIPsChan <- getLivingIPs(IPLifetimes)
		}
	}
}

func updateLivingIPs(IPLifetimes map[string]time.Time, newAddr net.Addr) map[string]time.Time {

	if newAddr == nil {
		for addrInList := range IPLifetimes {
			IPLifetimes[addrInList] = time.Now()
		}
	} else {
		_, ok := IPLifetimes[newAddr.String()]
		if !ok {
			fmt.Println("New node discovered: ", newAddr.String())
		}
		IPLifetimes[newAddr.String()] = time.Now().Add(NODE_LIFE)
	}
	return IPLifetimes
}

func getLivingIPs(m map[string]time.Time) []string {
	livingIPs := []string{}
	for address, death := range m {
		address = strings.Split(address, ":")[0]
		if death.After(time.Now()) {
			livingIPs = append(livingIPs, address)
		}
		sort.Strings(livingIPs)
	}
	return livingIPs
}
