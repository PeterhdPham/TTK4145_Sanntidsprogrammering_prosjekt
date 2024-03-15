package communication

import (
	"log"
	"net"
	"project/defs"
	"project/utility"
	"time"
)

var error_buffer = 3
var ShouldReconnect bool

// Implement or adjust broadcastMessage to be compatible with the above modifications
func BroadcastMessage(origin net.Conn, masterElevator *defs.MasterList) error {

	message := utility.MarshalJson(masterElevator)

	defs.ClientMutex.Lock()
	defer defs.ClientMutex.Unlock()

	message = append(message, '%')

	for conn := range defs.ClientConnections {
		// Check if the message is not from the server (origin != nil) and conn is the origin, then skip
		if origin != nil && conn == origin {
			log.Println("Skipping connection")
			continue // Skip sending the message back to the origin client
		}

		for {
			_, err := conn.Write(message)
			if err != nil {
				log.Printf("Failed to broadcast to client %s: %s\n", conn.RemoteAddr(), err)
				if defs.ErrorBuffer == 0 {
					log.Println("Too many consecutive errors, stopping...")
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

func SendMessage(conn net.Conn, message interface{}, prefix string) error {
	// Marshal the message into JSON
	messageJson := utility.MarshalJson(message)

	if prefix != "" {
		messageJson = append([]byte(prefix), messageJson...)
	}

	messageJson = append(messageJson, '%')
	for {
		_, err := conn.Write(messageJson)
		if err != nil {
			log.Printf("Error sending message: %s\n", err)
			if error_buffer == 0 {
				log.Println("Too many consecutive errors, stopping...")
				ShouldReconnect = true
				return err // Stop if there are too many consecutive errors
			} else {
				error_buffer--
			}
		} else {
			error_buffer = 3 // Reset the error buffer on successful send
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	ShouldReconnect = false
	return nil
}
