package elevalgo

import (
	"Driver-go/elevio"
	"project/elevData"
	"project/tcp"
	"time"
)

var N_FLOORS int
var doorOpenDuration time.Duration = 3 * time.Second
var MyIP string

func ElevAlgo(masterList *elevData.MasterList, elevStatus chan elevData.ElevStatus, orders chan [][]bool, init_order [][]bool, role elevData.ElevatorRole, N_Floors int) {
	var myStatus elevData.ElevStatus
	myOrders := init_order
	N_FLOORS = N_Floors
	MyIP, _ = tcp.GetPrimaryIP()

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)

	// Moves the elevator down if in between floors

	if elevio.GetFloor() == -1 {
		myStatus = FSM_InitBetweenFloors(myStatus)
	} else {
		FSM_State = Idle
	}

	for {
		select {
		case a := <-drvButtons:
			if role == elevData.Master {
				myStatus, myOrders = FSM_RequestFloor(masterList, a.Floor, int(a.Button), MyIP, role)
			}
			myStatus.Buttonfloor = a.Floor
			myStatus.Buttontype = int(a.Button)
		case a := <-drvFloors:
			myStatus = FSM_ArrivalAtFloor(myStatus, myOrders, a)
		case a := <-drvObstr:
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			if a {
				myStatus.Obstructed = true
			} else {
				myStatus.Obstructed = false
			}

		case <-timerChannel:
			timerStop()
			if myStatus.Obstructed {
				timerStart(doorOpenDuration)
			} else {
				myStatus, myOrders = FSM_onDoorTimeout(myStatus, myOrders, elevio.GetFloor())
			}

		}
		elevStatus <- myStatus
		orders <- myOrders
	}
}
