package variable

import (
	"context"
	"net"
	"sync"
)

var UpdateLocal bool = false
var UpdateMutex sync.Mutex
var UpdateOrdersFromMessage bool = false
var UpdateStatusFromMessage bool = false

var (
	ServerCancel          context.CancelFunc = func() {} // No-op cancel function by default
	ServerListening       bool               = false
	ClientConnections     map[net.Conn]bool  // Updated to track multiple client connections.
	ClientMutex           sync.Mutex         // Protects access to clientConnections
	ServerIP              string             //Server IP
	MyIP                  string             //IP address for current computer
	ShouldServerReconnect bool               //Flag to indicate if the server should reconnect
	ErrorBuffer           = 3
)



var ButtonReceived = make(chan ButtonEventWithIP)
var StatusReceived = make(chan string)
