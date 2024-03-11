package elevAlgo

import (
	"Driver-go/elevio"
	"project/variable"
)

var N_BUTTONS = 3

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

func requestShouldStop(status variable.ElevStatus, orders [][]bool, floor int) bool {
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

func requestClearAtFloor(myStatus variable.ElevStatus, myOrders [][]bool, floor int) (variable.ElevStatus, [][]bool) {
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

func requestShouldClearImmediately(myStatus variable.ElevStatus, myOrders [][]bool, floor int, btn int) bool {
	return myStatus.Floor == floor && ((myStatus.Direction == int(elevio.MD_Up) && btn == int(elevio.BT_HallUp)) ||
		(myStatus.Direction == int(elevio.MD_Down) && btn == int(elevio.BT_HallDown)) ||
		(myStatus.Direction == int(elevio.MD_Stop) && btn == int(elevio.BT_Cab)))
}

func requestsAbove(status variable.ElevStatus, orders [][]bool) bool {
	for f := status.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(status variable.ElevStatus, orders [][]bool) bool {
	for f := 0; f < status.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(status variable.ElevStatus, orders [][]bool) bool {
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

func requestsChooseDirection(status variable.ElevStatus, orders [][]bool) variable.DirnBehaviourPair {
	switch status.Direction {
	case int(elevio.MD_Up):
		if requestsAbove(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: variable.MOVING}
		} else if requestsHere(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: variable.DOOR_OPEN}
		} else if requestsBelow(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: variable.MOVING}
		} else {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: variable.IDLE}
		}
	case int(elevio.MD_Down):
		if requestsBelow(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: variable.MOVING}
		} else if requestsHere(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: variable.DOOR_OPEN}
		} else if requestsAbove(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: variable.MOVING}
		} else {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: variable.IDLE}
		}
	case int(elevio.MD_Stop):
		if requestsHere(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: variable.DOOR_OPEN}
		} else if requestsAbove(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: variable.MOVING}
		} else if requestsBelow(status, orders) {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: variable.MOVING}
		} else {
			return variable.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: variable.IDLE}
		}
	default:
		return variable.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: variable.IDLE}
	}
}
