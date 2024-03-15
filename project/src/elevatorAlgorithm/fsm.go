package elevatorAlgorithm

import (
	"Driver-go/elevio"
	"project/communication"
	"project/constants"
	"project/elevatorData"
	"project/orderAssignment"
	"project/types"
	"project/variables"
)

const DOOR_STUCK types.FailureMode = 0
const MOTOR_FAIL types.FailureMode = 1

func FSM_InitBetweenFloors(status types.ElevStatus) types.ElevStatus {
	elevio.SetMotorDirection(-1)
	status.FSM_State = constants.MOVING
	failureTimerStop()
	failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status types.ElevStatus, orders [][]bool, floor int) (types.ElevStatus, [][]bool) {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	status.Buttonfloor = -1
	status.Buttontype = -1
	switch status.FSM_State {
	case constants.MOVING:
		if requestShouldStop(status, orders, floor) {
			//Stops elevator and updates status accordingly
			elevio.SetMotorDirection(elevio.MD_Stop)

			//Opens elevator door and updates status accordingly
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			doorTimerStart(doorOpenDuration)
			status.FSM_State = constants.DOOR_OPEN
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(DOOR_STUCK))

			//Clears the request at current floor
			status, orders = requestClearAtFloor(status, orders, floor)
		} else {
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))

		}
	default:
		break
	}
	return status, orders
}

func FSM_RequestFloor(master *types.MasterList, status types.ElevStatus, orders [][]bool, floor int, button int, fromIP string, myRole types.ElevatorRole) (types.ElevStatus, [][]bool) {

	//Find the best elevator to take the order, update the masterlist and broadcast to all slaves
	if myRole == constants.MASTER {
		orderAssignment.FindAndAssign(master, floor, button, fromIP)
		elevatorData.UpdateLightsMasterList(master, variables.MyIP)
		communication.BroadcastMessage(nil, master)
	}
	elevatorData.SetAllLights(*master)

	//Check orders and starts moving
	for _, e := range master.Elevators {
		if e.Ip == variables.MyIP {
			orders = e.Orders
		}
	}
	switch status.FSM_State {
	case constants.DOOR_OPEN:
		if requestShouldClearImmediately(status, floor, button) {
			orders[floor][button] = false
			doorTimerStop()
			doorTimerStart(doorOpenDuration)
			status.FSM_State = constants.DOOR_OPEN
			status.Doors = true
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(DOOR_STUCK))

		}
	case constants.IDLE:
		floor = elevio.GetFloor()
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		status.FSM_State = pair.Behaviour
		if pair.Behaviour == constants.DOOR_OPEN {
			status, orders = requestClearAtFloor(status, orders, floor)
			elevio.SetDoorOpenLamp(true)
			status.Doors = true
			doorTimerStart(doorOpenDuration)
			status.FSM_State = constants.DOOR_OPEN
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(DOOR_STUCK))
		} else {
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))
		}

	}

	return status, orders
}

func FSM_onDoorTimeout(status types.ElevStatus, orders [][]bool, floor int) (types.ElevStatus, [][]bool) {

	switch status.FSM_State {
	case constants.DOOR_OPEN:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		status.FSM_State = pair.Behaviour

		switch status.FSM_State {
		case constants.DOOR_OPEN:
			doorTimerStart(doorOpenDuration)
			status, orders = requestClearAtFloor(status, orders, floor)
		case constants.MOVING, constants.IDLE:
			elevio.SetDoorOpenLamp(false)
			status.Doors = false
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))

		}

	default:
		// No action for default case
	}

	return status, orders
}
