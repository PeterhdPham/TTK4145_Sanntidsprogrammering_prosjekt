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
	elevator.Orders = InitOrdersAndLights(NumberOfFloors)
	elevator.Lights = InitOrdersAndLights(NumberOfFloors)
	elevator.Status.FSM_State = defs.IDLE
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

func UpdateLightsMasterList(masterList *defs.MasterList, newLights [][]bool, ip string) {
	for index, e := range masterList.Elevators {
		if e.Ip == ip {
			masterList.Elevators[index].Lights = newLights
		} else {
			for floorIndex := range e.Lights {
				masterList.Elevators[index].Lights[floorIndex][0] = newLights[floorIndex][0]
				masterList.Elevators[index].Lights[floorIndex][1] = newLights[floorIndex][1]
			}
		}

	}
}
