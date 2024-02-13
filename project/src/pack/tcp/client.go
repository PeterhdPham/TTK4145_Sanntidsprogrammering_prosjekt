package pack

import "net"

func startClient(serverAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	handleConnection(conn)
}

func TCP_Client() {
	serverAddr := "100.10.23.34:9999" // Use the server's actual IP address and port
	startClient(serverAddr)
}
