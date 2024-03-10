package tcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"project/elevData"
	"project/utility"
	"project/variable"
	"reflect"
	"strings"
	"time"
)

var ServerConnection net.Conn
var ServerError error
var ShouldReconnect bool
var error_buffer = 3

func connectToServer(serverIP string, pointerElevator *elevData.Elevator, masterElevator *elevData.MasterList) {
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

	// Start a goroutine to listen for messages from the server
	go func() {
		for {
			buffer := make([]byte, 1024)            // Create a buffer to store incoming data
			n, err := ServerConnection.Read(buffer) // Read data into buffer

			if err != nil {
				// Handle error or EOF
				return // Exit goroutine if connection is closed or an error occurs
			}

			messages := strings.Split(string(buffer[:n]), "\n") // Process each newline-separated message
			for _, message := range messages {
				if message == "" {
					continue // Skip empty messages
				}

				// Determine the struct type and unmarshal based on JSON content
				genericMessage, err := utility.DetermineStructTypeAndUnmarshal([]byte(message))
				if err != nil {
					fmt.Printf("Error determining struct type or unmarshaling message: %v\n", err)
					continue
				}

				// Now, handle the unmarshaled data based on its type
				switch msg := genericMessage.(type) {
				case elevData.MasterList:
					fmt.Println("Received MasterList message")
					// Process MasterList message
					*masterElevator = msg
					jsonData := utility.MarshalJson(msg)
					SendMessage(ServerConnection, jsonData, reflect.TypeOf(msg))
				case elevData.Elevator:
					fmt.Println("Received Elevator message")
					// Process Elevator message
				case elevData.ElevStatus:
					fmt.Println("Received ElevStatus message")
				default:
					fmt.Println("Received an unknown type of message")
				}

				variable.UpdateLocal = true // Assuming this triggers some update logic
			}
		}
	}()

	ticker = time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Run a separate goroutine to listen for the exit command
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if scanner.Text() == "exit" {
				fmt.Println("Disconnecting from server...")
				ticker.Stop() // Stop the ticker to exit the loop
				break
			}
		}
	}()

	for {
		if ShouldReconnect {
			break
		}
	}

	connected = false
}

func SendMessage(conn net.Conn, message []byte, responseType reflect.Type) error {
	message = append(message, '\n')
	for {
		_, err := conn.Write(message)
		if err != nil {
			fmt.Printf("Error sending message: %s\n", err)
			if error_buffer == 0 {
				fmt.Println("Too many consecutive errors, stopping...")
				ShouldReconnect = true
				return err // Stop if there are too many consecutive errors
			} else {
				error_buffer--
			}
		} else {
			error_buffer = 3 // Reset the error buffer on successful send
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if responseType.String() == "elevData.MasterList" {
		return nil
	}

	ShouldReconnect = false
	return nil
}
