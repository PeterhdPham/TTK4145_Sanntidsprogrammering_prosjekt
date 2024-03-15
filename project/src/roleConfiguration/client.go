package roleConfiguration

import (
	"net"
	"project/communication"
	"project/types"
	"project/variables"

	"project/utility"
	"strings"
)

const BUFFER_SIZE = 32768

var ServerConnection net.Conn
var ServerError error
var ShouldReconnect bool
var UpdateLocal bool = false

func connectToServer(serverIP string, pointerElevator *types.Elevator, masterElevator *types.MasterList) {
	serverAddr := serverIP
	for {
		ServerConnection, ServerError = net.Dial("tcp", serverAddr)
		if ServerError != nil {
			connected = false
		} else {
			break
		}
		if ActiveIPs[0] != strings.Split(serverAddr, ":")[0] {
			return
		}
	}
	connected = true
	ShouldReconnect = false

	communication.SendMessage(ServerConnection, *pointerElevator, "init")

	communication.SendMessage(ServerConnection, *masterElevator, "prev")

	go func() {
		for {
			buffer := make([]byte, BUFFER_SIZE)
			n, err := ServerConnection.Read(buffer)

			if err != nil {

				return
			}

			messages := strings.Split(string(buffer[:n]), "%")
			for _, message := range messages {
				if message == "" || message == " " || !strings.HasSuffix(message, "}]}") || !strings.HasPrefix(message, `{"elevators":`) {
					continue
				}

				genericMessage, err := utility.DetermineStructTypeAndUnmarshal([]byte(message))
				if err != nil {
					continue
				}

				switch msg := genericMessage.(type) {
				case types.MasterList:

					*masterElevator = msg
					communication.SendMessage(ServerConnection, msg, "")
					variables.UpdateLocal <- "true"
				default:
					continue
				}

			}
		}
	}()

	for {
		if ShouldReconnect {
			break
		}
	}

	connected = false
	ServerConnection.Close()
}
