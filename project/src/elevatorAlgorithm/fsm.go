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

func initBetweenFloors(myStatus types.ElevStatus) types.ElevStatus {
	elevio.SetMotorDirection(-1)
	myStatus.FSM_State = constants.MOVING
	failureTimerStop()
	failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))
	myStatus.Direction = -1

	return myStatus
}

func arrivalAtFloor(myStatus types.ElevStatus, myOrders [][]bool, floor int) (types.ElevStatus, [][]bool) {
	elevio.SetFloorIndicator(floor)
	myStatus.Floor = floor
	myStatus.Buttonfloor = -1
	myStatus.Buttontype = -1
	switch myStatus.FSM_State {
	case constants.MOVING:
		if requestShouldStop(myStatus, myOrders, floor) {

			elevio.SetMotorDirection(elevio.MD_Stop)

			elevio.SetDoorOpenLamp(true)
			myStatus.Doors = true
			doorTimerStart(doorOpenDuration)
			myStatus.FSM_State = constants.DOOR_OPEN
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(DOOR_STUCK))

			myStatus, myOrders = requestClearAtFloor(myStatus, myOrders, floor)
		} else {
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))

		}
	default:
		break
	}
	return myStatus, myOrders
}

func floorRequested(master *types.MasterList, myStatus types.ElevStatus, myOrders [][]bool, floor int, button int, fromIP string, myRole types.ElevatorRole) (types.ElevStatus, [][]bool) {

	if myRole == constants.MASTER {
		orderAssignment.FindAndAssign(master, floor, button, fromIP)
		elevatorData.UpdateLightsMasterList(master, variables.MyIP)
		communication.BroadcastMessage(master)
	}
	elevatorData.SetAllLights(*master)

	for _, e := range master.Elevators {
		if e.Ip == variables.MyIP {
			myOrders = e.Orders
		}
	}
	switch myStatus.FSM_State {
	case constants.DOOR_OPEN:
		if requestShouldClearImmediately(myStatus, floor, button) {
			myOrders[floor][button] = false
			doorTimerStop()
			doorTimerStart(doorOpenDuration)
			myStatus.FSM_State = constants.DOOR_OPEN
			myStatus.Doors = true
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(DOOR_STUCK))

		}
	case constants.IDLE:
		floor = elevio.GetFloor()
		pair := requestsChooseDirection(myStatus, myOrders)
		myStatus.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		myStatus.FSM_State = pair.Behaviour
		if pair.Behaviour == constants.DOOR_OPEN {
			myStatus, myOrders = requestClearAtFloor(myStatus, myOrders, floor)
			elevio.SetDoorOpenLamp(true)
			myStatus.Doors = true
			doorTimerStart(doorOpenDuration)
			myStatus.FSM_State = constants.DOOR_OPEN
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(DOOR_STUCK))
		} else {
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))
		}
	}
	return myStatus, myOrders
}

func onDoorTimeout(myStatus types.ElevStatus, myOrders [][]bool, floor int) (types.ElevStatus, [][]bool) {

	switch myStatus.FSM_State {
	case constants.DOOR_OPEN:
		pair := requestsChooseDirection(myStatus, myOrders)
		myStatus.Direction = int(pair.Dirn)
		myStatus.FSM_State = pair.Behaviour

		switch myStatus.FSM_State {
		case constants.DOOR_OPEN:
			doorTimerStart(doorOpenDuration)
			myStatus, myOrders = requestClearAtFloor(myStatus, myOrders, floor)
		case constants.MOVING, constants.IDLE:
			elevio.SetDoorOpenLamp(false)
			myStatus.Doors = false
			elevio.SetMotorDirection(elevio.MotorDirection(myStatus.Direction))
			failureTimerStop()
			failureTimerStart(failureTimeoutDuration, int(MOTOR_FAIL))
		}
	}

	return myStatus, myOrders
}
