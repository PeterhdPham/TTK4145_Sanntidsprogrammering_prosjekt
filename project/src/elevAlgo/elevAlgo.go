package elevAlgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"project/ip"
	"project/variable"
	"time"
)

var N_FLOORS int
var doorOpenDuration time.Duration = 3 * time.Second
var MyIP string

func ElevAlgo(masterList *elevData.MasterList, elevStatus chan elevData.ElevStatus, orders chan [][]bool, init_order [][]bool, role elevData.ElevatorRole, N_Floors int) {
	var myStatus elevData.ElevStatus
	myOrders := init_order
	N_FLOORS = N_Floors
	MyIP, _ = ip.GetPrimaryIP()

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
		myStatus.FSM_State = variable.Idle
		myStatus.Floor = elevio.GetFloor()
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
		case <-variable.MessageReceived:
			if variable.UpdateOrdersFromMessage {
				fmt.Print("UpdateFromMessage")
				variable.UpdateStatusFromMessage = false
				requestFloor := elevData.RemoteStatus.Buttonfloor
				requestButton := elevData.RemoteStatus.Buttontype
				myStatus, myOrders = FSM_RequestFloor(masterList, requestFloor, requestButton, variable.MyIP, elevData.Master)
			} else if variable.UpdateStatusFromMessage {
				fmt.Print("update status from message")
				myStatus = elevData.RemoteStatus
				variable.UpdateStatusFromMessage = false
			}
		}
		if variable.UpdateLocal {
			variable.UpdateLocal = false
			fmt.Print("UpdateLocal")
			myStatus, myOrders = FSM_RequestFloor(masterList, -1, -1, "", elevData.Slave)
		}

		elevStatus <- myStatus
		orders <- myOrders
	}
}
