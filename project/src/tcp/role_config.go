package tcp

import (
	"Driver-go/elevio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"project/broadcast"
	"project/cost"
	"project/defs"
	"project/elevData"
	"project/udp"
	"project/utility"
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

func Config_Roles(pointerElevator *defs.Elevator, masterElevator *defs.MasterList) {
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
			if !slicesAreEqual(ActiveIPs, livingIPs) {
				ActiveIPsMutex.Lock()
				if pointerElevator.Ip == livingIPs[0] {
					// If I'm the master i should reassign orders of the dead node
					ReassignOrders(masterElevator, ActiveIPs, livingIPs)
					jsonToSend := utility.MarshalJson(masterElevator)
					broadcast.BroadcastMessage(nil, jsonToSend)
				}
				ActiveIPs = livingIPs
				ActiveIPsMutex.Unlock()
				updateRole(pointerElevator, masterElevator)
			}
		case <-ticker.C:
			// Every 1 seconds, check the roles and updates if necessary.
			// updateRole(pointerElevator, masterElevator)
		}
	}
}

// Used when the ActiveIPs list is changed
func ReassignOrders(masterElevator *defs.MasterList, oldList []string, newList []string) {
	fmt.Println("Reassigning orders")
	for _, elevator := range oldList {
		if !utility.Contains(newList, elevator) {
			for _, e := range masterElevator.Elevators {
				if e.Ip == elevator {
					for floorIndex, floorOrders := range e.Orders {
						if floorOrders[elevio.BT_HallUp] {
							floorOrders[elevio.BT_HallUp] = false
							cost.FindAndAssign(masterElevator, floorIndex, int(elevio.BT_HallUp), elevator)
						}
						if floorOrders[elevio.BT_HallDown] {
							floorOrders[elevio.BT_HallDown] = false
							cost.FindAndAssign(masterElevator, floorIndex, int(elevio.BT_HallDown), elevator)
						}
					}
				}
			}
		}
	}
}

// Used when elevators still are online, but one or more elevators are inoperative
func ReassignOrders2(masterList *defs.MasterList) {
	operativeElevators := make([]string, 0)
	livingElevators := make([]string, 0)

	for _, e := range masterList.Elevators {
		livingElevators = append(livingElevators, e.Ip)
		if e.Status.Operative && (e.Ip != defs.MyIP) {
			operativeElevators = append(operativeElevators, e.Ip)
		}
	}
	if (len(livingElevators) > len(operativeElevators)) && (len(operativeElevators) > 0) {
		ReassignOrders(masterList, livingElevators, operativeElevators)
	}
}

func updateRole(pointerElevator *defs.Elevator, masterElevator *defs.MasterList) {
	ActiveIPsMutex.Lock()
	defer ActiveIPsMutex.Unlock()

	//Sets the role to master if there is not active IPs (Internet turned off while running)
	if len(ActiveIPs) == 0 {
		fmt.Println("No active IPs found. Waiting for discovery...")
		pointerElevator.Role = defs.MASTER
		return
	}

	//Finds the lowestIP and sets the ServerIP equal to it
	lowestIP := strings.Split(ActiveIPs[0], ":")[0]
	if defs.ServerIP != lowestIP {
		connected = false
		defs.ServerIP = lowestIP
	}
	//Sets role to master if lowestIP is localhost
	if lowestIP == "127.0.0.1" {
		fmt.Println("Running on localhost")
		pointerElevator.Role = defs.MASTER
		return
	}

	if defs.MyIP == lowestIP && !defs.ServerListening {
		//Set role to master and starts a new server on
		shutdownServer()
		// fmt.Println("This node is the server.")
		// port := strings.Split(ActiveIPs[0], ":")[1]
		go startServer(masterElevator) // Ensure server starts in a non-blocking manner
		pointerElevator.Role = defs.MASTER
	} else if defs.MyIP != lowestIP && defs.ServerListening {
		//Stops the server and switches from master to slave role
		// fmt.Println("This node is no longer the server, transitioning to client...")
		shutdownServer()                                                       // Stop the server
		go connectToServer(lowestIP+":55555", pointerElevator, masterElevator) // Transition to client
		pointerElevator.Role = defs.SLAVE
	} else if !defs.ServerListening {
		//Starts a client connection to the server, and sets role to slave
		if !connected {
			// fmt.Println("This node is a client.")
			go connectToServer(lowestIP+":55555", pointerElevator, masterElevator)
			pointerElevator.Role = defs.SLAVE
		}
	}
}

