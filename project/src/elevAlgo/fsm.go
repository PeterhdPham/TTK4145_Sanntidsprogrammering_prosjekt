package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
)

func FMS_ArrivalAtFloor(status elevData.ElevStatus, orders [][]bool, floor int) elevData.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	switch status.Direction {
	case 0:
		fmt.Println("Is stopped")
		break
	default:
		if requestShouldStop(status, orders, floor) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			status.Direction = 0
			status.Doors = true
		}
	}
	return status
}
