package elevData

import (
	"project/defs"
	"project/udp"
)

func InitElevator(NumberOfFloors int) defs.Elevator {
	var elevator defs.Elevator
	ip, _ := udp.GetPrimaryIP()
	elevator.Status.Buttonfloor = -1
	elevator.Status.Buttontype = -1
	elevator.Ip = ip
	elevator.Orders = InitOrders(NumberOfFloors)
	elevator.Status.FSM_State = defs.IDLE
	return elevator
}

func InitOrders(NumberOfFloors int) [][]bool {
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

func UpdateLightsMasterList(masterElevator *defs.MasterList, newLights [][]bool, ip string) {
	var elevator defs.Elevator
	var index int

	for eIndex, e := range masterElevator.Elevators {
		if e.Ip == ip {
			elevator = e
			index = eIndex
		}
	}
	for fIndex, floorLights := range elevator.Lights {
		for lIndex, light := range floorLights {
			if newLights[fIndex][lIndex] != light {
				masterElevator.Elevators[index].Lights[fIndex][lIndex] = newLights[fIndex][lIndex]
			}
		}
	}
}
