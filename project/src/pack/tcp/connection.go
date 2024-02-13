package pack

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

func handleConnection(conn net.Conn) {
	fmt.Println("Connected. Type messages to send, 'exit' to quit.")
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
			conn.Close()
			return
		}
		msg = strings.TrimSpace(msg)
		fmt.Println("\nReceived:", msg)

		mutex.Lock()
		if msg != lastSentMessage {
			// Send a confirmation if the received message is different from the last sent message
			confirmation := "Message received."
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
