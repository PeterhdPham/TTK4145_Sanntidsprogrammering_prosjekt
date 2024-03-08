package tcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"project/elevData"
	"project/udp"
	"strings"
	"sync"
	"time"
)

var (
	LivingIPsChan         = make(chan []string)         //Stores living IPs from the Look_for_life function
	ActiveIPsMutex        sync.Mutex                    //Mutex for protecting active IPs
	ActiveIPs             []string                      //List of active IPs
	connected             bool                  = false //Client connection state
	ServerIP              string                        //Server IP
	MyIP                  string                        //IP address for current computer
	ShouldServerReconnect bool                          //Flag to indicate if the server should reconnect
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
			if !slicesAreEqual(ActiveIPs, livingIPs) {
				ActiveIPsMutex.Lock()
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
	if ServerIP != lowestIP {
		connected = false
		ServerIP = lowestIP
	}
	//Sets role to master if lowestIP is localhost
	if lowestIP == "127.0.0.1" {
		fmt.Println("Running on localhost")
		pointerElevator.Role = elevData.Master
		return
	}

	if MyIP == lowestIP && !serverListening {
		//Set role to master and starts a new server on
		shutdownServer()
		fmt.Println("This node is the server.")
		// port := strings.Split(ActiveIPs[0], ":")[1]
		go startServer(masterElevator) // Ensure server starts in a non-blocking manner
		pointerElevator.Role = elevData.Master
	} else if MyIP != lowestIP && serverListening {
		//Stops the server and switches from master to slave role
		fmt.Println("This node is no longer the server, transitioning to client...")
		shutdownServer()                                                       // Stop the server
		go connectToServer(lowestIP+":55555", pointerElevator, masterElevator) // Transition to client
		pointerElevator.Role = elevData.Slave
	} else if !serverListening {
		//Starts a client connection to the server, and sets role to slave
		if !connected {
			fmt.Println("This node is a client.")
			go connectToServer(lowestIP+":55555", pointerElevator, masterElevator)
			pointerElevator.Role = elevData.Slave
		}
	}
}

var (
	// Global cancellation context to control the server lifecycle
	serverCancel    context.CancelFunc = func() {} // No-op cancel function by default
	serverListening bool               = false
	// Updated to track multiple client connections.
	clientConnections map[net.Conn]bool
	clientMutex       sync.Mutex // Protects access to clientConnections
)

func startServer(masterElevator *elevData.MasterList) {
	// Initialize the map to track client connections at the correct scope
	clientConnections = make(map[net.Conn]bool)
	_, err := udp.GetPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return
	}
	// Check if the server is already running, and if so, initiate shutdown for role switch
	if serverListening {
		fmt.Println("Server is already running, attempting to shut down for role switch...")
		serverCancel()              // Request server shutdown
		time.Sleep(1 * time.Second) // Give it a moment to shut down before restarting
	}

	// Create a new context for this server instance
	var ctx context.Context
	ctx, serverCancel = context.WithCancel(context.Background())
	serverListening = true

	listenAddr := "0.0.0.0:55555"
	fmt.Println("Starting server at: " + listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		serverListening = false // Ensure the state reflects that the server didn't start
		return
	}
	defer func() {
		listener.Close()
		fmt.Println("Server listener closed.")
	}()
	fmt.Println("Server listening on", listenAddr)

	// Goroutine for server admin to broadcast messages to all clients
	// go func() {
	// 	ticker := time.NewTicker(3 * time.Second)
	// 	defer ticker.Stop()

	// 	for range ticker.C {
	// 		// Check if the only active IP is the server itself
	// 		if len(ActiveIPs) == 1 && ActiveIPs[0] == MyIP {
	// 			continue // Skip broadcasting
	// 		}

	// 		jsonData, err := json.Marshal(masterElevator)
	// 		if err != nil {
	// 			fmt.Printf("Error occurred during marshaling: %v", err)
	// 			return
	// 		}
	// 		// Broadcast the message to all connected clients
	// 		BroadcastMessage(ServerConnection, []byte(jsonData)) // Passing nil as the origin since this message is from the server
	// 		if connected {
	// 			break
	// 		}
	// 	}
	// }()

	// Accept new connections unless server shutdown is requested
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done(): // Shutdown was requested
					fmt.Println("Server shutting down...")
					closeAllClientConnections() // Ensure all client connections are gracefully closed
					serverListening = false
					return
				default:
					fmt.Printf("Failed to accept connection: %s\n", err)
					continue
				}
			}
			go handleConnection(conn)
		}
	}()

	// Wait for the shutdown signal to clean up and exit the function
	<-ctx.Done()
	fmt.Println("Server shutdown completed.")
}

