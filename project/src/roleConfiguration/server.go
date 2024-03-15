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

func startServer(masterElevator *types.MasterList) {

	variables.ClientConnections = make(map[net.Conn]bool)
	ShouldReconnect = true

	if ServerListening {
		log.Println("Server is already running, attempting to shut down for role switch...")
		time.Sleep(1 * time.Second)
	}

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

	log.Println("Server listening on", listenAddr)

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
			go handleClientMessages(conn, masterElevator)
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

func handleClientMessages(conn net.Conn, masterElevator *types.MasterList) {
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
	log.Printf("Client connected: %s\n", clientAddr)

	for {
		buffer := make([]byte, 32768)
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
				log.Printf("Failed to determine struct type or unmarshal message from client %s: %s\n", clientAddr, err)
				continue
			}

			switch v := genericStruct.(type) {
			case types.MasterList:
				if reflect.DeepEqual(v, *masterElevator) {
					continue
				} else {
					if ReceivedPrevMasterList {
						if utility.IPInMasterList(variables.MyIP, v) {
							for index := range masterElevator.Elevators {

								for v_index := range v.Elevators {
									if masterElevator.Elevators[index].Ip == v.Elevators[v_index].Ip {
										combinedOrders := utility.CombineOrders(masterElevator.Elevators[index].Orders, v.Elevators[v_index].Orders)
										elevatorData.UpdateLightsMasterList(masterElevator, variables.MyIP)
										masterElevator.Elevators[index].Status = v.Elevators[v_index].Status
										masterElevator.Elevators[index].Orders = combinedOrders
									}
								}
								if masterElevator.Elevators[index].Ip == variables.MyIP {
									masterElevator.Elevators[index].IsOnline = true
								}
							}
							log.Println("Overwriting existing masterList")
						} else {
							for index := range v.Elevators {
								if !(utility.IPInMasterList(v.Elevators[index].Ip, *masterElevator)) {
									(*masterElevator).Elevators = append((*masterElevator).Elevators, v.Elevators[index])
									log.Printf("Adding %s to current masterList", v.Elevators[index].Ip)
								}
							}
						}

						communication.BroadcastMessage(masterElevator)
						elevatorData.UpdateLightsMasterList(masterElevator, variables.MyIP)
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

				if !utility.IPInMasterList(v.Ip, *masterElevator) {
					masterElevator.Elevators = append(masterElevator.Elevators, v)
				} else {
					if ReceivedFirstElevator {
						for index_master := range masterElevator.Elevators {
							if masterElevator.Elevators[index_master].Ip == v.Ip {
								masterElevator.Elevators[index_master].Orders = utility.CombineOrders(masterElevator.Elevators[index_master].Orders, v.Orders)
							}
						}
						ReceivedFirstElevator = false
					} else {
						elevatorData.UpdateOrdersMasterList(masterElevator, v.Orders, v.Ip)
					}

					elevatorData.UpdateLightsMasterList(masterElevator, v.Ip)
				}

				communication.BroadcastMessage(masterElevator)
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
	log.Println("Server has been shut down and all connections are closed.")
}
