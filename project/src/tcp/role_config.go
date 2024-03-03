package tcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"project/udp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	// This channel is used to receive the living IPs from the Look_for_life function.
	LivingIPsChan = make(chan []string)
	// Mutex to protect access to the ActiveIPs slice.
	ActiveIPsMutex sync.Mutex
	// Slice to store the active IPs.
	ActiveIPs        []string
	currentConnMutex sync.Mutex
	lastMessage      string
	connected        bool = false
	serverIP         string
	lowestIP         string
)

func Config_Roles() {
	go udp.BroadcastLife()
	go udp.LookForLife(LivingIPsChan)

	// Initialize a ticker that ticks every 1 seconds.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case livingIPs := <-LivingIPsChan:
			// Update the list of active IPs whenever a new list is received.
			ActiveIPsMutex.Lock()
			ActiveIPs = livingIPs
			ActiveIPsMutex.Unlock()
		case <-ticker.C:
			// Every 1 seconds, check the role and update if necessary.
			updateRole()
		}
	}
}
func updateRole() {
	ActiveIPsMutex.Lock()
	defer ActiveIPsMutex.Unlock()

	if len(ActiveIPs) == 0 {
		fmt.Println("No active IPs found. Waiting for discovery...")
		return
	}

	sort.Strings(ActiveIPs)

	myIP, err := GetPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return
	}
	lowestIP := strings.Split(ActiveIPs[0], ":")[0]
	if serverIP != lowestIP {
		connected = false
		serverIP = lowestIP
	}

	if !connected {
		if myIP == lowestIP && !serverListening {
			shutdownServer()
			fmt.Println("This node is the server.")
			port := strings.Split(ActiveIPs[0], ":")[1]
			go startServer(port) // Ensure server starts in a non-blocking manner
		} else if myIP != lowestIP && serverListening {
			fmt.Println("This node is no longer the server, transitioning to client...")
			shutdownServer() // Stop the server
			serverListening = false
			go connectToServer(ActiveIPs[0]) // Transition to client
			connected = true
		} else if !serverListening {
			if !connected {
				fmt.Println("This node is a client.")
				go connectToServer(ActiveIPs[0])
				connected = true
			}
		}
	}
	// else {
	// 	fmt.Println("Currently connected as a client, delaying role switch.")
	// }
}

var (
	// Global cancellation context to control the server lifecycle
	serverCancel    context.CancelFunc = func() {} // No-op cancel function by default
	serverListening bool               = false
	// Updated to track multiple client connections.
	clientConnections map[net.Conn]bool
	clientMutex       sync.Mutex // Protects access to clientConnections
)

func startServer(port string) {
	// Initialize the map to track client connections at the correct scope
	clientConnections = make(map[net.Conn]bool)

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

	listenAddr := "0.0.0.0:" + port
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
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter message to broadcast: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg) // Remove newline character
			lastMessage = msg
			// Broadcast the message to all connected clients
			broadcastMessage(msg, nil) // Passing nil as the origin since this message is from the server
			if connected {
				break
			}
		}

	}()

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
	// Additional cleanup can be performed here if necessary
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
func broadcastMessage(message string, origin net.Conn) {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	for conn := range clientConnections {
		// Check if the message is not from the server (origin != nil) and conn is the origin, then skip
		if origin != nil && conn == origin {
			continue // Skip sending the message back to the origin client
		}
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Failed to broadcast to client %s: %s\n", conn.RemoteAddr(), err)
			// Handle failed send e.g., by removing the client connection if necessary
		}
	}
}

// Handles individual client connections.
func handleConnection(conn net.Conn) {
	clientMutex.Lock()
	clientConnections[conn] = true
	clientMutex.Unlock()

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

// Placeholder for client connection logic.// Connects to the TCP server.
// Connects to the TCP server.
var error_buffer = 3

func connectToServer(serverIP string) {
	serverAddr := serverIP
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect to server: %s\n", err)
		connected = false
		return
	}
	defer conn.Close()
	fmt.Println("Connected to server at", serverAddr)
	connected = true

	// Start a goroutine to listen for messages from the server
	go func() {
		for {
			buffer := make([]byte, 1024) // Create a buffer to store incoming data
			n, err := conn.Read(buffer)  // Read data into buffer
			if err != nil {
				if err == io.EOF {
					fmt.Println("Server closed the connection.")
				} else {
					fmt.Printf("Error reading from server: %s\n", err)
				}
				connected = false
				conn.Close()
				return // Exit goroutine if connection is closed or an error occurs
			}

			// Convert the bytes read into a string and print it
			message := string(buffer[:n])
			fmt.Printf("Message from server: %s\n", message)
		}
	}()

	// Read messages from stdin and send them to the server
	fmt.Println("Enter messages to send to the server. Type 'exit' to disconnect:")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		if msg == "exit" {
			fmt.Println("Disconnecting from server...")
			break
		}

		// SendMessage is assumed to be a function that sends a message to the server.
		err := SendMessage(conn, msg)
		if err != nil {
			fmt.Printf("Error sending message: %s\n", err)
			if error_buffer == 0 {
				error_buffer = 3
				break
			} else {
				error_buffer--
			}

			// break // Exit if there was an error sending the message
		}
	}

	connected = false
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
