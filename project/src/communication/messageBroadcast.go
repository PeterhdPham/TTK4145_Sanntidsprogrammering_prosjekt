package communication

import (
	"log"
	"net"
	"project/types"
	"project/utility"
	"project/variables"
	"time"
)

var errorBuffer = 3
var ShouldReconnect bool

func BroadcastMessage(masterElevator *types.MasterList) error {

	message := utility.MarshalJson(masterElevator)

	variables.ClientMutex.Lock()
	defer variables.ClientMutex.Unlock()

	message = append(message, '%')

	for conn := range variables.ClientConnections {

		for {
			_, err := conn.Write(message)
			if err != nil {
				log.Printf("Failed to broadcast to client %s: %s\n", conn.RemoteAddr(), err)
				if errorBuffer == 0 {
					log.Println("Too many consecutive errors, stopping...")
					return err
				} else {
					errorBuffer--
				}
			} else {
				errorBuffer = 3
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

func SendMessage(conn net.Conn, message interface{}, prefix string) error {

	messageJson := utility.MarshalJson(message)

	if prefix != "" {
		messageJson = append([]byte(prefix), messageJson...)
	}

	messageJson = append(messageJson, '%')
	for {
		_, err := conn.Write(messageJson)
		if err != nil {
			log.Printf("Error sending message: %s\n", err)
			if errorBuffer == 0 {
				log.Println("Too many consecutive errors, stopping...")
				ShouldReconnect = true
				return err
			} else {
				errorBuffer--
			}
		} else {
			errorBuffer = 3
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	ShouldReconnect = false
	return nil
}
