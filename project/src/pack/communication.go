package pack

import (
	"fmt"
	"net"
	"os"
	"time"
)

func Broadcast_life() {
	fmt.Print("Running Broadcast_life")
	// Create a UDP address for broadcasting
	broadcastAddr := "localhost:9999" // Change this to your broadcast address and port
	conn, err := net.Dial("udp", broadcastAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Get the local address (sender's IP)
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			// Send a broadcast message
			message := fmt.Sprintf("Hello, UDP world! Time: %s, Sender IP: %s", t.Format(time.RFC3339), localAddr.IP.String())
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Broadcast message sent: %s\n", message)
		}
	}

}
