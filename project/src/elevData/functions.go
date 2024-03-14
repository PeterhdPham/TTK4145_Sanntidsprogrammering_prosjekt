package elevData

import (
	"project/defs"
	"project/udp"
)

func InitElevator(NumberOfFloors int) defs.Elevator {
	var elevator defs.Elevator
	ip, _, _ := udp.GetPrimaryIP()
	elevator.Status.Buttonfloor = -1
	elevator.Status.Buttontype = -1
	elevator.Ip = ip
	elevator.Orders = InitOrdersAndLights(NumberOfFloors)
	elevator.Lights = InitOrdersAndLights(NumberOfFloors)
	elevator.Status.FSM_State = defs.IDLE
	elevator.Status.Operative = true
	return elevator
}

func InitOrdersAndLights(NumberOfFloors int) [][]bool {
	orders := make([][]bool, NumberOfFloors)
	for i := range orders {
		orders[i] = make([]bool, 3)
	}
	return orders
}

func InitOrdersChan(orders chan [][]bool, numOfFloors int) {
	o := make([][]bool, numOfFloors)
	for i := 0; i < numOfFloors; i++ {
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
