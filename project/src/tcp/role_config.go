package tcp

import (
	"context"
	"fmt"
	"net"
	"project/pack"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	// This channel is used to receive the living IPs from the Look_for_life function.
	livingIPsChan = make(chan []string)
	// Mutex to protect access to the activeIPs slice.
	activeIPsMutex sync.Mutex
	// Slice to store the active IPs.
	activeIPs        []string
	currentConnMutex sync.Mutex
	lastMessage      string
	connected        bool = false
	serverIP         string
	lowestIP         string
)

func Config_Roles() {
	go pack.Broadcast_life()
	go pack.Look_for_life(livingIPsChan)

	// Initialize a ticker that ticks every 1 seconds.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case livingIPs := <-livingIPsChan:
			// Update the list of active IPs whenever a new list is received.
			activeIPsMutex.Lock()
			activeIPs = livingIPs
			activeIPsMutex.Unlock()
		case <-ticker.C:
			// Every 1 seconds, check the role and update if necessary.
			updateRole()
		}
	}
}
func updateRole() {
	activeIPsMutex.Lock()
	defer activeIPsMutex.Unlock()

	if len(activeIPs) == 0 {
		fmt.Println("No active IPs found. Waiting for discovery...")
		return
	}

	sort.Strings(activeIPs)

	myIP, err := getPrimaryIP()
	if err != nil {
		fmt.Println("Error obtaining the primary IP:", err)
		return
	}
	lowestIP := strings.Split(activeIPs[0], ":")[0]
	if serverIP != lowestIP {
		connected = false
		serverIP = lowestIP
	}

	if myIP == lowestIP && !serverListening {
		fmt.Println("This node is the server.")
		port := strings.Split(activeIPs[0], ":")[1]
		go startServer(port) // Ensure server starts in a non-blocking manner
		connected = false
	} else if myIP != lowestIP && serverListening {
		fmt.Println("This node is no longer the server, transitioning to client...")
		shutdownServer() // Stop the server
		serverListening = false
		go connectToServer(activeIPs[0]) // Transition to client
	} else if !serverListening {
		if !connected {
			fmt.Println("This node is a client.")
			go connectToServer(activeIPs[0])
			connected = true
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

func startServer(port string) {
	clientConnections = make(map[net.Conn]bool) // Ensure this is at the right scope to track connections

	if serverListening {
		fmt.Println("Server is already running, attempting to shut down for role switch...")
		serverCancel()              // Request server shutdown
		time.Sleep(1 * time.Second) // Give it a moment to shut down
	}

	var ctx context.Context
	ctx, serverCancel = context.WithCancel(context.Background())
	serverListening = true

	listenAddr := "0.0.0.0:" + port
	fmt.Println("Starting server at: " + listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on", listenAddr)

	for {
		// Accept new connections unless server shutdown is requested
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done(): // Check if shutdown was requested
				fmt.Println("Server shutting down...")
				serverListening = false
				return
			default:
				fmt.Printf("Failed to accept connection: %s\n", err)
				continue
			}
		}
		go handleConnection(conn)
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
}

// Placeholder for client connection logic.// Connects to the TCP server.
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
