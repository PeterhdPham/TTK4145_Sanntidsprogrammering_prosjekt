package elevatorAlgorithm

import (
	"Driver-go/elevio"
	"project/constants"
	"project/types"
)

func allOrdersFalse(orders [][]bool) bool {
	for _, floor := range orders {
		for _, order := range floor {
			if order {
				return false
			}
		}
	}
	return true
}

func requestShouldStop(status types.ElevStatus, orders [][]bool, floor int) bool {
	if allOrdersFalse(orders) {
		return true
	}
	switch status.Direction {
	case elevio.MD_Down:
		if orders[floor][elevio.BT_HallDown] || orders[floor][elevio.BT_Cab] || !requestsBelow(status, orders) {
			return true
		}
	case int(elevio.MD_Up):
		if orders[floor][elevio.BT_HallUp] || orders[floor][elevio.BT_Cab] || !requestsAbove(status, orders) {
			return true
		}
	}

	return false
}

func requestClearAtFloor(myStatus types.ElevStatus, myOrders [][]bool, floor int) (types.ElevStatus, [][]bool) {
	switch myStatus.Direction {
	case int(elevio.MD_Up):
		if !requestsAbove(myStatus, myOrders) && !myOrders[floor][elevio.BT_HallUp] {
			myOrders[floor][elevio.BT_HallDown] = false
		}
		myOrders[floor][elevio.BT_HallUp] = false

	case int(elevio.MD_Down):
		if !requestsBelow(myStatus, myOrders) && !myOrders[floor][elevio.BT_HallDown] {
			myOrders[floor][elevio.BT_HallUp] = false
		}
		myOrders[floor][elevio.BT_HallDown] = false
	default:
		myOrders[floor][elevio.BT_HallUp] = false
		myOrders[floor][elevio.BT_HallDown] = false

	}

	myOrders[floor][elevio.BT_Cab] = false

	return myStatus, myOrders
}

func requestShouldClearImmediately(myStatus types.ElevStatus, floor int, btn int) bool {
	return myStatus.Floor == floor && ((myStatus.Direction == int(elevio.MD_Up) && btn == int(elevio.BT_HallUp)) ||
		(myStatus.Direction == int(elevio.MD_Down) && btn == int(elevio.BT_HallDown)) ||
		myStatus.Direction == int(elevio.MD_Stop) ||
		btn == int(elevio.BT_Cab))
}

func requestsAbove(status types.ElevStatus, orders [][]bool) bool {
	for f := status.Floor + 1; f < constants.N_FLOORS; f++ {
		for btn := 0; btn < constants.N_BUTTONS; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(status types.ElevStatus, orders [][]bool) bool {
	for f := 0; f < status.Floor; f++ {
		for btn := 0; btn < constants.N_BUTTONS; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(status types.ElevStatus, orders [][]bool) bool {
	for btn := 0; btn < constants.N_BUTTONS; btn++ {
		if orders[status.Floor][btn] {
			return true
		}
	}
	return false
}

func requestsChooseDirection(status types.ElevStatus, orders [][]bool) types.DirnBehaviourPair {
	switch status.Direction {
	case int(elevio.MD_Up):
		if requestsAbove(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.MOVING}
		} else if requestsHere(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.DOOR_OPEN}
		} else if requestsBelow(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.MOVING}
		} else {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
		}
	case int(elevio.MD_Down):
		if requestsBelow(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.MOVING}
		} else if requestsHere(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.DOOR_OPEN}
		} else if requestsAbove(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.MOVING}
		} else {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
		}
	case int(elevio.MD_Stop):
		if requestsHere(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.DOOR_OPEN}
		} else if requestsAbove(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.MOVING}
		} else if requestsBelow(status, orders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.MOVING}
		} else {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
		}
	default:
		return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
	}
}
