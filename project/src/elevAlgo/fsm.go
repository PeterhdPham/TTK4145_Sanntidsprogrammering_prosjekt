package elevAlgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/broadcast"
	"project/cost"
	"project/elevData"
	"project/utility"
	"project/variable"
)

func FSM_InitBetweenFloors(status elevData.ElevStatus) elevData.ElevStatus {
	elevio.SetMotorDirection(-1)
	status.FSM_State = variable.Moving
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status elevData.ElevStatus, orders [][]bool, floor int) elevData.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	status.Buttonfloor = -1
	status.Buttontype = -1
	switch status.FSM_State {
	case variable.Moving:
		if requestShouldStop(status, orders, floor) {
			//Stops elevator and updates status accordingly
			elevio.SetMotorDirection(elevio.MD_Stop)

			//Opens elevator door and updates status accordingly
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = variable.DoorOpen

			//Clears the request at current floor
			status, orders = requestClearAtFloor(status, orders, floor)

			//Sets the lights according to the current orders
			SetAllLights(orders)
		}
	default:
		break
	}
	return status
}

func FSM_RequestFloor(master *elevData.MasterList, floor int, button int, fromIP string, myRole elevData.ElevatorRole) (elevData.ElevStatus, [][]bool) {

	//Find the best elevator to take the order, update the masterlist and broadcast to all slaves
	if myRole == elevData.Master {
		fmt.Println("I AM MASTER")
		cost.FindAndAssign(master, floor, button, fromIP)
		jsonToSend := utility.MarshalJson(master)
		broadcast.BroadcastMessage(nil, jsonToSend)
	}

	//Check orders and starts moving
	var status elevData.ElevStatus
	var orders [][]bool
	for _, e := range master.Elevators {
		if e.Ip == MyIP {
			status = e.Status
			orders = e.Orders
		}
	}

	switch status.FSM_State {
	case variable.DoorOpen:
		if requestShouldClearImmediately(status, orders, floor, button) {
			orders[floor][button] = false
			SetAllLights(orders)
			timerStop()
			timerStart(doorOpenDuration)
			status.FSM_State = variable.DoorOpen
			status.Doors = true
		}
	case variable.Idle:
		status.FSM_State = variable.Moving
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		status.FSM_State = pair.Behaviour
		if pair.Behaviour == variable.DoorOpen {
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = variable.DoorOpen
		}

	}

	return status, orders
}

func FSM_onDoorTimeout(status elevData.ElevStatus, orders [][]bool, floor int) (elevData.ElevStatus, [][]bool) {

	switch status.FSM_State {
	case variable.DoorOpen:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		status.FSM_State = pair.Behaviour

		switch status.FSM_State {
		case variable.DoorOpen:
			timerStart(doorOpenDuration)
			status, orders = requestClearAtFloor(status, orders, floor)
			SetAllLights(orders)
		case variable.Moving, variable.Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
		}

	default:
		// No action for default case
	}

	return status, orders
}
