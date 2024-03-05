package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
)

func ElevAlgo(masterList *elevData.MasterList, elevStatus chan elevData.ElevStatus, orders chan [][]bool, init_order [][]bool, role elevData.ElevatorRole) {
	var myStatus elevData.ElevStatus
	myOrders := init_order

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
				FSM_RequestFloor(masterList, a.Floor, int(a.Button))
			}
			myStatus.Buttonfloor = int(a.Button)
			myStatus.Buttontype = a.Floor
		case a := <-drvFloors:
			fmt.Println(a)
			myStatus = FMS_ArrivalAtFloor(myStatus, myOrders, a)
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
		}
		elevStatus <- myStatus
		orders <- myOrders
	}
}
