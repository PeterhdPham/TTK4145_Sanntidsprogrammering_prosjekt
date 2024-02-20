package tcp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
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
)

func Config_Roles() {
	go pack.Broadcast_life()
	go pack.Look_for_life(livingIPsChan)

	// Initialize a ticker that ticks every 10 seconds.
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
			// Every 10 seconds, check the role and update if necessary.
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

	if myIP == lowestIP && !serverListening {
		fmt.Println("This node is the server.")
		port := strings.Split(activeIPs[0], ":")[1]
		go startServer(port) // Ensure server starts in a non-blocking manner
	} else if myIP != lowestIP && serverListening {
		fmt.Println("This node is no longer the server, transitioning to client...")
		serverCancel() // Stop the server
		serverListening = false
		go connectToServer(activeIPs[0]) // Transition to client
	} else if !serverListening {
		fmt.Println("This node is a client.")
		go connectToServer(activeIPs[0])
	}
}

var (
	// Global cancellation context to control the server lifecycle
	serverCancel    context.CancelFunc = func() {} // No-op cancel function by default
	serverListening bool               = false
	currentConn     net.Conn
)

func startServer(port string) {
	if serverListening {
		fmt.Println("Server is already running, attempting to shut down for role switch...")
		serverCancel()              // Request server shutdown
		time.Sleep(1 * time.Second) // Give it a moment to shut down
	}

	var ctx context.Context
	ctx, serverCancel = context.WithCancel(context.Background())
	serverListening = true

	listenAddr := "0.0.0.0:" + port
	fmt.Println("Starting at: " + listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Server listening on %s\n", listenAddr)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter message to send: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg) // Remove newline character
			if currentConn != nil {
				_, err := currentConn.Write([]byte(msg))
				if err != nil {
					fmt.Printf("Failed to send message: %s\n", err)
					continue
				}
				lastMessage = msg
			} else {
				fmt.Println("No active connection to send a message.")
			}
		}
	}()

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
	// Ensure only one active connection is handled at a time
	currentConnMutex.Lock()
	if currentConn != nil {
		currentConn.Close() // Close the previous connection if any
	}
	currentConn = conn
	currentConnMutex.Unlock()

	defer func() {
		conn.Close()
		currentConnMutex.Lock()
		currentConn = nil // Reset currentConn when connection closes
		currentConnMutex.Unlock()
	}()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", clientAddr)

	for {
		buffer := make([]byte, 1024)
		n, readErr := conn.Read(buffer)
		if readErr != nil {
			fmt.Printf("Error reading from client %s: %s\n", clientAddr, readErr)
			break
		}
		message := string(buffer[:n])
		fmt.Printf("Received from client %s: %s\n", clientAddr, message)

		// Check if received message matches the last message
		if message != lastMessage {
			// If not, update the last message and send a confirmation
			lastMessage = message
			confirmationMsg := message
			_, writeErr := conn.Write([]byte(confirmationMsg))
			if writeErr != nil {
				fmt.Printf("Error sending confirmation to client %s: %s\n", clientAddr, writeErr)
				break
			}
		} else {
			// Optionally, handle the case where the message is a repeat
			fmt.Println("Received duplicate message, no confirmation sent.")
		}
	}
}

// Placeholder for client connection logic.// Connects to the TCP server.
func connectToServer(serverIP string) {
	serverAddr := fmt.Sprintf("%s", serverIP) // Ensure the server address format is correct, including port
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect to server: %s\n", err)
		return
	}
	defer conn.Close()
	fmt.Printf("Connected to server at %s\n", serverAddr)

	lastSentMessage := "" // Placeholder for the last message sent by the server
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			// If there's an error (including timeout), log it and break the loop
			fmt.Printf("Failed to read from server: %s\n", err)
			break // Exit the loop and function if we cannot read (server closed connection, etc.)
		}

		receivedMsg := string(buffer[:n])
		fmt.Printf("Received message: %s\n", receivedMsg)

		// Echo the received message back as confirmation if it does NOT match the last message sent by the server
		if receivedMsg != lastSentMessage {
			fmt.Println("Message does not match the last known sent message. Sending confirmation...")
			_, err = conn.Write([]byte(receivedMsg)) // Send confirmation back to the server
			if err != nil {
				fmt.Printf("Failed to send confirmation: %s\n", err)
				break // Break the loop if writing fails
			}
		} else {
			fmt.Println("Received message matches the last known sent message. No confirmation sent.")
		}

		// Assuming the received message becomes the new "last sent message" for subsequent comparisons
		lastSentMessage = receivedMsg
	}
}
