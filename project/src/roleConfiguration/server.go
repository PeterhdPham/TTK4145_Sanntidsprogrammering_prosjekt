package roleConfiguration

import (
	"Driver-go/elevio"
	"context"
	"io"
	"log"
	"net"
	"project/communication"
	"project/defs"
	"project/elevatorData"
	"project/utility"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	LivingIPsChan          = make(chan []string)         //Stores living IPs from the Look_for_life function
	ActiveIPsMutex         sync.Mutex                    //Mutex for protecting active IPs
	ActiveIPs              []string                      //List of active IPs
	connected              bool                  = false //Client connection state
	WaitingForConfirmation bool                          //
	ServerActive           = make(chan bool)             //Server state
	ReceivedPrevMasterList bool                          // Master list that server receives from client that used to be server
	ReceivedFirstElevator  bool                          // First elevator
)

func startServer(masterElevator *defs.MasterList) {
	// Initialize the map to track client connections at the correct scope
	defs.ClientConnections = make(map[net.Conn]bool)
	ShouldReconnect = true

	// Check if the server is already running, and if so, initiate shutdown for role switch
	if defs.ServerListening {
		log.Println("Server is already running, attempting to shut down for role switch...")
		time.Sleep(1 * time.Second) // Give it a moment to shut down before restarting
	}

	// Create a new context for this server instance
	var ctx context.Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defs.ServerListening = true

	listenAddr := "0.0.0.0:55555"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("Failed to start server: %s\n", err)
		defs.ServerListening = false // Ensure the state reflects that the server didn't start
		return
	}

	log.Println("Server listening on", listenAddr)

	// Accept new connections unless server shutdown is requested
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done(): // Shutdown was requested
					closeAllClientConnections() // Ensure all client connections are gracefully closed
					defs.ServerListening = false
					listener.Close()
					return
				case <-ServerActive:
					closeAllClientConnections() // Ensure all client connections are gracefully closed
					defs.ServerListening = false
					listener.Close()
					return
				default:
					log.Printf("Failed to accept connection: %s\n", err)
					time.Sleep(time.Second)
					continue
				}
			}
			go handleConnection(conn, masterElevator)
		}
	}()

	// Wait for the shutdown signal to clean up and exit the function
	// <-ctx.Done()
	select {
	case <-ServerActive:
		closeAllClientConnections() // Ensure all client connections are gracefully closed
		defs.ServerListening = false
		listener.Close()
		return
	}

}

// Ensure this function exists and is correctly implemented to close all client connections
func closeAllClientConnections() {
	defs.ClientMutex.Lock()
	defer defs.ClientMutex.Unlock()

	for conn := range defs.ClientConnections {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %s\n", err)
		}
		delete(defs.ClientConnections, conn)
	}
}

// Handles individual client connections.
func handleConnection(conn net.Conn, masterElevator *defs.MasterList) {
	defs.ClientMutex.Lock()
	defs.ClientConnections[conn] = true
	defs.ClientMutex.Unlock()

	defer func() {
		conn.Close()
		defs.ClientMutex.Lock()
		delete(defs.ClientConnections, conn)
		defs.ClientMutex.Unlock()
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

		// Process each newline-separated message
		messages := strings.Split(string(buffer[:n]), "%")
		for _, message := range messages {
			if message == "" || message == " " || (!strings.HasPrefix(message, `{"elevators":`) && !strings.HasPrefix(message, `{"ip":`) && !strings.HasPrefix(message, `{"direction":`) && !strings.HasPrefix(message, `prev`) && !strings.HasPrefix(message, `init`)) {
				continue // Skip empty messages
			}

			// Checks if the message contains a tag for previous master list
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

			// Attempt to determine the struct type from the JSON keys
			genericStruct, err := utility.DetermineStructTypeAndUnmarshal([]byte(message))
			if err != nil {
				log.Printf("Failed to determine struct type or unmarshal message from client %s: %s\n", clientAddr, err)
				continue
			}

			// Now handle the unmarshaled data based on its determined type
			switch v := genericStruct.(type) {
			case defs.MasterList:
				if reflect.DeepEqual(v, *masterElevator) {
					continue
				} else {
					if ReceivedPrevMasterList {
						if utility.IsIPInMasterList(defs.MyIP, v) {
							for index := range masterElevator.Elevators {

								for v_index := range v.Elevators {
									if masterElevator.Elevators[index].Ip == v.Elevators[v_index].Ip {
										combinedOrders := utility.CombineOrders(masterElevator.Elevators[index].Orders, v.Elevators[v_index].Orders)
										elevatorData.UpdateLightsMasterList(masterElevator, defs.MyIP)
										masterElevator.Elevators[index].Status = v.Elevators[v_index].Status
										masterElevator.Elevators[index].Orders = combinedOrders
									}
								}
								if masterElevator.Elevators[index].Ip == defs.MyIP {
									masterElevator.Elevators[index].IsOnline = true
								}
							}
							log.Println("Overwriting existing masterList")
						} else {
							for index := range v.Elevators {
								if !(utility.IsIPInMasterList(v.Elevators[index].Ip, *masterElevator)) {
									(*masterElevator).Elevators = append((*masterElevator).Elevators, v.Elevators[index])
									log.Printf("Adding %s to current masterList", v.Elevators[index].Ip)
								}
							}
						}

						communication.BroadcastMessage(nil, masterElevator)
						elevatorData.UpdateLightsMasterList(masterElevator, defs.MyIP)
						defs.UpdateLocal <- "true"
						ReceivedPrevMasterList = false
					}
				}
			case defs.ElevStatus:
				requestFloor := v.Buttonfloor
				requestButton := v.Buttontype
				// Handle ElevStatus-specific logic here
				if requestButton != -1 || requestFloor != -1 {
					defs.RemoteStatus = v
					defs.ButtonReceived <- defs.ButtonEventWithIP{
						Event: elevio.ButtonEvent{Floor: v.Buttonfloor, Button: elevio.ButtonType(v.Buttontype)},
						IP:    strings.Split(clientAddr, ":")[0],
					}
				} else {
					defs.RemoteStatus = v
					defs.StatusReceived <- strings.Split(clientAddr, ":")[0]
				}
			case defs.Elevator:
				// Handle Elevator-specific logic here
				if !utility.IsIPInMasterList(v.Ip, *masterElevator) {
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

				communication.BroadcastMessage(nil, masterElevator)
			default:
				log.Printf("Received unknown type from client %s\n", clientAddr)
			}
		}
	}
}

func shutdownServer() {

	// Close all active client connections
	defs.ClientMutex.Lock()
	for conn := range defs.ClientConnections {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %s\n", err)
		}
		delete(defs.ClientConnections, conn)
	}
	defs.ClientMutex.Unlock()

	// Finally, mark the server as not listening
	defs.ServerListening = false
	log.Println("Server has been shut down and all connections are closed.")
}
