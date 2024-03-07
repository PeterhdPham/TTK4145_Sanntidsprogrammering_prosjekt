package elevalgo

import (
	"Driver-go/elevio"
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
		return true
	}
	switch status.Direction {
	case elevio.MD_Down:
		if orders[floor][1] || orders[floor][2] || !requestsBelow(status, orders) {
			return true
		}
	case int(elevio.MD_Up):
		if orders[floor][0] || orders[floor][2] || !requestsAbove(status, orders) {
			return true
		}
	}

	return false
}

func requestClearAtFloor(myStatus elevData.ElevStatus, myOrders [][]bool, floor int) (elevData.ElevStatus, [][]bool) {
	switch myStatus.Direction {
	case 1:
		if !requestsAbove(myStatus, myOrders) && !myOrders[floor][0] {
			myOrders[floor][1] = false
		}
		myOrders[floor][0] = false

	case -1:
		if !requestsBelow(myStatus, myOrders) && !myOrders[floor][1] {
			myOrders[floor][0] = false
		}
		myOrders[floor][1] = false
	default:
		myOrders[floor][0] = false
		myOrders[floor][1] = false
	}

	myOrders[floor][2] = false

	return myStatus, myOrders
}

func requestsAbove(status elevData.ElevStatus, orders [][]bool) bool {
	for f := status.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(status elevData.ElevStatus, orders [][]bool) bool {
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

func SetAllLights(orders [][]bool) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := elevio.BT_HallUp; btn <= elevio.BT_Cab; btn++ {
			elevio.SetButtonLamp(btn, floor, orders[floor][btn])
		}
	}
}

func requestsChooseDirection(status elevData.ElevStatus, orders [][]bool) DirnBehaviourPair {
	switch status.Direction {
	case int(elevio.MD_Up):
		if requestsAbove(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: Moving}
		} else if requestsHere(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: DoorOpen}
		} else if requestsBelow(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
		}
	case int(elevio.MD_Down):
		if requestsBelow(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: Moving}
		} else if requestsHere(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: DoorOpen}
		} else if requestsAbove(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
		}
	case int(elevio.MD_Stop):
		if requestsHere(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: DoorOpen}
		} else if requestsAbove(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: Moving}
		} else if requestsBelow(status, orders) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: Idle}
	}
}
