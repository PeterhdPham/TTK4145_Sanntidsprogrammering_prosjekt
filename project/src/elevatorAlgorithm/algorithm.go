package elevatorAlgorithm

import (
	"Driver-go/elevio"
	"project/communication"
	"project/constants"
	"project/elevatorData"
	"project/roleConfiguration"
	"project/types"
	"project/variables"

	"time"
)

var doorOpenDuration time.Duration = 3 * time.Second
var failureTimeoutDuration time.Duration = 7 * time.Second

func ElevatorControlLoop(masterList *types.MasterList, elevStatus chan types.ElevStatus, orders chan [][]bool, init_order [][]bool, role types.ElevatorRole) {
	myStatus := elevatorData.InitStatus()
	myOrders := init_order

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)

	if elevio.GetFloor() == -1 {
		myStatus = initBetweenFloors(myStatus)
	} else {
		myStatus.FSM_State = constants.IDLE
	}

	for {
		select {
		case a := <-drvButtons:
			if role == constants.MASTER {
				myStatus, myOrders = floorRequested(masterList, myStatus, myOrders, a.Floor, int(a.Button), variables.MyIP, role)
			}
			myStatus.Buttonfloor = a.Floor
			myStatus.Buttontype = int(a.Button)
		case a := <-drvFloors:
			myStatus, myOrders = arrivalAtFloor(myStatus, myOrders, a)
			if role == constants.MASTER {
				elevatorData.UpdateLightsMasterList(masterList, variables.MyIP)
			}
		case a := <-drvObstr:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			if a {
				myStatus.Obstructed = true
				if !(myStatus.FSM_State == constants.MOVING) {
					elevio.SetMotorDirection(elevio.MD_Stop)
				}
			} else {
				myStatus.Obstructed = false
				myStatus.Operative = true
				failureTimerStop()
			}

		case <-doorTimerChannel:
			doorTimerStop()
			if myStatus.Obstructed {
				doorTimerStart(doorOpenDuration)
			} else {
				failureTimerStop()
				myStatus, myOrders = onDoorTimeout(myStatus, myOrders, elevio.GetFloor())
			}
		case a := <-variables.ButtonReceived:
			floor := a.Event.Floor
			button := int(a.Event.Button)
			myStatus, myOrders = floorRequested(masterList, myStatus, myOrders, floor, button, a.IP, constants.MASTER)

		case ipAddress := <-variables.StatusReceived:
			elevatorData.UpdateStatusMasterList(masterList, variables.RemoteStatus, ipAddress)
			if variables.RemoteStatus.Operative {
				roleConfiguration.ReassignOrdersIfInoperative(masterList)
			}
			communication.BroadcastMessage(masterList)
		case <-variables.UpdateLocal:
			myStatus, myOrders = floorRequested(masterList, myStatus, myOrders, -1, -1, "", constants.SLAVE)
			elevatorData.SetAllLights(*masterList)

		case mode := <-failureTimerChannel:
			failureTimerStop()

			switch mode {
			case int(DOOR_STUCK):
				myStatus.Operative = false
			case int(MOTOR_FAIL):
				if myStatus.FSM_State != constants.IDLE {
					elevio.SetMotorDirection(elevio.MotorDirection(myStatus.Direction))
					myStatus.Operative = false
				}
			}
			if myStatus.Doors || myStatus.FSM_State != constants.IDLE {
				failureTimerStop()
				failureTimerStart(failureTimeoutDuration, mode)
			}
			if (role == constants.MASTER) && !(myStatus.Operative) {
				elevatorData.UpdateStatusMasterList(masterList, myStatus, variables.MyIP)
				roleConfiguration.ReassignOrdersIfInoperative(masterList)
				communication.BroadcastMessage(masterList)
			}
		}

		elevatorData.SetAllLights(*masterList)

		elevStatus <- myStatus
		orders <- myOrders
	}
}
