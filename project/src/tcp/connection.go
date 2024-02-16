package tcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var addresses = make(map[string]string)
var myAddress string

func GetMyAdress(msg string) {
	if myAddress == "" {
		parts := strings.Split(msg, "&")
		if parts[0] == "IP-address" {
			// Keep the part after "&"
			myAddress = parts[1]
		}
	}
}

func AddressFinders(msg string) {
	parts := strings.Split(msg, "&")
	if parts[0] == "Port-connected" {
		// Keep the part after "&"
		afterAmpersand := parts[1]
		_, ok := addresses[afterAmpersand]
		if !ok {
			addresses[afterAmpersand] = strings.Replace(strings.Replace(afterAmpersand, ":", "", -1), ".", "", -1)
		}
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("Connected...")
	reader := bufio.NewReader(os.Stdin)

	var lastSentMessage string
	var mutex sync.Mutex // Used to synchronize access to lastSentMessage

	// Concurrently read from stdin and send messages
	go func() {
		for {
			fmt.Print("Enter message: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)

			if msg == "exit" {
				fmt.Println("Exiting...")
				conn.Close()
				os.Exit(0)
			}

			mutex.Lock()
			lastSentMessage = msg // Update the last sent message before sending
			mutex.Unlock()

			_, err := conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Println("Error sending message:", err)
				continue
			}
		}
	}()

	// Listen for messages from the connection
	for {
		netReader := bufio.NewReader(conn)
		msg, err := netReader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from peer.")

			if LowestIPAddress() == myAddress {
				fmt.Println("I'm backup server")
				fmt.Println(myAddress)
				masterIPAddress = myAddress
				startServer(myAddress)
			} else {
				fmt.Println("I'm client")
				startClient(LowestIPAddress())
			}
			conn.Close()
			return
		}
		msg = strings.TrimSpace(msg)
		fmt.Println("\nReceived:", msg)

		AddressFinders(msg)
		GetMyAdress(msg)

		mutex.Lock()
		if msg != lastSentMessage {
			// Send a confirmation if the received message is different from the last sent message
			confirmation := msg
			_, err = conn.Write([]byte(confirmation + "\n"))
			if err != nil {
				fmt.Println("Error sending confirmation:", err)
				mutex.Unlock() // Ensure mutex is unlocked before returning
				return
			}
		}
		mutex.Unlock() // Ensure mutex is unlocked after handling the message
	}
}

func Listening(conn net.Conn, lastMsg string) {
	lastSentMessage := lastMsg
	var mutex sync.Mutex // Used to synchronize access to lastSentMessage

	// Listen for messages from the connections
	for {
		netReader := bufio.NewReader(conn)
		msg, err := netReader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from peer.")
			conn.Close()
			return
		}
		msg = strings.TrimSpace(msg)
		fmt.Println("\nReceived:", msg)

		mutex.Lock()
		if msg != lastSentMessage {
			// Send a confirmation if the received message is different from the last sent message
			confirmation := msg
			_, err = conn.Write([]byte(confirmation + "\n"))
			if err != nil {
				fmt.Println("Error sending confirmation:", err)
				mutex.Unlock() // Ensure mutex is unlocked before returning
				return
			}
		}
		mutex.Unlock() // Ensure mutex is unlocked after handling the message
	}
}

func LowestIPAddress() string {
	var lowestIP string
	var lowestIP_value string
	for address, address_value := range addresses {
		if lowestIP == "" {
			lowestIP = address
			lowestIP_value = address_value
		} else if address_value < lowestIP_value {
			lowestIP = address
			lowestIP_value = address_value
		}
	}
	return lowestIP
}
