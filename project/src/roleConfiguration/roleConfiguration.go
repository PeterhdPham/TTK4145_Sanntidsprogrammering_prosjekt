package roleConfiguration

import (
	"Driver-go/elevio"
	"log"
	"project/aliveMessages"
	"project/communication"
	"project/defs"
	"project/elevatorData"
	"project/orderAssignment"
	"project/utility"
	"strings"
	"time"
)

var ServerListening bool = false // Flag to indicate if server is listening for new connections
var ServerIP string              //Server IP

func Config_Roles(pointerElevator *defs.Elevator, masterElevator *defs.MasterList) {
	//Go routines for finding active IPs
	go aliveMessages.BroadcastLife()
	go aliveMessages.LookForLife(LivingIPsChan)

	time.Sleep(3 * time.Second)

	for {
		select {
		case livingIPs := <-LivingIPsChan:
			// Update the list of active IPs whenever a new list is received.
			if !utility.StringSlicesAreEqual(ActiveIPs, livingIPs) {
				ActiveIPsMutex.Lock()
				// check if livingIPs is empty or not
				if len(livingIPs) == 0 {
					livingIPs = append(livingIPs, "127.0.0.1")
				}
				if pointerElevator.Ip == livingIPs[0] {
					// If I'm the master i should reassign orders of the dead node
					elevatorData.UpdateIsOnline(masterElevator, ActiveIPs, livingIPs)
					ReassignOrders(masterElevator, ActiveIPs, livingIPs)
					communication.BroadcastMessage(nil, masterElevator)
				}
				ActiveIPs = livingIPs
				ActiveIPsMutex.Unlock()
				updateRole(pointerElevator, masterElevator)
			}
		}
	}
}

// Used when the ActiveIPs list is changed
func ReassignOrders(masterElevator *defs.MasterList, oldList []string, newList []string) {
	var counter int
	for _, elevIP := range oldList {
		if !utility.Contains(newList, elevIP) {
			log.Println("Reassigning from: ", elevIP)
			for _, e := range masterElevator.Elevators {
				if e.Ip == elevIP {
					for floorIndex, floorOrders := range e.Orders {
						if floorOrders[elevio.BT_HallUp] {
							floorOrders[elevio.BT_HallUp] = false
							orderAssignment.FindAndAssign(masterElevator, floorIndex, int(elevio.BT_HallUp), elevIP)
							counter++
						}
						if floorOrders[elevio.BT_HallDown] {
							floorOrders[elevio.BT_HallDown] = false
							orderAssignment.FindAndAssign(masterElevator, floorIndex, int(elevio.BT_HallDown), elevIP)
							counter++
						}
					}
				}
			}
		}
	}
	log.Println(counter, " orders reassigned")
}

// Used when elevators still are online, but one or more elevators are inoperative
func ReassignOrders2(masterList *defs.MasterList) {
	operativeElevators := make([]string, 0)
	onlineElevators := make([]string, 0)

	for _, e := range masterList.Elevators {
		if e.IsOnline {
			onlineElevators = append(onlineElevators, e.Ip)
		}
		if e.Status.Operative {
			operativeElevators = append(operativeElevators, e.Ip)
		}
	}

	if (len(onlineElevators) > len(operativeElevators)) && (len(operativeElevators) > 0) {
		ReassignOrders(masterList, onlineElevators, operativeElevators)
	}
}

func updateRole(pointerElevator *defs.Elevator, masterElevator *defs.MasterList) {
	ActiveIPsMutex.Lock()
	defer ActiveIPsMutex.Unlock()

	//Sets the role to master if there is not active IPs (Internet turned off while running)
	if len(ActiveIPs) == 0 {
		pointerElevator.Role = defs.MASTER
		return
	}
	//Finds the lowestIP and sets the ServerIP equal to it
	lowestIP := strings.Split(ActiveIPs[0], ":")[0]
	if ServerIP != lowestIP {
		connected = false
		ServerIP = lowestIP
	}
	//Sets role to master if lowestIP is localhost
	if lowestIP == "127.0.0.1" {
		pointerElevator.Role = defs.MASTER
		return
	}

	if defs.MyIP == lowestIP && !ServerListening {
		//Set role to master and starts a new server
		shutdownServer()
		go startServer(masterElevator) // Ensure server starts in a non-blocking manner
		pointerElevator.Role = defs.MASTER
	} else if defs.MyIP != lowestIP && ServerListening {
		//Stops the server and switches from master to slave role
		shutdownServer()
		ServerActive <- false                                                  // Stop the server
		go connectToServer(lowestIP+":55555", pointerElevator, masterElevator) // Transition to client
		pointerElevator.Role = defs.SLAVE
	} else if !ServerListening {
		//Starts a client connection to the server, and sets role to slave
		if !connected {
			go connectToServer(lowestIP+":55555", pointerElevator, masterElevator)
			pointerElevator.Role = defs.SLAVE
		}
	}
}
