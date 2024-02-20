package tcp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	// Map to store connections
	connections = make(map[string]net.Conn)
	// Mutex to protect access to the connections map
	connMutex = &sync.Mutex{}

	lastMsg string
	mutex   sync.Mutex // Used to synchronize access to lastMsg
)

func SendMessage(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

// Example function to broadcast a message to all connections
func BroadcastMessage(message string) {
	connMutex.Lock()
	defer connMutex.Unlock()

	for _, conn := range connections {
		_, err := conn.Write([]byte(message + "\n"))
		if err != nil {
			fmt.Println("Error broadcasting message to", conn.RemoteAddr(), ":", err)
			continue
		}
	}
	fmt.Println("Broadcasted message to all connections")
}

func ConnectionMessage(conn net.Conn) {
	for key := range connections {
		BroadcastMessage("Port-connected&" + key)
	}
}

func SendToAll() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter message: ")
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)

		mutex.Lock()
		lastMsg = msg
		mutex.Unlock()

		BroadcastMessage(msg)
	}
}

func ServerListening(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from connection: %s\n", err)
			}
			break
		}
		message = strings.TrimSpace(message)

		fmt.Printf("\033[2J\033[H")
		fmt.Printf("Confirmation received for message: %s\n", message)
	}
}

func TCP_Server() {
	port := ":9999" // Example port
	startServer(port)
}
