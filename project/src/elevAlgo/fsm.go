package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
)

var FSM_State string

const (
	Idle     = "EB_Idle"
	Moving   = "EB_Moving"
	DoorOpen = "EB_DoorOpen"
)

func FSM_InitBetweenFloors(status elevData.ElevStatus) elevData.ElevStatus {
	elevio.SetMotorDirection(-1)
	FSM_State = Moving
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status elevData.ElevStatus, orders [][]bool, floor int) elevData.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	status.Buttonfloor = -1
	status.Buttontype = -1
	switch FSM_State {
	case Moving:
		if requestShouldStop(status, orders, floor) {
			//Stops elevator and updates status accordingly
			elevio.SetMotorDirection(elevio.MD_Stop)
			status.Direction = 0

			//Opens elevator door and updates status accordingly
			elevio.SetDoorOpenLamp(true)
			status.Doors = true

			//Clears the request at current floor
			status, orders = requestClearAtFloor(status, orders, floor)

			timerStart(doorOpenDuration)

			setAllLights(orders)
			FSM_State = DoorOpen
		}
	default:
		break
	}
	return status
}

func FSM_RequestFloor(master *elevData.MasterList, floor int, button int) {

}

func FSM_onDoorTimeout(status elevData.ElevStatus, orders [][]bool, floor int) (elevData.ElevStatus, [][]bool) {
	fmt.Printf("\n\n%s()\n", "FSM_OnDoorTimeout")

	switch FSM_State {
	case DoorOpen:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		FSM_State = pair.Behaviour

		switch FSM_State {
		case DoorOpen:
			timerStart(doorOpenDuration)
			status, orders = requestClearAtFloor(status, orders, floor)
			setAllLights(orders)
		case Moving, Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
		}

	default:
		// No action for default case
	}

	return status, orders
}
