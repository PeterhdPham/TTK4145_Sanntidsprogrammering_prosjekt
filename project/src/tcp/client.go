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
			fmt.Println("Debugging 1")

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

			var incomingMasterElevator elevData.MasterList
			fmt.Println("Debugging 2")
			responseType := utility.UnmarshalJson(buffer[:n], &incomingMasterElevator)
			fmt.Println("Debugging 3")

			// Serialize masterElevator to JSON
			jsonData := utility.MarshalJson(&incomingMasterElevator)
			fmt.Println("Debugging 4")

			// Send jsonData back to the primary
			err = SendMessage(ServerConnection, jsonData, responseType)
			if err != nil {
				fmt.Printf("Error sending updated masterElevator: %v\n", err)
			} else {
				fmt.Println("Updated and sent masterElevator")
				*masterElevator = incomingMasterElevator
			}
			fmt.Println("Debugging 5")

			UpdateLocal = true
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

	// // Read the response from the server
	// response := make([]byte, 1024)
	// _, err := conn.Read(response)
	// if err != nil {
	// 	fmt.Printf("Error reading response: %s\n", err)
	// 	return err
	// }

	// // Compare the response with the message that was sent
	// if !bytes.Equal(message, response) {
	// 	fmt.Println("Server did not receive the correct message")
	// 	return errors.New("server did not receive the correct message")
	// }

	ShouldReconnect = false
	return nil
}
