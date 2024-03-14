package tcp

import (
	"fmt"
	"net"
	"project/communication"
	"project/defs"
	"project/utility"
	"strings"
	"time"
)

var ServerConnection net.Conn
var ServerError error
var ShouldReconnect bool
var UpdateLocal bool = false

func connectToServer(serverIP string, pointerElevator *defs.Elevator, masterElevator *defs.MasterList) {
	serverAddr := serverIP
	ServerConnection, ServerError = net.Dial("tcp", serverAddr)
	if ServerError != nil {
		fmt.Printf("Failed to connect to server: %s\n", ServerError)
		connected = false
		return
	}
	defer ServerConnection.Close()
	fmt.Println("Connected to server at", serverAddr)
	connected = true
	ShouldReconnect = false

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	communication.SendMessage(ServerConnection, *pointerElevator, "init")

	// Sends previous master list
	communication.SendMessage(ServerConnection, *masterElevator, "prev")

	// Start a goroutine to listen for messages from the server
	go func() {
		for {
			buffer := make([]byte, 8192)            // Create a buffer to store incoming data
			n, err := ServerConnection.Read(buffer) // Read data into buffer

			if err != nil {
				// Handle error or EOF
				return // Exit goroutine if connection is closed or an error occurs
			}

			messages := strings.Split(string(buffer[:n]), "%") // Process each newline-separated message
			for _, message := range messages {
				if message == "" || message == " " || !strings.HasSuffix(message, "}]}") || !strings.HasPrefix(message, `{"elevators":`) {
					continue // Skip empty messages
				}
				fmt.Println("Message: ", message)

				// Determine the struct type and unmarshal based on JSON content
				genericMessage, err := utility.DetermineStructTypeAndUnmarshal([]byte(message))
				if err != nil {
					fmt.Printf("Error determining struct type or unmarshaling message: %v\n", err)
					continue
				}

				// Now, handle the unmarshaled data based on its type
				switch msg := genericMessage.(type) {
				case defs.MasterList:
					// Process MasterList message
					*masterElevator = msg
					communication.SendMessage(ServerConnection, msg, "")
					defs.UpdateLocal <- "true" // Assuming this triggers some update logic
				case defs.Elevator:
					fmt.Println("Received Elevator message")
					// Process Elevator message
				case defs.ElevStatus:
					fmt.Println("Received ElevStatus message")
				default:
					fmt.Println("Received an unknown type of message")
				}

			}
		}
	}()

	for {
		if ShouldReconnect {
			break
		}
	}

	connected = false
	fmt.Println("Shutting down client connection...")
	ServerConnection.Close() // Explicitly close the connection
}
