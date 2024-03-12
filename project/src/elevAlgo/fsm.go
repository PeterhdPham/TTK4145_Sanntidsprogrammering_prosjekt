package elevAlgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/broadcast"
	"project/cost"
	"project/defs"
	"project/utility"
)

func FSM_InitBetweenFloors(status defs.ElevStatus) defs.ElevStatus {
	elevio.SetMotorDirection(-1)
	status.FSM_State = defs.MOVING
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status defs.ElevStatus, orders [][]bool, floor int) defs.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	status.Buttonfloor = -1
	status.Buttontype = -1
	switch status.FSM_State {
	case defs.MOVING:
		if requestShouldStop(status, orders, floor) {
			//Stops elevator and updates status accordingly
			elevio.SetMotorDirection(elevio.MD_Stop)

			//Opens elevator door and updates status accordingly
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = defs.DOOR_OPEN

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

func FSM_RequestFloor(master *defs.MasterList, floor int, button int, fromIP string, myRole defs.ElevatorRole) (defs.ElevStatus, [][]bool) {

	//Find the best elevator to take the order, update the masterlist and broadcast to all slaves
	if myRole == defs.MASTER {
		fmt.Println("I AM MASTER")
		cost.FindAndAssign(master, floor, button, fromIP)
		jsonToSend := utility.MarshalJson(master)
		broadcast.BroadcastMessage(nil, jsonToSend)
	}

	//Check orders and starts moving
	var status defs.ElevStatus
	var orders [][]bool
	for _, e := range master.Elevators {
		if e.Ip == defs.MyIP {
			status = e.Status
			orders = e.Orders
			fmt.Println(orders)
		}
	}

	switch status.FSM_State {
	case defs.DOOR_OPEN:
		if requestShouldClearImmediately(status, orders, floor, button) {
			orders[floor][button] = false
			SetAllLights(orders)
			timerStop()
			timerStart(doorOpenDuration)
			status.FSM_State = defs.DOOR_OPEN
			status.Doors = true
		}
	case defs.IDLE:
		status.FSM_State = defs.MOVING
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		status.FSM_State = pair.Behaviour
		if pair.Behaviour == defs.DOOR_OPEN {
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = defs.DOOR_OPEN
		}

	}

	return status, orders
}

func FSM_onDoorTimeout(status defs.ElevStatus, orders [][]bool, floor int) (defs.ElevStatus, [][]bool) {

	switch status.FSM_State {
	case defs.DOOR_OPEN:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		status.FSM_State = pair.Behaviour

		switch status.FSM_State {
		case defs.DOOR_OPEN:
			timerStart(doorOpenDuration)
			status, orders = requestClearAtFloor(status, orders, floor)
			SetAllLights(orders)
		case defs.MOVING, defs.IDLE:
			elevio.SetDoorOpenLamp(false)
			status.Doors = false
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
		}

	default:
		// No action for default case
	}

	return status, orders
}
