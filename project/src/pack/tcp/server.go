package pack

import (
	"fmt"
	"net"
)

func startServer(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("Server listening on port", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func TCP_Server() {
	port := "9999" // Example port
	startServer(port)
}