// Ensure this function exists and is correctly implemented to close all client connections
func closeAllClientConnections() {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	for conn := range clientConnections {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
		}
		delete(clientConnections, conn)
	}
}

// Implement or adjust broadcastMessage to be compatible with the above modifications
func BroadcastMessage(origin net.Conn, message []byte) error {
	// fmt.Println("Server sending message: ", string(message))
	// Ensure the message ends with a newline character, which may be needed depending on the server's reading logic.
	if !bytes.HasSuffix(message, []byte("\n")) {
		message = append(message, '\n')
	}

	clientMutex.Lock()
	defer clientMutex.Unlock()

	for conn := range clientConnections {
		// Check if the message is not from the server (origin != nil) and conn is the origin, then skip
		if origin != nil && conn == origin {
			fmt.Println("Skipping connection")
			continue // Skip sending the message back to the origin client
		}

		for {
			_, err := conn.Write(message)
			fmt.Println("Error: ", err)
			if err != nil {
				fmt.Printf("Failed to broadcast to client %s: %s\n", conn.RemoteAddr(), err)
				if error_buffer == 0 {
					fmt.Println("Too many consecutive errors, stopping...")
					ShouldServerReconnect = true
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

		// Read the response from the client
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Failed to read response from client %s: %s\n", conn.RemoteAddr(), err)
			return err
		}

		// Unmarshal the response into a MasterList
		var responsemessage elevData.MasterList
		err = json.Unmarshal(buffer[:n], &responsemessage)
		if err != nil {
			fmt.Printf("Failed to unmarshal responsemessage: %s\n", err)
			return err
		}

		// Convert responsemessage to []byte
		responseBytes, err := json.Marshal(responsemessage)
		if err != nil {
			fmt.Printf("Failed to marshal responsemessage: %s\n", err)
			return err
		}

		// Compare the responseBytes with the message that was sent
		if !CompareMasterLists(message, responseBytes) {
			fmt.Printf("Client %s did not receive the correct masterList\n", conn.RemoteAddr())
			return errors.New("client did not receive the correct masterList")
		}else{
			fmt.Printf("Client %s received the correct masterList\n", conn.RemoteAddr())
		}
	}

	ShouldServerReconnect = false
	return nil
}

// Implement or adjust compareMasterLists to be compatible with the above modifications
func CompareMasterLists(list1, list2 []byte) bool {
	return bytes.Equal(list1, list2)

} // Handles individual client connections.
func handleConnection(conn net.Conn) {
	// addNew := true
	// for c, _ := range clientConnections {
	// 	if strings.Split(conn.RemoteAddr().String(), ":")[0] == strings.Split(c.RemoteAddr().String(), ":")[0] {
	// 		addNew = false
	// 		clientMutex.Lock()
	// 		delete(clientConnections, c)
	// 		clientMutex.Unlock()
	// 		break
	// 	}
	// }
	// if addNew {
	clientMutex.Lock()
	clientConnections[conn] = true
	clientMutex.Unlock()
	// }
	defer func() {
		conn.Close()
		clientMutex.Lock()
		delete(clientConnections, conn)
		clientMutex.Unlock()
	}()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", clientAddr)

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client %s disconnected gracefully.\n", clientAddr)
			} else {
				fmt.Printf("Error reading from client %s: %s\n", clientAddr, err)
			}
			break
		}
		message := string(buffer[:n])
		fmt.Printf("Received from client %s: %s\n", clientAddr, message)

		// Previously here was the logic to send a confirmation back to the client, which has been removed as per request.
	}
}

func shutdownServer() {
	// First, cancel the server context to stop accepting new connections
	serverCancel()

	// Next, close all active client connections
	clientMutex.Lock()
	for conn := range clientConnections {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
		}
		delete(clientConnections, conn)
	}
	clientMutex.Unlock()

	// Finally, mark the server as not listening
	serverListening = false
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
