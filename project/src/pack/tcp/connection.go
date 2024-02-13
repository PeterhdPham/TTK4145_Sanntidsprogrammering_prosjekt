package pack

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn) {
	fmt.Println("Connected. Type messages to send, 'exit' to quit.")
	reader := bufio.NewReader(os.Stdin)

	// Concurrently read from stdin and the connection
	go func() {
		for {
			// Read message from stdin
			fmt.Print("Enter message: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)

			// Exit if the user types "exit"
			if msg == "exit" {
				fmt.Println("Exiting...")
				conn.Close()
				os.Exit(0)
			}

			// Send message
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

		// Send back a confirmation
		confirmation := "Message received."
		_, err = conn.Write([]byte(confirmation + "\n"))
		if err != nil {
			fmt.Println("Error sending confirmation:", err)
			return
		}
	}
}
