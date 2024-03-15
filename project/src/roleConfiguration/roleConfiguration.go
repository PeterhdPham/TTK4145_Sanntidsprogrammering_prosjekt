package roleConfiguration

import (
	"Driver-go/elevio"
	"project/aliveMessages"
	"project/communication"
	"project/constants"
	"project/elevatorData"
	"project/orderAssignment"
	"project/types"
	"project/utility"
	"project/variables"
	"strings"
	"time"
)

const SERVERPORT = ":55555"
const DELAY_INIT_ROLES = 3 * time.Second

var ServerListening bool = false
var ServerIP string

func ConfigureRoles(pointerElevator *types.Elevator, masterElevator *types.MasterList) {

	go aliveMessages.BroadcastLife()
	go aliveMessages.LookForLife(LivingIPsChan)

	time.Sleep(DELAY_INIT_ROLES)

	for {
		select {
		case livingIPs := <-LivingIPsChan:

			if !utility.StringSlicesAreEqual(ActiveIPs, livingIPs) {
				ActiveIPsMutex.Lock()

				if len(livingIPs) == 0 {
					livingIPs = append(livingIPs, "127.0.0.1")
				}
				if pointerElevator.Ip == livingIPs[0] {

					elevatorData.UpdateIsOnline(masterElevator, ActiveIPs, livingIPs)
					ReassignOrders(masterElevator, ActiveIPs, livingIPs)
					communication.BroadcastMessage(masterElevator)
				}
				ActiveIPs = livingIPs
				ActiveIPsMutex.Unlock()
				updateRoles(pointerElevator, masterElevator)
			}
		}
	}
}

func ReassignOrders(masterElevator *types.MasterList, oldList []string, newList []string) {
	var counter int
	for _, elevIP := range oldList {
		if !utility.Contains(newList, elevIP) {
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
}

func ReassignOrdersIfInoperative(masterList *types.MasterList) {
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

func updateRoles(pointerElevator *types.Elevator, masterElevator *types.MasterList) {
	ActiveIPsMutex.Lock()
	defer ActiveIPsMutex.Unlock()

	if len(ActiveIPs) == 0 {
		pointerElevator.Role = constants.MASTER
		return
	}

	lowestIP := strings.Split(ActiveIPs[0], ":")[0]
	if ServerIP != lowestIP {
		connected = false
		ServerIP = lowestIP
	}

	if lowestIP == "127.0.0.1" {
		pointerElevator.Role = constants.MASTER
		return
	}

	if variables.MyIP == lowestIP && !ServerListening {

		shutdownServer()
		go startServer(masterElevator)
		pointerElevator.Role = constants.MASTER
	} else if variables.MyIP != lowestIP && ServerListening {

		shutdownServer()
		ServerActive <- false
		go connectToServer(lowestIP+SERVERPORT, pointerElevator, masterElevator)
		pointerElevator.Role = constants.SLAVE
	} else if !ServerListening {

		if !connected {
			go connectToServer(lowestIP+SERVERPORT, pointerElevator, masterElevator)
			pointerElevator.Role = constants.SLAVE
		}
	}
}
