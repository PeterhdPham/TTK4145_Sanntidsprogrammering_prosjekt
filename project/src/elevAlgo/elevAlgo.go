package elevAlgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"project/variable"
	"time"
)

var N_FLOORS int
var doorOpenDuration time.Duration = 3 * time.Second

func ElevAlgo(masterList *variable.MasterList, elevStatus chan variable.ElevStatus, orders chan [][]bool, init_order [][]bool, role variable.ElevatorRole, N_Floors int) {
	var myStatus variable.ElevStatus
	myOrders := init_order
	N_FLOORS = N_Floors

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
		myStatus.FSM_State = variable.IDLE
		myStatus.Floor = elevio.GetFloor()
	}

	for {
		select {
		case a := <-drvButtons:
			if role == variable.MASTER {
				myStatus, myOrders = FSM_RequestFloor(masterList, a.Floor, int(a.Button), variable.MyIP, role)
			} else {
				myStatus.Buttonfloor = a.Floor
				myStatus.Buttontype = int(a.Button)
				fmt.Printf("Floor %d and Button %d\n", a.Floor, int(a.Button))
			}
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
		case a := <-variable.ButtonReceived:
			fmt.Println("Update Orders")
			requestFloor := a.Event.Floor
			requestButton := int(a.Event.Button)
			myStatus, myOrders = FSM_RequestFloor(masterList, requestFloor, requestButton, a.IP, variable.MASTER)

		case ipAddress := <-variable.StatusReceived:
			fmt.Printf("update status from %s\n", ipAddress)
			elevData.UpdateMasterList(masterList, elevData.RemoteStatus, ipAddress)
		case a := <-variable.UpdateLocal:
			fmt.Println("Update Local Master List: ", a)
			myStatus, myOrders = FSM_RequestFloor(masterList, -1, -1, "", variable.SLAVE)
		}
		// if variable.UpdateLocal {
		// 	variable.UpdateLocal = false
		// 	fmt.Println("Update Local Master List")
		// 	myStatus, myOrders = FSM_RequestFloor(masterList, -1, -1, "", variable.SLAVE)
		// }

		elevStatus <- myStatus
		orders <- myOrders
	}
}
