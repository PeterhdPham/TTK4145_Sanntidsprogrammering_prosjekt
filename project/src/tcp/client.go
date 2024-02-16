package tcp

import (
	"fmt"
	"net"
	"time"
)

func startClient(master_ip_addr string) (net.Conn, bool) {
	serverAddr := master_ip_addr + ":9999" // Use the server's actual IP address and port
	const maxAttempts = 3
	var conn net.Conn
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err = net.Dial("tcp", serverAddr)
		if err == nil {
			// Connection successful, do not defer conn.Close() here,
			// it should be managed by the caller of startClient.
			return conn, false // Return the connection and false indicating success.
		}
		fmt.Printf("Attempt %d failed: %v\n", attempt, err)
		time.Sleep(time.Second * 2) // Wait for 2 seconds before retrying.
	}

	return nil, true // Return nil for the connection and true indicating failure to connect.
}

func TCP_Client(conn net.Conn) {
	handleConnection(conn) // Assuming this function exists and processes the connection as needed

}
