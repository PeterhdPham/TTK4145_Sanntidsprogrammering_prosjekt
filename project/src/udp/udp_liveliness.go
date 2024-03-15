package udp

import (
	"fmt"
	"log"
	"net"
	"os"
	"project/defs"
	"sort"
	"strconv"
	"strings"
	"time"
)

const PORT = "9999" // Port used to broadcast and listen to "I'm alive"-messages

const BROADCAST_ADDR = "255.255.255.255:" + PORT // Address to broadcast "I'm alive"-msg
const BROADCAST_PERIOD = 100 * time.Millisecond  // Time to wait before broadcasting new msg
const LISTEN_ADDR = "0.0.0.0:" + PORT            // Address to listen for "I'm alive"-msg
const LISTEN_TIMEOUT = 10 * time.Second          // Time to listen before giving up
const NODE_LIFE = 5 * time.Second                // Time added to node-lifetime when msg is received
const ALLOWED_CONSECUTIVE_ERRORS = 100           // Number of allowed consecutive udp error

func BroadcastLife() {

	// Dial the UDP connection using the IPv4 broadcast address
	conn, err := net.Dial("udp4", BROADCAST_ADDR) // "udp4" to explicitly use IPv4
	if err != nil {
		log.Println(err)
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
			// log.Println("Error sending udp-message: ", err)
			if errCount > ALLOWED_CONSECUTIVE_ERRORS {
				log.Println("Too many consecutive udp errors, Restarting UDP connection")
				conn.Close()
				conn, err = net.Dial("udp4", BROADCAST_ADDR) // "udp4" to explicitly use IPv4
				if err != nil {
					log.Println(err)
					return
				}
				errCount = 0
			}
		}
	}
}

func LookForLife(livingIPsChan chan<- []string) {

	myIP := defs.MyIP

	IPLifetimes := make(map[string]time.Time)
	IPLifetimes[myIP] = time.Now().Add(time.Hour)

	// Create a UDP socket and listen on the port.
	pc, err := net.ListenPacket("udp", LISTEN_ADDR) // 'udp' listens for both udp4 and udp6 connections
	if err != nil {
		log.Println(err)
		return
	}
	defer pc.Close()

	// Create a buffer to store received messages.
	buffer := make([]byte, 8192)

	for {
		if defs.MyIP != myIP {
			IPLifetimes[myIP] = time.Now()
			myIP = defs.MyIP
			IPLifetimes[myIP] = time.Now().Add(time.Hour)
		}

		err := pc.SetReadDeadline(time.Now().Add(LISTEN_TIMEOUT))
		if err != nil {
			log.Println("Failed to set a deadline for the read operation:", err)
		}

		// Read from the UDP socket.
		_, addr, err := pc.ReadFrom(buffer)
		var addrString string
		if addr != nil {
			addrString = strings.Split(addr.String(), ":")[0]
		} else {
			addrString = ""
		}

		if addrString == "10.100.23.34" {
			fmt.Println("Received Message from kristian")
			fmt.Println(IPLifetimes)
		}

		if err != nil {
			if os.IsTimeout(err) {
				IPLifetimes = updateLivingIPs(IPLifetimes, "", myIP)
				livingIPsChan <- getLivingIPs(IPLifetimes)
				continue
			} else {
				log.Println("Read error:", err)
				continue
			}
		} else {
			// Handle the received message.
			IPLifetimes = updateLivingIPs(IPLifetimes, addrString, myIP)
			livingIPsChan <- getLivingIPs(IPLifetimes)
		}
	}
}

func updateLivingIPs(IPLifetimes map[string]time.Time, newAddr string, myIP string) map[string]time.Time {

	if newAddr == "" {
		for addrInList := range IPLifetimes {
			if addrInList != myIP {
				IPLifetimes[addrInList] = time.Now()
			}
		}
	} else {
		_, ok := IPLifetimes[newAddr]
		if !ok {
			if newAddr != myIP {
				log.Println("New node discovered: ", newAddr)
			} else {
				log.Println()
				log.Println("This is my IP: ", myIP)
			}
			IPLifetimes[newAddr] = time.Now().Add(NODE_LIFE)
		} else {
			IPLifetimes[newAddr] = IPLifetimes[newAddr].Add(NODE_LIFE)
		}
	}
	return IPLifetimes
}

func getLivingIPs(m map[string]time.Time) []string {
	livingIPs := []string{}
	for address, death := range m {
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
