package tcp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"project/elevData"
	"project/utility"
	"reflect"
	"strings"
	"time"
)

var ServerConnection net.Conn
var ServerError error
var ShouldReconnect bool
var error_buffer = 3
var UpdateLocal bool = false

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
				if err == io.EOF {
					fmt.Println("Server closed the connection.")
				} else {
					fmt.Printf("Error reading from server: %s\n", err)
				}
				connected = false
				ServerConnection.Close()
				return // Exit goroutine if connection is closed or an error occurs
			}

			// Process each newline-separated message
			messages := strings.Split(string(buffer[:n]), "\n")
			for _, message := range messages {
				if message == "" {
					continue // Skip empty messages
				}

				// Attempt to unmarshal the message into a generic interface
				var genericMessage interface{}
				responseType, err := utility.UnmarshalJson([]byte(message), &genericMessage)
				if err != nil {
					fmt.Printf("Error unmarshaling message: %v\n", err)
					continue // Skip to the next message
				}

				// Use type switch to handle different possible struct types
				switch responseType.String() {
				case "elevData.MasterList":
					fmt.Println("Received MasterList message")
					// Process MasterList message
					// // *masterElevator = msg
					// jsonData := utility.MarshalJson(msg)
					// SendMessage(ServerConnection, jsonData, reflect.TypeOf(msg))
				case "elevData.Elevator":
					fmt.Println("Received Elevator message")
					// Process Elevator message
				case "elevData.ElevStatus":
					fmt.Println("Received ElevStatus message")
					// Process ElevStatus message
				default:
					fmt.Printf("Received an unknown type of message: %v\n", responseType.String())
				}

				UpdateLocal = true
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
