package roleConfiguration

import (
	"Driver-go/elevio"
	"context"
	"io"
	"log"
	"net"
	"project/communication"
	"project/elevatorData"
	"project/types"
	"project/utility"
	"project/variables"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	LivingIPsChan          = make(chan []string)
	ActiveIPsMutex         sync.Mutex
	ActiveIPs              []string
	connected              bool = false
	WaitingForConfirmation bool
	ServerActive           = make(chan bool)
	ReceivedPrevMasterList bool
	ReceivedFirstElevator  bool
)

func startServer(masterList *types.MasterList) {

	variables.ClientConnections = make(map[net.Conn]bool)
	ShouldReconnect = true

	var ctx context.Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ServerListening = true

	listenAddr := "0.0.0.0:55555"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("Failed to start server: %s\n", err)
		ServerListening = false
		return
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					closeAllClientConnections()
					ServerListening = false
					listener.Close()
					return
				case <-ServerActive:
					closeAllClientConnections()
					ServerListening = false
					listener.Close()
					return
				default:
					log.Printf("Failed to accept connection: %s\n", err)
					time.Sleep(time.Second)
					continue
				}
			}
			go handleClientMessages(conn, masterList)
		}
	}()

	select {
	case <-ServerActive:
		closeAllClientConnections()
		ServerListening = false
		listener.Close()
		return
	}

}

func closeAllClientConnections() {
	variables.ClientMutex.Lock()
	defer variables.ClientMutex.Unlock()

	for conn := range variables.ClientConnections {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %s\n", err)
		}
		delete(variables.ClientConnections, conn)
	}
}

func handleClientMessages(conn net.Conn, masterList *types.MasterList) {
	variables.ClientMutex.Lock()
	variables.ClientConnections[conn] = true
	variables.ClientMutex.Unlock()

	defer func() {
		conn.Close()
		variables.ClientMutex.Lock()
		delete(variables.ClientConnections, conn)
		variables.ClientMutex.Unlock()
	}()

	clientAddr := conn.RemoteAddr().String()

	for {
		buffer := make([]byte, BUFFER_SIZE)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected gracefully.\n", clientAddr)
			} else {
				log.Printf("Error reading from client %s: %s\n", clientAddr, err)
			}
			break
		}

		messages := strings.Split(string(buffer[:n]), "%")
		for _, message := range messages {
			if message == "" || message == " " || (!strings.HasPrefix(message, `{"elevators":`) && !strings.HasPrefix(message, `{"ip":`) && !strings.HasPrefix(message, `{"direction":`) && !strings.HasPrefix(message, `prev`) && !strings.HasPrefix(message, `init`)) {
				continue
			}

			if strings.HasPrefix(message, "prev") {
				message = strings.TrimPrefix(message, "prev")
				ReceivedPrevMasterList = true
			} else {
				ReceivedPrevMasterList = false
			}

			if strings.HasPrefix(message, "init") {
				message = strings.TrimPrefix(message, "init")
				ReceivedFirstElevator = true
			} else {
				ReceivedFirstElevator = false
			}

			genericStruct, err := utility.DetermineStructTypeAndUnmarshal([]byte(message))
			if err != nil {
				continue
			}

			switch v := genericStruct.(type) {
			case types.MasterList:
				if reflect.DeepEqual(v, *masterList) {
					continue
				} else {
					if ReceivedPrevMasterList {
						if utility.IPInMasterList(variables.MyIP, v) {
							for index := range masterList.Elevators {

								for v_index := range v.Elevators {
									if masterList.Elevators[index].Ip == v.Elevators[v_index].Ip {
										combinedOrders := utility.CombineOrders(masterList.Elevators[index].Orders, v.Elevators[v_index].Orders)
										elevatorData.UpdateLightsMasterList(masterList, variables.MyIP)
										masterList.Elevators[index].Status = v.Elevators[v_index].Status
										masterList.Elevators[index].Orders = combinedOrders
									}
								}
								if masterList.Elevators[index].Ip == variables.MyIP {
									masterList.Elevators[index].IsOnline = true
								}
							}
						} else {
							for index := range v.Elevators {
								if !(utility.IPInMasterList(v.Elevators[index].Ip, *masterList)) {
									(*masterList).Elevators = append((*masterList).Elevators, v.Elevators[index])
								}
							}
						}

						communication.BroadcastMessage(masterList)
						elevatorData.UpdateLightsMasterList(masterList, variables.MyIP)
						variables.UpdateLocal <- "true"
						ReceivedPrevMasterList = false
					}
				}
			case types.ElevStatus:
				requestFloor := v.Buttonfloor
				requestButton := v.Buttontype

				if requestButton != -1 || requestFloor != -1 {
					variables.RemoteStatus = v
					variables.ButtonReceived <- types.ButtonEventWithIP{
						Event: elevio.ButtonEvent{Floor: v.Buttonfloor, Button: elevio.ButtonType(v.Buttontype)},
						IP:    strings.Split(clientAddr, ":")[0],
					}
				} else {
					variables.RemoteStatus = v
					variables.StatusReceived <- strings.Split(clientAddr, ":")[0]
				}
			case types.Elevator:

				if !utility.IPInMasterList(v.Ip, *masterList) {
					masterList.Elevators = append(masterList.Elevators, v)
				} else {
					if ReceivedFirstElevator {
						for index_master := range masterList.Elevators {
							if masterList.Elevators[index_master].Ip == v.Ip {
								masterList.Elevators[index_master].Orders = utility.CombineOrders(masterList.Elevators[index_master].Orders, v.Orders)
							}
						}
						ReceivedFirstElevator = false
					} else {
						elevatorData.UpdateOrdersMasterList(masterList, v.Orders, v.Ip)
					}

					elevatorData.UpdateLightsMasterList(masterList, v.Ip)
				}

				communication.BroadcastMessage(masterList)
			default:
				log.Printf("Received unknown type from client %s\n", clientAddr)
			}
		}
	}
}

func shutdownServer() {

	variables.ClientMutex.Lock()
	for conn := range variables.ClientConnections {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %s\n", err)
		}
		delete(variables.ClientConnections, conn)
	}
	variables.ClientMutex.Unlock()

	ServerListening = false
}
