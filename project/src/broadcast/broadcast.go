package broadcast

import (
	"fmt"
	"net"
	"project/variable"
	"time"
)

// Implement or adjust broadcastMessage to be compatible with the above modifications
func BroadcastMessage(origin net.Conn, message []byte) error {
	variable.ClientMutex.Lock()
	defer variable.ClientMutex.Unlock()

	message = append(message, '\n')

	for conn := range variable.ClientConnections {
		// Check if the message is not from the server (origin != nil) and conn is the origin, then skip
		if origin != nil && conn == origin {
			fmt.Println("Skipping connection")
			continue // Skip sending the message back to the origin client
		}

		for {
			_, err := conn.Write(message)
			fmt.Println("Error: ", err)
			if err != nil {
				fmt.Printf("Failed to broadcast to client %s: %s\n", conn.RemoteAddr(), err)
				if variable.ErrorBuffer == 0 {
					fmt.Println("Too many consecutive errors, stopping...")
					variable.ShouldServerReconnect = true
					return err // Stop if there are too many consecutive errors
				} else {
					variable.ErrorBuffer--
				}
			} else {
				variable.ErrorBuffer = 3 // Reset the error buffer on successful send
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	variable.ShouldServerReconnect = false
	fmt.Println("Broadcasting done...")
	return nil
}
