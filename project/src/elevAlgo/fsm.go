package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
)

var FSM_State string

const (
	Idle    = "EB_Idle"
	Moving  = "EB_Moving"
	Stopped = "EB_Stopped"
)

func FSM_InitBetweenFloors(status elevData.ElevStatus) elevData.ElevStatus {
	elevio.SetMotorDirection(-1)
	FSM_State = Moving
	status.Direction = -1

	return status
}

func FMS_ArrivalAtFloor(status elevData.ElevStatus, orders [][]bool, floor int) elevData.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	switch status.Direction {
	case 0:
		fmt.Println("Is stopped")
		break
	default:
		fmt.Println("Is moving")
		if requestShouldStop(status, orders, floor) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			status.Direction = 0
			status.Doors = true
		}
	}
	return status
}
