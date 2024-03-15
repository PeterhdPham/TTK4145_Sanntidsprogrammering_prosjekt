package variables

import (
	"net"
	"project/types"
	"sync"
)

// Variables for server
var ClientConnections map[net.Conn]bool // Updated to track multiple client connections.
var ClientMutex sync.Mutex              // Protects access to clientConnections

// IP variables

var MyIP string //IP address for current computer

// Channels used

var UpdateLocal = make(chan string)                     // Channel for activing local update of MasterList on clients
var ButtonReceived = make(chan types.ButtonEventWithIP) // Channel for activing a new order assignment on server
var StatusReceived = make(chan string)                  // Channel for update of status on the MasterList on server side

var RemoteStatus types.ElevStatus // ElevStatus for updating
