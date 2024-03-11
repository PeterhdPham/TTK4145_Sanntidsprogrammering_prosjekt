package elevAlgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/broadcast"
	"project/cost"
	"project/utility"
	"project/variable"
)

func FSM_InitBetweenFloors(status variable.ElevStatus) variable.ElevStatus {
	elevio.SetMotorDirection(-1)
	status.FSM_State = variable.MOVING
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status variable.ElevStatus, orders [][]bool, floor int) variable.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	status.Buttonfloor = -1
	status.Buttontype = -1
	switch status.FSM_State {
	case variable.MOVING:
		if requestShouldStop(status, orders, floor) {
			//Stops elevator and updates status accordingly
			elevio.SetMotorDirection(elevio.MD_Stop)

			//Opens elevator door and updates status accordingly
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = variable.DOOR_OPEN

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

func FSM_RequestFloor(master *variable.MasterList, floor int, button int, fromIP string, myRole variable.ElevatorRole) (variable.ElevStatus, [][]bool) {

	//Find the best elevator to take the order, update the masterlist and broadcast to all slaves
	if myRole == variable.MASTER {
		fmt.Println("I AM MASTER")
		cost.FindAndAssign(master, floor, button, fromIP)
		jsonToSend := utility.MarshalJson(master)
		broadcast.BroadcastMessage(nil, jsonToSend)
	}

	//Check orders and starts moving
	var status variable.ElevStatus
	var orders [][]bool
	for _, e := range master.Elevators {
		if e.Ip == variable.MyIP {
			status = e.Status
			orders = e.Orders
		}
	}

	switch status.FSM_State {
	case variable.DOOR_OPEN:
		if requestShouldClearImmediately(status, orders, floor, button) {
			orders[floor][button] = false
			SetAllLights(orders)
			timerStop()
			timerStart(doorOpenDuration)
			status.FSM_State = variable.DOOR_OPEN
			status.Doors = true
		}
	case variable.IDLE:
		status.FSM_State = variable.MOVING
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		status.FSM_State = pair.Behaviour
		if pair.Behaviour == variable.DOOR_OPEN {
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = variable.DOOR_OPEN
		}

	}

	return status, orders
}

func FSM_onDoorTimeout(status variable.ElevStatus, orders [][]bool, floor int) (variable.ElevStatus, [][]bool) {

	switch status.FSM_State {
	case variable.DOOR_OPEN:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		status.FSM_State = pair.Behaviour

		switch status.FSM_State {
		case variable.DOOR_OPEN:
			timerStart(doorOpenDuration)
			status, orders = requestClearAtFloor(status, orders, floor)
			SetAllLights(orders)
		case variable.MOVING, variable.IDLE:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
		}

	default:
		// No action for default case
	}

	return status, orders
}
