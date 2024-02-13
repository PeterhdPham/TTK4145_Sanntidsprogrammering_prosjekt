package tcp

import "net"

func startClient(serverAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	handleConnection(conn)
}

func TCP_Client(master_ip_addr string) {
	serverAddr := master_ip_addr + ":9999" // Use the server's actual IP address and port
	startClient(serverAddr)
}
