package elevatorData

import (
	"Driver-go/elevio"
	"project/aliveMessages"
	"project/constants"
	"project/types"
	"project/utility"
	"project/variables"
)

func InitElevator() types.Elevator {
	var elevator types.Elevator
	elevator.IsOnline = true
	elevator.Ip = aliveMessages.GetPrimaryIP()
	elevator.Orders = InitOrdersAndLights()
	elevator.Lights = InitOrdersAndLights()
	elevator.Status = InitStatus()
	return elevator
}

func InitStatus() types.ElevStatus {
	var status types.ElevStatus
	status.Direction = elevio.MD_Stop
	status.Floor = -1
	status.Doors = false
	status.Obstructed = false
	status.Buttonfloor = -1
	status.Buttontype = -1
	status.FSM_State = constants.IDLE
	status.Operative = true
	return status
}

func InitOrdersAndLights() [][]bool {
	orders := make([][]bool, constants.N_FLOORS)
	for i := range orders {
		orders[i] = make([]bool, 3)
	}
	return orders
}

func InitOrdersChan(orders chan [][]bool) {
	o := make([][]bool, constants.N_FLOORS)
	for i := 0; i < constants.N_FLOORS; i++ {
		o[i] = make([]bool, 3) // Assuming 3 buttons per floor.
	}
	// Send the initialized slice of slices through the channel.
	orders <- o
}

func UpdateStatusMasterList(masterList *types.MasterList, newStatus types.ElevStatus, ip string) {
	for i := 0; i < len(masterList.Elevators); i++ {
		if masterList.Elevators[i].Ip == ip {
			masterList.Elevators[i].Status = newStatus
		}
	}
}
func UpdateOrdersMasterList(masterList *types.MasterList, newOrders [][]bool, ip string) {
	for i := 0; i < len(masterList.Elevators); i++ {
		if masterList.Elevators[i].Ip == ip {
			masterList.Elevators[i].Orders = newOrders
		}
	}
}

func UpdateLightsMasterList(masterList *types.MasterList, ip string) {
	for floor := 0; floor < constants.N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			lightActive := false
			for index := range masterList.Elevators {
				if masterList.Elevators[index].Orders[floor][btn] {
					lightActive = true
					break
				}
			}
			for index := range masterList.Elevators {
				masterList.Elevators[index].Lights[floor][btn] = lightActive
			}
		}
		for index := range masterList.Elevators {
			if masterList.Elevators[index].Orders[floor][2] {
				masterList.Elevators[index].Lights[floor][2] = true
			} else {
				masterList.Elevators[index].Lights[floor][2] = false
			}
		}
	}
}

func UpdateIsOnline(masterElevator *types.MasterList, oldList []string, newList []string) {
	for _, elevIP := range oldList {
		if !utility.Contains(newList, elevIP) {
			for indx, e := range masterElevator.Elevators {
				if e.Ip == elevIP {
					masterElevator.Elevators[indx].IsOnline = false
				}
			}
		}
	}
	for _, elevIP := range newList {
		if !utility.Contains(oldList, elevIP) {
			for indx, e := range masterElevator.Elevators {
				if e.Ip == elevIP {
					masterElevator.Elevators[indx].IsOnline = true
				}
			}
		}
	}
}

func SetAllLights(masterList types.MasterList) {
	for index, e := range masterList.Elevators {
		if e.Ip == variables.MyIP {
			for floor := 0; floor < constants.N_FLOORS; floor++ {
				for btn := elevio.BT_HallUp; btn <= elevio.BT_Cab; btn++ {
					elevio.SetButtonLamp(btn, floor, masterList.Elevators[index].Lights[floor][btn])
				}
			}
		}
	}
}
