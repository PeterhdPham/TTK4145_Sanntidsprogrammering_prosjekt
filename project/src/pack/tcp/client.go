package pack

import (
	"bufio"
	"fmt"
	"net"
)

func TCP_Client() {
	serverAddr := "127.0.0.1:9999" // Server address
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Connected to server")

	// Send a message to the server
	conn.Write([]byte("Hello from the client!\n"))

	// Optionally, listen for a response
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from server:", err)
		return
	}
	fmt.Print("Server says: ", response)
}
