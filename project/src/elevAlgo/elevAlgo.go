package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"project/tcp"
	"time"
)

var N_FLOORS int
var doorOpenDuration time.Duration = 3 * time.Second
var MyIP string
var failureTimeoutDuration time.Duration = 7 * time.Second

type FailureMode int

const (
	DoorStuck FailureMode = 0
	MotorFail FailureMode = 1
)

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
		myStatus.FSM_State = Idle
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
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			if a {
				myStatus.Obstructed = true
				if !(myStatus.FSM_State == Moving) {
					elevio.SetMotorDirection(elevio.MD_Stop)
				}
			} else {
				myStatus.Obstructed = false
				myStatus.Operative = true
				failureTimerStop()
			}

		case <-timerChannel:
			timerStop()
			if myStatus.Obstructed {
				timerStart(doorOpenDuration)
			} else {
				myStatus, myOrders = FSM_onDoorTimeout(myStatus, myOrders, elevio.GetFloor())
				failureTimerStop()
			}
		case mode := <-failureTimerChannel:
			failureTimerStop()
			switch mode {
			case 0:
				fmt.Println("DOORS ARE STUCK")
			case 1:
				fmt.Println("MOTOR HAS FAILED. TRYING AGAIN")
				elevio.SetMotorDirection(elevio.MotorDirection(myStatus.Direction))
			}
			myStatus.Operative = false
			failureTimerStart(failureTimeoutDuration, mode)
		}
		if tcp.UpdateLocal {
			tcp.UpdateLocal = false
			myStatus, myOrders = FSM_RequestFloor(masterList, -1, -1, "", elevData.Slave)
		}

		elevStatus <- myStatus
		orders <- myOrders
	}
}
