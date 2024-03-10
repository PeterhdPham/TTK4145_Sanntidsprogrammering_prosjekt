package tcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"project/broadcast"
	"project/elevData"
	"project/udp"
	"project/utility"
	"project/variable"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	LivingIPsChan  = make(chan []string)         //Stores living IPs from the Look_for_life function
	ActiveIPsMutex sync.Mutex                    //Mutex for protecting active IPs
	ActiveIPs      []string                      //List of active IPs
	connected      bool                  = false //Client connection state

	WaitingForConfirmation bool //
)

func Config_Roles(pointerElevator *elevData.Elevator, masterElevator *elevData.MasterList) {
	//Go routines for finding active IPs
	go udp.BroadcastLife()
	go udp.LookForLife(LivingIPsChan)

	// Initialize a ticker that ticks every 1 seconds.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case livingIPs := <-LivingIPsChan:
			// Update the list of active IPs whenever a new list is received.
			if !utility.SlicesAreEqual(ActiveIPs, livingIPs) {
				ActiveIPsMutex.Lock()
				ActiveIPs = livingIPs
				fmt.Println("Active IPs updated")
				ActiveIPsMutex.Unlock()
				updateRole(pointerElevator, masterElevator)
			}
			// fmt.Print("Active IPs: ", ActiveIPs, "\n", "Living IP: ", livingIPs, "\n")

		case <-ticker.C:
			// Every 1 seconds, check the roles and updates if necessary.
			// updateRole(pointerElevator, masterElevator)
		}
	}
}
func updateRole(pointerElevator *elevData.Elevator, masterElevator *elevData.MasterList) {
	ActiveIPsMutex.Lock()
	defer ActiveIPsMutex.Unlock()

	//Sets the role to master if there is not active IPs (Internet turned off while running)
	if len(ActiveIPs) == 0 {
		fmt.Println("No active IPs found. Waiting for discovery...")
		pointerElevator.Role = elevData.Master
		return
	}

	//Find the IP for the current computer
	MyIP, err := udp.GetPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return
	}
	//Finds the lowestIP and sets the ServerIP equal to it
	lowestIP := strings.Split(ActiveIPs[0], ":")[0]
	if variable.ServerIP != lowestIP {
		connected = false
		variable.ServerIP = lowestIP
	}
	//Sets role to master if lowestIP is localhost
	if lowestIP == "127.0.0.1" {
		fmt.Println("Running on localhost")
		pointerElevator.Role = elevData.Master
		return
	}

	if MyIP == lowestIP && !variable.ServerListening {
		//Set role to master and starts a new server on
		shutdownServer()
		fmt.Println("This node is the server.")
		// port := strings.Split(ActiveIPs[0], ":")[1]
		go startServer(masterElevator) // Ensure server starts in a non-blocking manner
		pointerElevator.Role = elevData.Master
	} else if MyIP != lowestIP && variable.ServerListening {
		//Stops the server and switches from master to slave role
		fmt.Println("This node is no longer the server, transitioning to client...")
		shutdownServer()                                                       // Stop the server
		go connectToServer(lowestIP+":55555", pointerElevator, masterElevator) // Transition to client
		pointerElevator.Role = elevData.Slave
	} else if !variable.ServerListening {
		//Starts a client connection to the server, and sets role to slave
		if !connected {
			fmt.Println("This node is a client.")
			go connectToServer(lowestIP+":55555", pointerElevator, masterElevator)
			pointerElevator.Role = elevData.Slave
		}
	}
}

func startServer(masterElevator *elevData.MasterList) {
	// Initialize the map to track client connections at the correct scope
	variable.ClientConnections = make(map[net.Conn]bool)
	_, err := udp.GetPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return
	}
	// Check if the server is already running, and if so, initiate shutdown for role switch
	if variable.ServerListening {
		fmt.Println("Server is already running, attempting to shut down for role switch...")
		variable.ServerCancel()     // Request server shutdown
		time.Sleep(1 * time.Second) // Give it a moment to shut down before restarting
	}

	// Create a new context for this server instance
	var ctx context.Context
	ctx, variable.ServerCancel = context.WithCancel(context.Background())
	variable.ServerListening = true

	listenAddr := "0.0.0.0:55555"
	fmt.Println("Starting server at: " + listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		variable.ServerListening = false // Ensure the state reflects that the server didn't start
		return
	}
	defer func() {
		listener.Close()
		fmt.Println("Server listener closed.")
	}()
	fmt.Println("Server listening on", listenAddr)

	// Accept new connections unless server shutdown is requested
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done(): // Shutdown was requested
					fmt.Println("Server shutting down...")
					closeAllClientConnections() // Ensure all client connections are gracefully closed
					variable.ServerListening = false
					return
				default:
					fmt.Printf("Failed to accept connection: %s\n", err)
					continue
				}
			}
			go handleConnection(conn, masterElevator)
		}
	}()

	// Wait for the shutdown signal to clean up and exit the function
	<-ctx.Done()
	fmt.Println("Server shutdown completed.")
}

