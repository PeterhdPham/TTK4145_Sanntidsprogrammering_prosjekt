package tcp

import (
	"fmt"
)

var masterIPAddress string

func TCP_Testing() {
	if masterIPAddress == "" {
		masterIPAddress = "127.0.0.1:9999"
	}
	conn, failed := startClient(masterIPAddress)
	if failed {
		fmt.Println("No active server found. Starting TCP Server...")
		TCP_Server()
	} else {
		fmt.Println("Connected to the server as a client.")
		TCP_Client(conn)
	}
}
