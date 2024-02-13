package pack

import (
	"fmt"
	"net"
)

func handleClient(conn net.Conn) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", clientAddr)

	// Send a welcome message to the client
	conn.Write([]byte("Welcome, you are connected to the server.\n"))

	// Read messages from client and print them
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Client %s disconnected.\n", clientAddr)
			break
		}
		fmt.Printf("Message from %s: %s", clientAddr, string(buffer[:n]))
	}
}

func TCP_Server() {
	listenAddr := "0.0.0.0:9999" // Listen on all available interfaces
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("Server listening on", listenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn) // Handle each client connection concurrently
	}
}
