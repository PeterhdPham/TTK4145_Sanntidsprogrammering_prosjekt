package elevalgo

import (
	"Driver-go/elevio"
	"encoding/json"
	"fmt"
	"project/elevData"
	"project/tcp"
)

var FSM_State string

const (
	Idle     = "EB_Idle"
	Moving   = "EB_Moving"
	DoorOpen = "EB_DoorOpen"
)

func FSM_InitBetweenFloors(status elevData.ElevStatus) elevData.ElevStatus {
	elevio.SetMotorDirection(-1)
	FSM_State = Moving
	status.Direction = -1

	return status
}

func FSM_ArrivalAtFloor(status elevData.ElevStatus, orders [][]bool, floor int) elevData.ElevStatus {
	elevio.SetFloorIndicator(floor)
	status.Floor = floor
	status.Buttonfloor = -1
	status.Buttontype = -1
	switch FSM_State {
	case Moving:
		if requestShouldStop(status, orders, floor) {
			//Stops elevator and updates status accordingly
			elevio.SetMotorDirection(elevio.MD_Stop)

			//Opens elevator door and updates status accordingly
			elevio.SetDoorOpenLamp(true)
			status.Doors = true

			//Clears the request at current floor
			status, orders = requestClearAtFloor(status, orders, floor)

			timerStart(doorOpenDuration)

			SetAllLights(orders)
			FSM_State = DoorOpen
			fmt.Println("State: Open")
		}
	default:
		break
	}
	return status
}

func FSM_RequestFloor(master *elevData.MasterList, floor int, button int, fromIP string) (elevData.ElevStatus, [][]bool) {
	bestElevIP := findBestElevIP(master)
	if button == int(elevio.BT_Cab) {
		for elevator := range master.Elevators {
			if master.Elevators[elevator].Ip == fromIP {
				master.Elevators[elevator].Orders[floor][int(elevio.BT_Cab)] = true
			}
		}
	} else {
		for elevator := range master.Elevators {
			if master.Elevators[elevator].Ip == bestElevIP {
				master.Elevators[elevator].Orders[floor][button] = true
			}
		}
	}
	jsonToSend, err := json.Marshal(master)
	if err != nil {
		print("Error marshalling master: ", err)
	}
	tcp.BroadcastMessage(string(jsonToSend), nil)

	//Check orders and starts moving

	var status elevData.ElevStatus
	var orders [][]bool
	for _, e := range master.Elevators {
		if e.Ip == fromIP {
			status = e.Status
			orders = e.Orders
		}
	}
	if FSM_State == Idle {
		FSM_State = Moving
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		elevio.SetMotorDirection(pair.Dirn)
		FSM_State = pair.Behaviour
		fmt.Println("Direction and behaviors: ", pair)
		requestClearAtFloor(status, orders, floor)
	}

	return status, orders
}

func findBestElevIP(master *elevData.MasterList) string {
	numRequests := make(map[string]int, len(master.Elevators))
	for _, elevator := range master.Elevators {
		for _, floor := range elevator.Orders {
			for _, requested := range floor {
				if requested {
					numRequests[elevator.Ip]++
				}
			}
		}
	}
	var bestElevIP string = master.Elevators[0].Ip
	var bestElevVal int = 1e10
	for ip, value := range numRequests {
		if value > bestElevVal {
			bestElevVal = value
			bestElevIP = ip
		}
	}
	return bestElevIP
}

func FSM_onDoorTimeout(status elevData.ElevStatus, orders [][]bool, floor int) (elevData.ElevStatus, [][]bool) {
	fmt.Printf("\n\n%s()\n", "FSM_OnDoorTimeout")

	switch FSM_State {
	case DoorOpen:
		pair := requestsChooseDirection(status, orders)
		status.Direction = int(pair.Dirn)
		FSM_State = pair.Behaviour

		fmt.Println("Direction and behaviors FSM Door Open: ", pair)

		switch FSM_State {
		case DoorOpen:
			timerStart(doorOpenDuration)
			status, orders = requestClearAtFloor(status, orders, floor)
			SetAllLights(orders)
		case Moving, Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(status.Direction))
		}

	default:
		// No action for default case
	}

	return status, orders
}
