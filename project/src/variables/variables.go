package variables

import (
	"net"
	"project/types"
	"sync"
)

var ClientConnections map[net.Conn]bool
var ClientMutex sync.Mutex

var MyIP string

var UpdateLocal = make(chan string)
var ButtonReceived = make(chan types.ButtonEventWithIP)
var StatusReceived = make(chan string)

var RemoteStatus types.ElevStatus
