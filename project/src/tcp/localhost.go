package tcp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	mutex1            = &sync.Mutex{}
	peers             = make(map[string]struct{}) // Track peers by port as key
	basePort          = 8000
	maxPort           = 8005
	myPort            int // This should be set to the port number the client or server is using
	isServer          = false
	serverAddress     = "0.0.0.0" // CHANGE THIS to the server's local network IP address
	wg                sync.WaitGroup
	serverConnections = make([]*net.Conn, 0)
	myIpAddr          = "0.0.0.0" // CHANGE THIS
)

func SetServerAddress(ip string) {
	fmt.Println("Setting server address: " + ip)
	if serverAddress == "0.0.0.0" {
		serverAddress = ip
		myIpAddr = ip
	} else if ip != serverAddress {
		serverAddress = ip
		HandleNewServerRole(ip)
	}
}

// Entry point to initialize TCP client or server
func NewTCP() {
	fmt.Println("Starting TCP client")
	isClient := false

	// Attempt to connect to a server on the known ports
	for port := basePort; port <= maxPort; port++ {
		fmt.Printf("Attempting to connect to %s\n", serverAddress)
		conn, err := net.Dial("tcp", serverAddress)
		if err == nil {
			fmt.Printf("Successfully connected to %s\n", serverAddress)
			isClient = true
			wg.Add(1) // Add to the wait group before starting the goroutine
			go handleClient(conn)
			break
		} else {
			fmt.Printf("Failed to connect to %s: %v\n", serverAddress, err)
		}
	}

	if !isClient {
		fmt.Println("No server found; acting as a server.")
		wg.Add(1) // Add to the wait group before starting the server goroutine
		go startServer1()
	}

	wg.Wait() // Wait here for all goroutines to finish
}

// HandleNewServerRole switches the role based on the new server IP
func HandleNewServerRole(newServerIP string) {
	if myIpAddr == newServerIP {
		// This machine is the new server
		becomeServer()
	} else {
		// This machine is a client to the new server
		becomeClient(newServerIP)
	}
}

func becomeServer() {
	// Logic to start server operations
	// This includes stopping any client operations if they were running
	SetServerAddress(myIpAddr)
	startServer1() // Assuming this is your server starting function
}

func becomeClient(serverIP string) {
	// Logic to start client operations directed to the new server IP
	SetServerAddress(serverIP)
	NewTCP() // This will now connect to the new server IP
}

func startServer1() {
	defer wg.Done() // Signal completion when returning

	isServer = true
	listener, err := net.Listen("tcp", fmt.Sprintf("%s", serverAddress))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return
	}
	defer listener.Close()

	// Retrieve the chosen port
	addr := listener.Addr().(*net.TCPAddr)
	myPort = addr.Port
	fmt.Printf("Serving as server on %s\n", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go handleServerConnection(conn)
	}
}

func handleClient(conn net.Conn) {
	defer wg.Done() // Signal completion when returning

	defer func() {
		conn.Close()
		fmt.Println("Disconnected from server, attempting to reconnect...")
		attemptReconnectionOrBecomeServer()
	}()

	fmt.Println("Clients connected")
	sendHeartbeat := time.NewTicker(5 * time.Second)
	defer sendHeartbeat.Stop()

	for {
		select {
		case <-sendHeartbeat.C:
			heartbeatMessage := fmt.Sprintf("Heartbeat from port %d\n", myPort)
			if _, err := conn.Write([]byte(heartbeatMessage)); err != nil {
				fmt.Printf("Failed to send heartbeat: %v\n", err)
				return // Connection will be closed and defer will trigger reconnection
			}
		}
	}
}

func closeServerConnections() {
	mutex1.Lock()
	defer mutex1.Unlock()
	for _, conn := range serverConnections {
		(*conn).Close()
	}
	serverConnections = make([]*net.Conn, 0) // Reset the slice
}

func resetConnections() {
	fmt.Println("Resetting connections due to server address change...")

	// Close server connections
	closeServerConnections()

	// Additional logic to stop the server listen loop if necessary

	// Reset client connections if applicable
	// ...

	// Restart client or server based on the current role
	if isServer {
		// Restart server logic
		go startServer1()
	} else {
		// Restart client connection attempts
		go NewTCP()
	}
}

func handleServerConnection(conn net.Conn) {

	mutex1.Lock()
	serverConnections = append(serverConnections, &conn)
	mutex1.Unlock()

	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected")
			} else {
				fmt.Printf("Error reading from client: %v\n", err)
			}
			return
		}
		fmt.Printf("Received from client: %s", message)
		// Echo the message back to the client
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Error sending to client: %v\n", err)
			return
		}
	}
}

func distributePeers(conn net.Conn) {
	mutex1.Lock()
	defer mutex1.Unlock()

	for peer := range peers {
		fmt.Fprintf(conn, "Peer connected: %s\n", peer)
	}
}

func attemptReconnectionOrBecomeServer() {
	if !tryReconnect() {
		checkAndBecomeServer()
	}
}

// Add this function to determine whether to act as a server or client
func DetermineRole(lowestIP string) {
	if myIpAddr == lowestIP {
		// This machine has the lowest IP, so it should act as the server
		fmt.Println("Acting as server.")
		startServer1()
	} else {
		// This machine does not have the lowest IP, so it should act as a client
		fmt.Println("Acting as client.")
		NewTCP()
	}
}

func tryReconnect() bool {
	for port := basePort; port <= maxPort; port++ {
		if port == myPort {
			continue // Skip own port if it was a server
		}
		print(port)
		address := fmt.Sprintf("%s:%d", serverAddress, port)
		conn, err := net.Dial("tcp", address)
		if err == nil {
			fmt.Printf("Reconnected as client to %s\n", address)
			wg.Add(1) // Add to the wait group before starting the goroutine
			go handleClient(conn)
			return true
		}
	}
	return false
}

func checkAndBecomeServer() {
	mutex1.Lock()
	defer mutex1.Unlock()

	// Convert peers map keys to a slice of ints for sorting
	var peerPorts []int
	for peer := range peers {
		port, err := strconv.Atoi(peer)
		if err != nil {
			continue
		}
		peerPorts = append(peerPorts, port)
	}

	// Include own port in the sorting process to find the lowest
	peerPorts = append(peerPorts, myPort)
	sort.Ints(peerPorts)

	// If myPort is the lowest, become the server
	if peerPorts[0] == myPort {
		fmt.Println("Becoming the server...")
		SetServerAddress(myIpAddr)
		startServer1()
	}
}
