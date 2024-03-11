package defs

import (
	"net"
	"sync"
)

var UpdateMutex sync.Mutex
var UpdateOrdersFromMessage bool = false
var UpdateStatusFromMessage bool = false

var (
	ServerListening       bool              = false
	ClientConnections     map[net.Conn]bool // Updated to track multiple client connections.
	ClientMutex           sync.Mutex        // Protects access to clientConnections
	ServerIP              string            //Server IP
	MyIP                  string            //IP address for current computer
	ShouldServerReconnect bool              //Flag to indicate if the server should reconnect
	ErrorBuffer           = 3
)

var UpdateLocal = make(chan string)
var ButtonReceived = make(chan ButtonEventWithIP)
var StatusReceived = make(chan string)
