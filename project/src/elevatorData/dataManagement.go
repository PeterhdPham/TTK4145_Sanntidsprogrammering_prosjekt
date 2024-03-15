package elevatorData

import (
	"Driver-go/elevio"
	"project/aliveMessages"
	"project/defs"
	"project/utility"
)

func InitElevator() defs.Elevator {
	var elevator defs.Elevator
	elevator.IsOnline = true
	elevator.Ip = aliveMessages.GetPrimaryIP()
	elevator.Orders = InitOrdersAndLights()
	elevator.Lights = InitOrdersAndLights()
	elevator.Status = InitStatus()
	return elevator
}

func InitStatus() defs.ElevStatus {
	var status defs.ElevStatus
	status.Direction = elevio.MD_Stop
	status.Floor = -1
	status.Doors = false
	status.Obstructed = false
	status.Buttonfloor = -1
	status.Buttontype = -1
	status.FSM_State = defs.IDLE
	status.Operative = true
	return status
}

func InitOrdersAndLights() [][]bool {
	orders := make([][]bool, defs.N_FLOORS)
	for i := range orders {
		orders[i] = make([]bool, 3)
	}
	return orders
}

func InitOrdersChan(orders chan [][]bool) {
	o := make([][]bool, defs.N_FLOORS)
	for i := 0; i < defs.N_FLOORS; i++ {
		o[i] = make([]bool, 3) // Assuming 3 buttons per floor.
	}
	// Send the initialized slice of slices through the channel.
	orders <- o
}

func UpdateStatusMasterList(masterList *defs.MasterList, newStatus defs.ElevStatus, ip string) {
	for i := 0; i < len(masterList.Elevators); i++ {
		if masterList.Elevators[i].Ip == ip {
			masterList.Elevators[i].Status = newStatus
		}
	}
}
func UpdateOrdersMasterList(masterList *defs.MasterList, newOrders [][]bool, ip string) {
	for i := 0; i < len(masterList.Elevators); i++ {
		if masterList.Elevators[i].Ip == ip {
			masterList.Elevators[i].Orders = newOrders
		}
	}
}

func UpdateLightsMasterList(masterList *defs.MasterList, ip string) {
	for floor := 0; floor < defs.N_FLOORS; floor++ {
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

func UpdateIsOnline(masterElevator *defs.MasterList, oldList []string, newList []string) {
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

func SetAllLights(masterList defs.MasterList) {
	for index, e := range masterList.Elevators {
		if e.Ip == defs.MyIP {
			for floor := 0; floor < defs.N_FLOORS; floor++ {
				for btn := elevio.BT_HallUp; btn <= elevio.BT_Cab; btn++ {
					elevio.SetButtonLamp(btn, floor, masterList.Elevators[index].Lights[floor][btn])
				}
			}
		}
	}
}