// Ensure this function exists and is correctly implemented to close all client connections
func closeAllClientConnections() {
	variable.ClientMutex.Lock()
	defer variable.ClientMutex.Unlock()

	for conn := range variable.ClientConnections {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
		}
		delete(variable.ClientConnections, conn)
	}
}

// Implement or adjust compareMasterLists to be compatible with the above modifications
func CompareMasterLists(list1, list2 []byte) bool {
	return bytes.Equal(list1, list2)
}

// Handles individual client connections.
func handleConnection(conn net.Conn, masterElevator *elevData.MasterList) {
	variable.ClientMutex.Lock()
	variable.ClientConnections[conn] = true
	variable.ClientMutex.Unlock()

	defer func() {
		conn.Close()
		variable.ClientMutex.Lock()
		delete(variable.ClientConnections, conn)
		variable.ClientMutex.Unlock()
	}()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", clientAddr)

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client %s disconnected gracefully.\n", clientAddr)
			} else {
				fmt.Printf("Error reading from client %s: %s\n", clientAddr, err)
			}
			break
		}

		// Process each newline-separated message
		messages := strings.Split(string(buffer[:n]), "\n")
		for _, message := range messages {
			if message == "" {
				continue // Skip empty messages
			}

			// Attempt to determine the struct type from the JSON keys
			genericStruct, err := utility.DetermineStructTypeAndUnmarshal([]byte(message))
			if err != nil {
				fmt.Printf("Failed to determine struct type or unmarshal message from client %s: %s\n", clientAddr, err)
				continue
			}

			// Now handle the unmarshaled data based on its determined type
			switch v := genericStruct.(type) {
			case elevData.MasterList:
				fmt.Printf("Unmarshaled MasterList from client %s.\n", clientAddr)
				if reflect.DeepEqual(v, *masterElevator) {
					fmt.Println("Server received the correct masterList")
					jsonToSend := utility.MarshalJson(masterElevator)
					fmt.Println("Master:", string(jsonToSend))
				} else {
					fmt.Println("Server did not receive the correct confirmation")
				}
			case elevData.ElevStatus:
				fmt.Printf("Unmarshaled ElevStatus from client %s.\n", clientAddr)
				fmt.Printf("Data: %v\n", v)
				requestFloor := v.Buttonfloor
				requestButton := v.Buttontype
				// Handle ElevStatus-specific logic here
				if requestButton != -1 || requestFloor != -1 {
					fmt.Println("Button from remote")
					elevData.RemoteStatus = v
					variable.UpdateOrdersFromMessage = true
					fmt.Printf("Variable: %v\n", variable.UpdateOrdersFromMessage)
				} else {
					elevData.RemoteStatus = v
					variable.UpdateStatusFromMessage = true
					// IP IMportant
				}
				// Now, signal the messageReceived channel to run its case
				select {
				case variable.MessageReceived <- struct{}{}: // Non-blocking send, in case no receiver is ready
				default:
				}

			case elevData.Elevator:
				fmt.Printf("Unmarshaled Elevator from client %s.\n", clientAddr)
				// Handle Elevator-specific logic here
				if !utility.IsIPInMasterList(v.Ip, *masterElevator) {
					masterElevator.Elevators = append(masterElevator.Elevators, v)
				}

				jsonToSend := utility.MarshalJson(masterElevator)
				fmt.Println("Broadcasting master")
				broadcast.BroadcastMessage(nil, jsonToSend)
			default:
				fmt.Printf("Received unknown type from client %s\n", clientAddr)
			}
		}
	}
}

func shutdownServer() {
	// First, cancel the server context to stop accepting new connections
	variable.ServerCancel()

	// Next, close all active client connections
	variable.ClientMutex.Lock()
	for conn := range variable.ClientConnections {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
		}
		delete(variable.ClientConnections, conn)
	}
	variable.ClientMutex.Unlock()

	// Finally, mark the server as not listening
	variable.ServerListening = false
	fmt.Println("Server has been shut down and all connections are closed.")
}
