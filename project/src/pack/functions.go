package pack

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func MoveTowardsFloor(floor_order int) {
	fmt.Printf("FUnc")
	var current_floor = elevio.GetFloor()
	fmt.Printf("Current Floor: %d", current_floor)
	if floor_order == current_floor {
		elevio.SetMotorDirection(elevio.MD_Stop)
		fmt.Println("Stop")
	} else if floor_order < current_floor {
		elevio.SetMotorDirection(elevio.MD_Down)
		fmt.Println("Down")
	} else if floor_order > current_floor {
		elevio.SetMotorDirection(elevio.MD_Up)
		fmt.Println("Up")
	}
}

func TurnOffLight(currentFloor int, dir elevio.MotorDirection) {
	if currentFloor == elevio.GetFloor() {
		if dir == 1 {
			elevio.SetButtonLamp(0, currentFloor, false)
		} else if dir == -1 {
			elevio.SetButtonLamp(1, currentFloor, false)
		}
	}
}

func OpenDoor(dir elevio.MotorDirection) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	fmt.Println("Door opening...")
	elevio.SetDoorOpenLamp(true)
	time.Sleep(2000 * time.Millisecond)
	fmt.Println("Door closing...")
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(dir)
}

func CheckHallCall(dir elevio.MotorDirection, floor int) {
	if dir == -1 {
		if elevio.GetButton(1, floor) {
			OpenDoor(dir)
		}
	} else if dir == 1 {
		if elevio.GetButton(0, floor) {
			OpenDoor(dir)
		}
	}
}
