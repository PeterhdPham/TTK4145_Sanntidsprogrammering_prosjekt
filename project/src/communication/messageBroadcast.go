package communication

import (
	"log"
	"net"
	"project/types"
	"project/utility"
	"project/variables"
	"time"
)

const DELAY_ATTEMPTS = 100 * time.Millisecond
const MAX_ALLOWED_CONSECUTIVE_ERRORS = 3

var errorBuffer = 0
var ShouldReconnect bool

func BroadcastMessage(masterList *types.MasterList) error {

	message := utility.MarshalJson(masterList)

	variables.ClientMutex.Lock()
	defer variables.ClientMutex.Unlock()

	message = append(message, '%')

	for conn := range variables.ClientConnections {

		for {
			_, err := conn.Write(message)
			if err != nil {
				log.Printf("Failed to broadcast to client %s: %s\n", conn.RemoteAddr(), err)
				if errorBuffer == MAX_ALLOWED_CONSECUTIVE_ERRORS {
					log.Println("Failed to broadcast message to client consecutive times, stopping...")
					return err
				} else {
					errorBuffer++
				}
			} else {
				errorBuffer = 0
				break
			}
			time.Sleep(DELAY_ATTEMPTS)
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
			if errorBuffer == MAX_ALLOWED_CONSECUTIVE_ERRORS {
				log.Println("Failed to send message to server consecutive times, stopping...")
				ShouldReconnect = true
				return err
			} else {
				errorBuffer++
			}
		} else {
			errorBuffer = 0
			break
		}
		time.Sleep(DELAY_ATTEMPTS)
	}

	ShouldReconnect = false
	return nil
}
