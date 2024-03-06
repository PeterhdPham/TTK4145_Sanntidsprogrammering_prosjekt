package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
)

var N_BUTTONS = 3

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour string
}

func areAllOrdersFalse(orders [][]bool) bool {
	for _, floor := range orders {
		for _, order := range floor {
			if order {
				// Found an order (true), so not all values are false.
				return false
			}
		}
	}
	// If we get here, it means all values were false.
	return true
}
func requestShouldStop(status elevData.ElevStatus, orders [][]bool, floor int) bool {
	if areAllOrdersFalse(orders) {
		fmt.Println("No orders found")
		return true
	}
	switch status.Direction {
	case -1:
		if orders[floor][1] || orders[floor][2] {
			return true
		}
	case 1:
		if orders[floor][0] || orders[floor][2] {
			return true
		}
	}

	return false
}

func requestClearAtFloor(myStatus elevData.ElevStatus, myOrders [][]bool, floor int) (elevData.ElevStatus, [][]bool) {
	fmt.Println("Request Clear")

	switch myStatus.Direction {
	case 1:
		if !requestsAbove(myStatus, myOrders, floor) && !myOrders[floor][0] {
			myOrders[floor][1] = false
		}
		myOrders[floor][0] = false

	case -1:
		if !requestsBelow(myStatus, myOrders, floor) && !myOrders[floor][1] {
			myOrders[floor][0] = false
		}
		myOrders[floor][1] = false

	case 0:
		fallthrough
	default:
		myOrders[floor][0] = false
		myOrders[floor][1] = false
	}

	return myStatus, myOrders
}

func requestsAbove(status elevData.ElevStatus, orders [][]bool, floor int) bool {
	for f := status.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(status elevData.ElevStatus, orders [][]bool, floor int) bool {
	for f := 0; f < status.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(status elevData.ElevStatus, orders [][]bool) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if orders[status.Floor][btn] {
			return true
		}
	}
	return false
}

func UpdateLights(master elevData.MasterList, ip string) {
	for _, e := range master.Elevators {
		if e.Ip == ip {
			setAllLights(e.Orders)
		}
	}
}

func setAllLights(orders [][]bool) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := elevio.BT_HallUp; btn <= elevio.BT_Cab; btn++ {
			elevio.SetButtonLamp(btn, floor, orders[floor][btn])
		}
	}
}

func requestsChooseDirection(status elevData.ElevStatus, orders [][]bool) DirnBehaviourPair {
	switch status.Direction {
	case int(elevio.MD_Up):
		fmt.Println("Direction: Up")
		if requestsAbove(status, orders, status.Floor) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: Moving}
		} else if requestsHere(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: DoorOpen}
		} else if requestsBelow(status, orders, status.Floor) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
		}
	case int(elevio.MD_Down):
		fmt.Println("Direction: Down")
		if requestsBelow(status, orders, status.Floor) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: Moving}
		} else if requestsHere(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: DoorOpen}
		} else if requestsAbove(status, orders, status.Floor) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
		}
	case int(elevio.MD_Stop):
		fmt.Println("Direction: Stop")
		if requestsHere(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: DoorOpen}
		} else if requestsAbove(status, orders, status.Floor) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: Moving}
		} else if requestsBelow(status, orders, status.Floor) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
	}
}
