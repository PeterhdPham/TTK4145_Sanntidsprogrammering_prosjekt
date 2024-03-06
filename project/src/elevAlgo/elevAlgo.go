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

func ElevAlgo(masterList *elevData.MasterList, elevStatus chan elevData.ElevStatus, orders chan [][]bool, init_order [][]bool, role elevData.ElevatorRole, N_Floors int) {
	var myStatus elevData.ElevStatus
	myOrders := init_order
	N_FLOORS = N_Floors
	MyIP, _ = tcp.GetPrimaryIP()

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)

	// Moves the elevator down if in between floors
	fmt.Printf("Current floor %d\n", elevio.GetFloor())
	if elevio.GetFloor() == -1 {
		myStatus = FSM_InitBetweenFloors(myStatus)
		fmt.Println("In between")
	} else {
		FSM_State = Idle
	}

	for {
		select {
		// case NewElevator := <-:
		// 	elevStatus = NewElevator
		// case NewOrders := <-o:
		// 	orders = NewOrders
		case a := <-drvButtons:
			fmt.Println(a)
			if role == elevData.Master {
				myStatus, myOrders = FSM_RequestFloor(masterList, a.Floor, int(a.Button), MyIP)
			}
			myStatus.Buttonfloor = a.Floor
			myStatus.Buttontype = int(a.Button)
		case a := <-drvFloors:
			fmt.Println(a)
			myStatus = FSM_ArrivalAtFloor(myStatus, myOrders, a)
		case a := <-drvObstr:
			fmt.Printf("%+v\n", a)
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

		case a := <-drvStop:
			fmt.Printf("%+v\n", a)
			// TODO: Clear all orders and lights from elevator

		case <-timerChannel:
			fmt.Println("Timer timed out")
			timerStop()
			myStatus, myOrders = FSM_onDoorTimeout(myStatus, myOrders, elevio.GetFloor())
		}
		elevStatus <- myStatus
		orders <- myOrders
	}
}
