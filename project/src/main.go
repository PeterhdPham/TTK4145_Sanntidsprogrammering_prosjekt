package main

import (
	"Driver-go/elevio"
	"log"
	"project/aliveMessages"
	"project/communication"
	"project/constants"
	"project/elevatorAlgorithm"
	"project/elevatorData"
	"project/roleConfiguration"
	"project/types"
	"project/utility"
	"project/variables"
	"reflect"
	"time"
)

var elevator types.Elevator
var masterElevator types.MasterList

func main() {
	elevio.Init("localhost:15657", constants.N_FLOORS)
	elevator = elevatorData.InitElevator()
	masterElevator.Elevators = append(masterElevator.Elevators, elevator)
	myStatus := make(chan types.ElevStatus)
	myOrders := make(chan [][]bool)

	go elevatorData.InitOrdersChan(myOrders)

	variables.MyIP = aliveMessages.GetPrimaryIP()
	ticker := time.NewTicker(5 * time.Second)

	go roleConfiguration.ConfigureRoles(&elevator, &masterElevator)
	go elevatorAlgorithm.ElevatorControlLoop(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role)

	for {
		select {
		case newStatus := <-myStatus:
			if roleConfiguration.ServerConnection != nil && elevator.Role == constants.SLAVE {
				if !reflect.DeepEqual(elevator.Status, newStatus) {
					elevator.Status = newStatus

					err := communication.SendMessage(roleConfiguration.ServerConnection, newStatus, "")
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			} else if elevator.Role == constants.MASTER {
				elevator.Status = newStatus
				elevatorData.UpdateStatusMasterList(&masterElevator, elevator.Status, variables.MyIP)
				communication.BroadcastMessage(&masterElevator)
			}
			elevatorData.SetAllLights(masterElevator)

		case newOrders := <-myOrders:
			if elevator.Role == constants.MASTER {
				elevatorData.UpdateLightsMasterList(&masterElevator, variables.MyIP)
				elevatorData.SetAllLights(masterElevator)
			}
			if !utility.SlicesAreEqual(elevator.Orders, newOrders) {
				elevator.Orders = newOrders
				if roleConfiguration.ServerConnection != nil && elevator.Role == constants.SLAVE {

					err := communication.SendMessage(roleConfiguration.ServerConnection, elevator, "")
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			}
		case <-ticker.C:
			currentIP := aliveMessages.GetPrimaryIP()
			if variables.MyIP != currentIP && currentIP != "" {
				variables.MyIP = currentIP
				for index := range masterElevator.Elevators {
					masterElevator.Elevators[index].Ip = variables.MyIP
				}
			}
			continue
		}
	}
}
