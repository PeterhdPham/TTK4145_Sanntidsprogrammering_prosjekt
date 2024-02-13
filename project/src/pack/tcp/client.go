package pack

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func TCP_Client() {
	serverAddr := "10.100.23.34:9999" // Use the server's IP address and port
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Connected to server", serverAddr)

	// Send a message to the server
	message := "Hello, server!"
	conn.Write([]byte(message))

	// Read the server's response
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}
	fmt.Print("Server says: ", response)
}
