package broadcast

import (
	"fmt"
	"net"
	"project/defs"
	"project/utility"
	"time"
)

// Implement or adjust broadcastMessage to be compatible with the above modifications
func BroadcastMessage(origin net.Conn, masterElevator *defs.MasterList) error {

	message := utility.MarshalJson(masterElevator)

	defs.ClientMutex.Lock()
	defer defs.ClientMutex.Unlock()

	message = append(message, '%')

	for conn := range defs.ClientConnections {
		// Check if the message is not from the server (origin != nil) and conn is the origin, then skip
		if origin != nil && conn == origin {
			fmt.Println("Skipping connection")
			continue // Skip sending the message back to the origin client
		}

		for {
			_, err := conn.Write(message)
			if err != nil {
				fmt.Printf("Failed to broadcast to client %s: %s\n", conn.RemoteAddr(), err)
				if defs.ErrorBuffer == 0 {
					fmt.Println("Too many consecutive errors, stopping...")
					defs.ShouldServerReconnect = true
					return err // Stop if there are too many consecutive errors
				} else {
					defs.ErrorBuffer--
				}
			} else {
				defs.ErrorBuffer = 3 // Reset the error buffer on successful send
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	defs.ShouldServerReconnect = false
	return nil
}
