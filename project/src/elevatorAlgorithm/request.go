package elevatorAlgorithm

import (
	"Driver-go/elevio"
	"project/constants"
	"project/types"
)

func allOrdersFalse(myOrders [][]bool) bool {
	for _, floor := range myOrders {
		for _, order := range floor {
			if order {
				return false
			}
		}
	}
	return true
}

func requestShouldStop(myStatus types.ElevStatus, myOrders [][]bool, floor int) bool {
	if allOrdersFalse(myOrders) {
		return true
	}
	switch myStatus.Direction {
	case elevio.MD_Down:
		if myOrders[floor][elevio.BT_HallDown] || myOrders[floor][elevio.BT_Cab] || !requestsBelow(myStatus, myOrders) {
			return true
		}
	case int(elevio.MD_Up):
		if myOrders[floor][elevio.BT_HallUp] || myOrders[floor][elevio.BT_Cab] || !requestsAbove(myStatus, myOrders) {
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

func requestsAbove(myStatus types.ElevStatus, myOrders [][]bool) bool {
	for f := myStatus.Floor + 1; f < constants.N_FLOORS; f++ {
		for btn := 0; btn < constants.N_BUTTONS; btn++ {
			if myOrders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(myStatus types.ElevStatus, myOrders [][]bool) bool {
	for f := 0; f < myStatus.Floor; f++ {
		for btn := 0; btn < constants.N_BUTTONS; btn++ {
			if myOrders[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(myStatus types.ElevStatus, myOrders [][]bool) bool {
	for btn := 0; btn < constants.N_BUTTONS; btn++ {
		if myOrders[myStatus.Floor][btn] {
			return true
		}
	}
	return false
}

func requestsChooseDirection(myStatus types.ElevStatus, myOrders [][]bool) types.DirnBehaviourPair {
	switch myStatus.Direction {
	case int(elevio.MD_Up):
		if requestsAbove(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.MOVING}
		} else if requestsHere(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.DOOR_OPEN}
		} else if requestsBelow(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.MOVING}
		} else {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
		}
	case int(elevio.MD_Down):
		if requestsBelow(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.MOVING}
		} else if requestsHere(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.DOOR_OPEN}
		} else if requestsAbove(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.MOVING}
		} else {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
		}
	case int(elevio.MD_Stop):
		if requestsHere(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.DOOR_OPEN}
		} else if requestsAbove(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: constants.MOVING}
		} else if requestsBelow(myStatus, myOrders) {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: constants.MOVING}
		} else {
			return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
		}
	default:
		return types.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: constants.IDLE}
	}
}
