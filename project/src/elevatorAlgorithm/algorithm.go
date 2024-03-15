package elevatorAlgorithm

import (
	"Driver-go/elevio"
	"log"
	"project/communication"
	"project/defs"
	"project/elevatorData"
	"project/roleConfiguration"
	"time"
)

var doorOpenDuration time.Duration = 3 * time.Second
var failureTimeoutDuration time.Duration = 7 * time.Second

func ElevatorControlLoop(masterList *defs.MasterList, elevStatus chan defs.ElevStatus, orders chan [][]bool, init_order [][]bool, role defs.ElevatorRole) {
	myStatus := elevatorData.InitStatus()
	myOrders := init_order

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
				myStatus, myOrders = FSM_RequestFloor(masterList, myStatus, myOrders, a.Floor, int(a.Button), defs.MyIP, role)
			}
			myStatus.Buttonfloor = a.Floor
			myStatus.Buttontype = int(a.Button)
		case a := <-drvFloors:
			myStatus, myOrders = FSM_ArrivalAtFloor(myStatus, myOrders, a)
			if role == defs.MASTER {
				elevatorData.UpdateLightsMasterList(masterList, defs.MyIP)
			}
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
				myStatus, myOrders = FSM_onDoorTimeout(myStatus, myOrders, elevio.GetFloor())
			}
		case a := <-defs.ButtonReceived:
			requestFloor := a.Event.Floor
			requestButton := int(a.Event.Button)
			myStatus, myOrders = FSM_RequestFloor(masterList, myStatus, myOrders, requestFloor, requestButton, a.IP, defs.MASTER)

		case ipAddress := <-defs.StatusReceived:
			elevatorData.UpdateStatusMasterList(masterList, defs.RemoteStatus, ipAddress)
			if defs.RemoteStatus.Operative {
				roleConfiguration.ReassignOrders2(masterList)
			}
			communication.BroadcastMessage(nil, masterList)
		case <-defs.UpdateLocal:
			myStatus, myOrders = FSM_RequestFloor(masterList, myStatus, myOrders, -1, -1, "", defs.SLAVE)
			elevatorData.SetAllLights(*masterList)

		case mode := <-failureTimerChannel:
			failureTimerStop()

			switch mode {
			case 0:
				log.Println("DOORS ARE STUCK")
				myStatus.Operative = false
			case 1:
				if myStatus.FSM_State != defs.IDLE {
					log.Println("MOTOR HAS FAILED. TRYING AGAIN")
					elevio.SetMotorDirection(elevio.MotorDirection(myStatus.Direction))
					myStatus.Operative = false
				}
			}
			if myStatus.Doors || myStatus.FSM_State != defs.IDLE {
				failureTimerStop()
				failureTimerStart(failureTimeoutDuration, mode)
			}
			if (role == defs.MASTER) && !(myStatus.Operative) {
				elevatorData.UpdateStatusMasterList(masterList, myStatus, defs.MyIP)
				roleConfiguration.ReassignOrders2(masterList)
				communication.BroadcastMessage(nil, masterList)
			}
		}

		elevatorData.SetAllLights(*masterList)

		elevStatus <- myStatus
		orders <- myOrders
	}
}
