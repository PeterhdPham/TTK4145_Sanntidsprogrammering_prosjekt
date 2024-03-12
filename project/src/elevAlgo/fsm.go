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
	failureTimerStop()
	failureTimerStart(failureTimeoutDuration, int(defs.MOTOR_FAIL))
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status defs.ElevStatus, orders [][]bool, lights [][]bool, floor int) defs.ElevStatus {
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
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(defs.DOOR_STUCK))

			//Clears the request at current floor
			status, lights = requestClearAtFloor(status, orders, lights, floor)

			//Sets the lights according to the current orders
			SetAllLights(lights)
		} else {
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(defs.MOTOR_FAIL))

		}
	default:
		break
	}
	return status
}

func FSM_RequestFloor(master *defs.MasterList, floor int, button int, fromIP string, myRole defs.ElevatorRole) (defs.ElevStatus, [][]bool, [][]bool) {

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
	var lights [][]bool
	for _, e := range master.Elevators {
		if e.Ip == defs.MyIP {
			status = e.Status
			orders = e.Orders
			lights = e.Lights
		}
	}

	switch status.FSM_State {
	case defs.DOOR_OPEN:
		if requestShouldClearImmediately(status, floor, button) {
			orders[floor][button] = false

			SetAllLights(lights)
			timerStop()
			timerStart(doorOpenDuration)
			status.FSM_State = defs.DOOR_OPEN
			status.Doors = true
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(defs.DOOR_STUCK))

		}
	case defs.IDLE:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		status.FSM_State = pair.Behaviour
		if pair.Behaviour == defs.DOOR_OPEN {
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = defs.DOOR_OPEN
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(defs.DOOR_STUCK))
		} else {
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(defs.MOTOR_FAIL))
		}

	}

	return status, orders, lights
}

func FSM_onDoorTimeout(status defs.ElevStatus, orders [][]bool, lights [][]bool, floor int) (defs.ElevStatus, [][]bool) {

	switch status.FSM_State {
	case defs.DOOR_OPEN:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		status.FSM_State = pair.Behaviour

		switch status.FSM_State {
		case defs.DOOR_OPEN:
			timerStart(doorOpenDuration)
			status, lights = requestClearAtFloor(status, orders, lights, floor)
			SetAllLights(lights)
		case defs.MOVING, defs.IDLE:
			elevio.SetDoorOpenLamp(false)
			status.Doors = false
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(defs.MOTOR_FAIL))

		}

	default:
		// No action for default case
	}

	return status, orders
}
