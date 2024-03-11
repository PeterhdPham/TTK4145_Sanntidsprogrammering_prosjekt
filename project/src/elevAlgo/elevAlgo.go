package elevAlgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/defs"
	"project/elevData"
	"time"
)

var N_FLOORS int
var doorOpenDuration time.Duration = 3 * time.Second

func ElevAlgo(masterList *defs.MasterList, elevStatus chan defs.ElevStatus, orders chan [][]bool, init_order [][]bool, role defs.ElevatorRole, N_Floors int) {
	var myStatus defs.ElevStatus
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
		myStatus.FSM_State = defs.IDLE
		myStatus.Floor = elevio.GetFloor()
	}

	for {
		select {
		case a := <-drvButtons:
			if role == defs.MASTER {
				myStatus, myOrders = FSM_RequestFloor(masterList, a.Floor, int(a.Button), defs.MyIP, role)
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
		case a := <-defs.ButtonReceived:
			requestFloor := a.Event.Floor
			requestButton := int(a.Event.Button)
			myStatus, myOrders = FSM_RequestFloor(masterList, requestFloor, requestButton, a.IP, defs.MASTER)

		case ipAddress := <-defs.StatusReceived:
			elevData.UpdateStatusMasterList(masterList, elevData.RemoteStatus, ipAddress)
		case <-defs.UpdateLocal:
			myStatus, myOrders = FSM_RequestFloor(masterList, -1, -1, "", defs.SLAVE)
		}

		elevStatus <- myStatus
		orders <- myOrders
	}
}
