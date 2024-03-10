package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/cost"
	"project/elevData"
	"project/tcp"
	"project/utility"
)

const (
	Idle     = "EB_Idle"
	Moving   = "EB_Moving"
	DoorOpen = "EB_DoorOpen"
)

func FSM_InitBetweenFloors(status elevData.ElevStatus) elevData.ElevStatus {
	elevio.SetMotorDirection(-1)
	status.FSM_State = Moving
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status elevData.ElevStatus, orders [][]bool, floor int) elevData.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	status.Buttonfloor = -1
	status.Buttontype = -1
	switch status.FSM_State {
	case Moving:
		if requestShouldStop(status, orders, floor) {
			//Stops elevator and updates status accordingly
			elevio.SetMotorDirection(elevio.MD_Stop)

			//Opens elevator door and updates status accordingly
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = DoorOpen
			failureTimerStart(failureTimeoutDuration, int(DoorStuck))

			//Clears the request at current floor
			status, orders = requestClearAtFloor(status, orders, floor)

			//Sets the lights according to the current orders
			SetAllLights(orders)
		} else {
			failureTimerStart(failureTimeoutDuration, int(MotorFail))
		}
	default:
		break
	}
	status.Operative = true
	return status
}

func FSM_RequestFloor(master *elevData.MasterList, floor int, button int, fromIP string, myRole elevData.ElevatorRole) (elevData.ElevStatus, [][]bool) {

	//Find the best elevator to take the order, update the masterlist and broadcast to all slaves
	if myRole == elevData.Master {
		// fmt.Println("IM MASTER")
		cost.FindAndAssign(master, floor, button, fromIP)
		jsonToSend := utility.MarshalJson(master)
		fmt.Println("Broadcasting master")
		tcp.BroadcastMessage(nil, jsonToSend)
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
	case DoorOpen:
		if requestShouldClearImmediately(status, orders, floor, button) {
			orders[floor][button] = false
			timerStop()
			timerStart(doorOpenDuration)
			status.FSM_State = DoorOpen
			status.Doors = true
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(DoorStuck))
		}
	case Idle:
		status.FSM_State = Moving
		failureTimerStop()
		failureTimerStart(failureTimeoutDuration, int(MotorFail))
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		status.FSM_State = pair.Behaviour
		if pair.Behaviour == DoorOpen {
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			timerStart(doorOpenDuration)
			status.FSM_State = DoorOpen
		}

	}

	return status, orders
}

func FSM_onDoorTimeout(status elevData.ElevStatus, orders [][]bool, floor int) (elevData.ElevStatus, [][]bool) {

	switch status.FSM_State {
	case DoorOpen:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		status.FSM_State = pair.Behaviour

		switch status.FSM_State {
		case DoorOpen:
			timerStart(doorOpenDuration)
			status, orders = requestClearAtFloor(status, orders, floor)
			SetAllLights(orders)
		case Moving, Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
		}

	default:
		// No action for default case
	}

	return status, orders
}
