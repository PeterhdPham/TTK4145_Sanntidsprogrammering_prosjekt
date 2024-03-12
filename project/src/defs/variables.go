package defs

import (
	"net"
	"sync"
)

// Variables for server
var (
	ServerListening       bool              = false // Flag to indicate if server is listening for new connections
	ClientConnections     map[net.Conn]bool         // Updated to track multiple client connections.
	ClientMutex           sync.Mutex                // Protects access to clientConnections
	ShouldServerReconnect bool                      // Flag to indicate if the server should reconnect
	ErrorBuffer           int               = 3     // How many error we can receive before giving up
)

// IP variables
var (
	ServerIP string //Server IP
	MyIP     string //IP address for current computer
)

// Channels used
var (
	UpdateLocal    = make(chan string)            // Channel for activing local update of MasterList on clients
	ButtonReceived = make(chan ButtonEventWithIP) // Channel for activing a new order assignment on server
	StatusReceived = make(chan string)            // Channel for update of status on the MasterList on server side
)

var RemoteStatus ElevStatus // ElevStatus for updating
