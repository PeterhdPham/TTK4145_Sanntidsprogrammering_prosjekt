package tcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"project/elevData"
	"time"
)

var ServerConnection net.Conn
var ServerError error
var ShouldReconnect bool
var error_buffer = 3

func connectToServer(serverIP string, pointerElevator *elevData.Elevator) {

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

	jsonData, err := json.Marshal(pointerElevator)
	if err != nil {
		fmt.Printf("Error occurred during marshaling: %v", err)
	}

	// Send jsonData using SendMessage
	SendMessage(ServerConnection, []byte(jsonData))
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

			// Convert the bytes read into a string and print it
			message := string(buffer[:n])
			fmt.Printf("Message from server: %s\n", message)

			var masterList elevData.MasterList

			err = json.Unmarshal(buffer, &masterList)
			if err != nil {
				fmt.Printf("Error occurred during unmarshaling: %v", err)
			}
			
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
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

func SendMessage(conn net.Conn, message []byte) error {
	fmt.Println("Sending message: ", string(message))
	// Ensure the message ends with a newline character, which may be needed depending on the server's reading logic.
	if !bytes.HasSuffix(message, []byte("\n")) {
		message = append(message, '\n')
	}
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
	ShouldReconnect = false
	return nil
}