func startServer(masterElevator *defs.MasterList) {
	// Initialize the map to track client connections at the correct scope
	defs.ClientConnections = make(map[net.Conn]bool)
	_, err := udp.GetPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return
	}
	// Check if the server is already running, and if so, initiate shutdown for role switch
	if defs.ServerListening {
		fmt.Println("Server is already running, attempting to shut down for role switch...")
		time.Sleep(1 * time.Second) // Give it a moment to shut down before restarting
	}

	// Create a new context for this server instance
	var ctx context.Context
	ctx, _ = context.WithCancel(context.Background())
	defs.ServerListening = true

	listenAddr := "0.0.0.0:55555"
	// fmt.Println("Starting server at: " + listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		defs.ServerListening = false // Ensure the state reflects that the server didn't start
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
					defs.ServerListening = false
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
	defs.ClientMutex.Lock()
	defer defs.ClientMutex.Unlock()

	for conn := range defs.ClientConnections {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
		}
		delete(defs.ClientConnections, conn)
	}
}

// Implement or adjust compareMasterLists to be compatible with the above modifications
func CompareMasterLists(list1, list2 []byte) bool {
	return bytes.Equal(list1, list2)
}

// Handles individual client connections.
func handleConnection(conn net.Conn, masterElevator *defs.MasterList) {
	defs.ClientMutex.Lock()
	defs.ClientConnections[conn] = true
	defs.ClientMutex.Unlock()

	defer func() {
		conn.Close()
		defs.ClientMutex.Lock()
		delete(defs.ClientConnections, conn)
		defs.ClientMutex.Unlock()
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
			case defs.MasterList:
				fmt.Printf("Unmarshaled MasterList from client %s.\n", clientAddr)
				if reflect.DeepEqual(v, *masterElevator) {
					fmt.Println("Server received the correct masterList")
				} else {
					fmt.Println("Server did not receive the correct confirmation")
				}
			case defs.ElevStatus:
				fmt.Printf("Unmarshaled ElevStatus from client %s.\n", clientAddr)
				requestFloor := v.Buttonfloor
				requestButton := v.Buttontype
				// Handle ElevStatus-specific logic here
				if requestButton != -1 || requestFloor != -1 {
					defs.RemoteStatus = v
					defs.ButtonReceived <- defs.ButtonEventWithIP{
						Event: elevio.ButtonEvent{Floor: v.Buttonfloor, Button: elevio.ButtonType(v.Buttontype)},
						IP:    strings.Split(clientAddr, ":")[0],
					}
				} else {
					defs.RemoteStatus = v
				}
			case defs.Elevator:
				fmt.Printf("Unmarshaled Elevator from client %s.\n", clientAddr)
				// Handle Elevator-specific logic here
				if !utility.IsIPInMasterList(v.Ip, *masterElevator) {
					masterElevator.Elevators = append(masterElevator.Elevators, v)
				} else {
					elevData.UpdateOrdersMasterList(masterElevator, v.Orders, v.Ip)
				}

				jsonToSend := utility.MarshalJson(masterElevator)
				broadcast.BroadcastMessage(nil, jsonToSend)
			default:
				fmt.Printf("Received unknown type from client %s\n", clientAddr)
			}
		}
	}
}

func shutdownServer() {

	// Close all active client connections
	defs.ClientMutex.Lock()
	for conn := range defs.ClientConnections {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
		}
		delete(defs.ClientConnections, conn)
	}
	defs.ClientMutex.Unlock()

	// Finally, mark the server as not listening
	defs.ServerListening = false
	fmt.Println("Server has been shut down and all connections are closed.")
}

func slicesAreEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
