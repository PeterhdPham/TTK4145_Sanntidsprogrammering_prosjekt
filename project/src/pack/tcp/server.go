package pack

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Client connected:", conn.RemoteAddr())

	// Read data from the connection
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}
	fmt.Println("Received:", string(buffer[:n]))

	// Optionally, send a response
	_, err = conn.Write([]byte("Hello, client!"))
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	}
}

func TCP_Server() {
	listenAddr := "0.0.0.0:9999" // Listen on all network interfaces
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
		go handleConnection(conn)
	}
}
