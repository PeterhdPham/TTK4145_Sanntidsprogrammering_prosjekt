package movement

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"time"
)

const POLLING_INTERVAL = 100 * time.Millisecond

func FulfillRequest(
	requests []int,
	elevStatusChan chan elevData.ElevStatus,
	direction chan elevio.MotorDirection,
	doorOpen chan bool,
) {
	if len(requests) == 0 {
		fmt.Println("No requests")
		return
	}

	fmt.Println("Starting request:", requests[0])
	var currentStatus elevData.ElevStatus
	go keepStatusUpdated(elevStatusChan, &currentStatus)

	requestedFloor := requests[0]
	temp := <-elevStatusChan
	currentFloor := temp.Floor

	if currentFloor == requestedFloor {
		fmt.Println("Already at requested floor:", currentFloor)
	} else {
		moveElevator(requestedFloor, &currentStatus, direction)
	}

	doorCycle(&currentStatus)
	fmt.Println("Request finished")
}

func moveElevator(requestedFloor int, currentStatus *elevData.ElevStatus, direction chan elevio.MotorDirection) {
	currentFloor := currentStatus.Floor
	moveDirection := determineDirection(currentFloor, requestedFloor)

	fmt.Println("Moving", moveDirection)
	elevio.SetMotorDirection(moveDirection)
	direction <- moveDirection

	for currentFloor != requestedFloor {
		time.Sleep(POLLING_INTERVAL)
		currentFloor = currentStatus.Floor
	}

	elevio.SetMotorDirection(elevio.MD_Stop)
	direction <- elevio.MD_Stop
}

func determineDirection(currentFloor, requestedFloor int) elevio.MotorDirection {
	if currentFloor < requestedFloor {
		return elevio.MD_Up
	}
	return elevio.MD_Down
}

func keepStatusUpdated(channel <-chan elevData.ElevStatus, status *elevData.ElevStatus) {
	for newStatus := range channel {
		*status = newStatus
		fmt.Println("keepstatusupdated run:", *status)
	}
}

func doorCycle(currentStatus *elevData.ElevStatus) {
	fmt.Println("Opening doors")
	elevio.SetDoorOpenLamp(true)
	currentStatus.Doors = true
	time.Sleep(time.Second * 2)
	for currentStatus.Obstructed {
		time.Sleep(POLLING_INTERVAL)
	}
	elevio.SetDoorOpenLamp(false)
	currentStatus.Doors = false
	fmt.Println("Closing doors")
}
