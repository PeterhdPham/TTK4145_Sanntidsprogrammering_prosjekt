package movement

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"time"
)

const POLLING_INTERVAL = 100 * time.Millisecond

func FulfillRequests(
	requests chan []int,
	elevStatusChan chan elevData.ElevStatus,
    direction chan elevio.MotorDirection,
    doorOpen chan bool,
){
	for newRequests := range requests {
		fmt.Println("Servicing new request:", newRequests[0])
		currentStatus := <-elevStatusChan
		currentFloor := currentStatus.Floor

		if currentFloor == newRequests[0]{
			doorOpen <- true
			doorCycle(doorOpen)
			doorOpen <- false
		}else{
			if newRequests[0] > currentFloor{
				fmt.Println("UP")
				elevio.SetMotorDirection(elevio.MD_Up)
				direction <- elevio.MD_Down
				for currentFloor != newRequests[0]{
					time.Sleep(time.Millisecond*100)
					currentStatus = <-elevStatusChan
					currentFloor = currentStatus.Floor
				}
				elevio.SetMotorDirection(elevio.MD_Stop)
				direction <- elevio.MD_Stop
				doorOpen <- true
				doorCycle(doorOpen)
				doorOpen <- false
			} else{
				fmt.Println("DOWN")
				elevio.SetMotorDirection(elevio.MD_Down)
                direction <- elevio.MD_Down
                for currentFloor!= newRequests[0]{
                    time.Sleep(time.Millisecond*100)
                    currentStatus = <-elevStatusChan
                    currentFloor = currentStatus.Floor
                }
                elevio.SetMotorDirection(elevio.MD_Stop)
                direction <- elevio.MD_Stop
				doorOpen <- true
				doorCycle(doorOpen)
				doorOpen <- false
			}
		}
	fmt.Println("Finished request")
	}
}



func doorCycle(doorOpenChan chan bool){
	fmt.Println("Opening door...")
	// doorOpenChan <- true // FIX THIS
	elevio.SetDoorOpenLamp(true)
	time.Sleep(time.Second*2)
	fmt.Println("Closing door...")
	// doorOpenChan <- false // FIX THIS
	fmt.Println("here")
	elevio.SetDoorOpenLamp(false)
	fmt.Println("door closed")
}