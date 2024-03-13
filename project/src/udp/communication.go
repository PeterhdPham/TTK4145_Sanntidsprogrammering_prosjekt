package udp

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
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

	var errCount int = 0

	for range ticker.C {
		// Construct the message with timestamp and sender's IP
		message := "Please give us an A on the project:)" // Simplified message for demonstration
		_, err := conn.Write([]byte(message))
		if err != nil {
			errCount++
			fmt.Println("Error sending udp-message: ", err)
			if errCount > 10 {
				fmt.Println("Too many consecutive udp errors, Restarting UDP connection")
				conn.Close()
				conn, err = net.Dial("udp4", BROADCAST_ADDR) // "udp4" to explicitly use IPv4
				if err != nil {
					fmt.Println(err)
					return
				}
				errCount = 0
			}
		}
	}
}

func LookForLife(livingIPsChan chan<- []string) {

	myIP, err := GetPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return
	}

	IPLifetimes := make(map[string]time.Time)

	// Create a UDP socket and listen on the port.
	pc, err := net.ListenPacket("udp", LISTEN_ADDR) // 'udp' listens for both udp4 and udp6 connections
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pc.Close()

	// Create a buffer to store received messages.
	buffer := make([]byte, 4096)

	for {

		err := pc.SetReadDeadline(time.Now().Add(LISTEN_TIMEOUT))
		if err != nil {
			fmt.Println("Failed to set a deadline for the read operation:", err)
		}

		// Read from the UDP socket.
		_, addr, err := pc.ReadFrom(buffer)

		if err != nil {
			if os.IsTimeout(err) {
				IPLifetimes = updateLivingIPs(IPLifetimes, addr, myIP)
				livingIPsChan <- getLivingIPs(IPLifetimes)
				continue
			} else {
				fmt.Println("Read error:", err)
				continue
			}
		} else {
			// Handle the received message.
			IPLifetimes = updateLivingIPs(IPLifetimes, addr, myIP)
			livingIPsChan <- getLivingIPs(IPLifetimes)
		}
	}
}

func updateLivingIPs(IPLifetimes map[string]time.Time, newAddr net.Addr, myIP string) map[string]time.Time {

	if newAddr == nil {
		for addrInList := range IPLifetimes {
			IPLifetimes[addrInList] = time.Now()
		}
	} else {
		_, ok := IPLifetimes[newAddr.String()]
		if !ok {
			if strings.Split(newAddr.String(), ":")[0] != myIP {
				fmt.Println("New node discovered: ", newAddr.String())
			} else {
				fmt.Println("This is my IP: ", myIP)
			}
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
	}
	livingIPs = ipSorter(livingIPs)
	return livingIPs
}

func ipSorter(ipStrings []string) []string {
	ipMap := make(map[string]int)
	var ipStringsNew []string
	var ipIntsNew []int
	for _, ipStr := range ipStrings {
		ipInt, _ := strconv.Atoi(strings.Replace(ipStr, ".", "", -1))
		ipMap[ipStr] = ipInt
		ipIntsNew = append(ipIntsNew, ipInt)
	}
	sort.Ints(ipIntsNew)
	for _, ipInt := range ipIntsNew {
		for ipStr, ipInt2 := range ipMap {
			if ipInt == ipInt2 {
				ipStringsNew = append(ipStringsNew, ipStr)
			}
		}
	}
	return ipStringsNew
}
