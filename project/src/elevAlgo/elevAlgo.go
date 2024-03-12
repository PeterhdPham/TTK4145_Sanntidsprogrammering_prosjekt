package elevAlgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/broadcast"
	"project/defs"
	"project/elevData"
	"project/tcp"
	"project/utility"
	"time"
)

var doorOpenDuration time.Duration = 3 * time.Second
var failureTimeoutDuration time.Duration = 7 * time.Second

func ElevAlgo(masterList *defs.MasterList, elevStatus chan defs.ElevStatus, orders chan [][]bool, lights chan [][]bool, init_order [][]bool, role defs.ElevatorRole) {
	var myStatus defs.ElevStatus
	myOrders := init_order
	myLights := init_order

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
	}

	for {
		select {
		case a := <-drvButtons:
			if role == defs.MASTER {
				myStatus, myOrders, myLights = FSM_RequestFloor(masterList, a.Floor, int(a.Button), defs.MyIP, role)
			}
			myStatus.Buttonfloor = a.Floor
			myStatus.Buttontype = int(a.Button)
		case a := <-drvFloors:
			myStatus = FSM_ArrivalAtFloor(myStatus, myOrders, myLights, a)
		case a := <-drvObstr:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			if a {
				myStatus.Obstructed = true
				if !(myStatus.FSM_State == defs.MOVING) {
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
				failureTimerStop()
				myStatus, myOrders = FSM_onDoorTimeout(myStatus, myOrders, myLights, elevio.GetFloor())
			}
		case a := <-defs.ButtonReceived:
			requestFloor := a.Event.Floor
			requestButton := int(a.Event.Button)
			myStatus, myOrders, myLights = FSM_RequestFloor(masterList, requestFloor, requestButton, a.IP, defs.MASTER)

		case ipAddress := <-defs.StatusReceived:
			elevData.UpdateStatusMasterList(masterList, defs.RemoteStatus, ipAddress)
		case <-defs.UpdateLocal:
			myStatus, myOrders, myLights = FSM_RequestFloor(masterList, -1, -1, "", defs.SLAVE)
		case mode := <-failureTimerChannel:
			failureTimerStop()
			if (role == defs.MASTER) && (myStatus.Operative) {
				tcp.ReassignOrders2(masterList)
				jsonToSend := utility.MarshalJson(masterList)
				broadcast.BroadcastMessage(nil, jsonToSend)
			}

			switch mode {
			case 0:
				fmt.Println("DOORS ARE STUCK")
				myStatus.Operative = false
			case 1:
				if myStatus.FSM_State != defs.IDLE {
					fmt.Println("MOTOR HAS FAILED. TRYING AGAIN")
					elevio.SetMotorDirection(elevio.MotorDirection(myStatus.Direction))
					myStatus.Operative = false
				}
			}
			if myStatus.Doors || myStatus.FSM_State != defs.IDLE {
				failureTimerStop()
				failureTimerStart(failureTimeoutDuration, mode)
			}
		}
		if tcp.UpdateLocal {
			tcp.UpdateLocal = false
			myStatus, myOrders, myLights = FSM_RequestFloor(masterList, -1, -1, "", defs.SLAVE)
		}

		elevStatus <- myStatus
		orders <- myOrders
		lights <- myLights
	}
}
