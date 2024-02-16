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
	mutex1        = &sync.Mutex{}
	peers         = make(map[string]struct{}) // Track peers by port as key
	basePort      = 8000
	maxPort       = 8005
	myPort        int // This should be set to the port number the client or server is using
	isServer      = false
	serverAddress = "localhost"
)
var wg sync.WaitGroup

// Entry point to initialize TCP client or server
func NewTCP() {
	isClient := false

	// Attempt to connect to a server on the known ports
	for port := basePort; port <= maxPort; port++ {
		address := fmt.Sprintf("%s:%d", serverAddress, port)
		fmt.Printf("Attempting to connect to %s\n", address)
		conn, err := net.Dial("tcp", address)
		if err == nil {
			fmt.Printf("Successfully connected to %s\n", address)
			isClient = true
			wg.Add(1) // Add to the wait group before starting the goroutine
			go handleClient(conn)
			break
		} else {
			fmt.Printf("Failed to connect to %s: %v\n", address, err)
		}
	}

	if !isClient {
		fmt.Println("No server found; acting as a server.")
		wg.Add(1) // Add to the wait group before starting the server goroutine
		go startServer1()
	}

	wg.Wait() // Wait here for all goroutines to finish
}

func startServer1() {
	defer wg.Done() // Signal completion when returning

	isServer = true
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:8000", serverAddress))
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

	fmt.Println("test")
	defer func() {
		conn.Close()
		fmt.Println("Disconnected from server, attempting to reconnect...")
		attemptReconnectionOrBecomeServer()
	}()

	fmt.Println("CLients connected")
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

func handleServerConnection(conn net.Conn) {
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

func tryReconnect() bool {
	for port := basePort; port <= maxPort; port++ {
		if port == myPort {
			continue // Skip own port if it was a server
		}
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
		startServer1()
	}
}
