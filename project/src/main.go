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

const UPDATE_IP_PERIOD = 5 * time.Second

var elevator types.Elevator
var masterList types.MasterList

func main() {
	elevio.Init("localhost:15657", constants.N_FLOORS)
	elevator = elevatorData.InitElevator()
	masterList.Elevators = append(masterList.Elevators, elevator)
	myStatus := make(chan types.ElevStatus)
	myOrders := make(chan [][]bool)

	go elevatorData.InitOrdersChan(myOrders)

	variables.MyIP = aliveMessages.GetPrimaryIP()
	ipCheck := time.NewTicker(UPDATE_IP_PERIOD)

	go roleConfiguration.ConfigureRoles(&elevator, &masterList)
	go elevatorAlgorithm.ElevatorControlLoop(&masterList, myStatus, myOrders, elevator.Orders, elevator.Role)

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
				elevatorData.UpdateStatusMasterList(&masterList, elevator.Status, variables.MyIP)
				communication.BroadcastMessage(&masterList)
			}
			elevatorData.SetAllLights(masterList)

		case newOrders := <-myOrders:
			if elevator.Role == constants.MASTER {
				elevatorData.UpdateLightsMasterList(&masterList, variables.MyIP)
				elevatorData.SetAllLights(masterList)
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
		case <-ipCheck.C:
			currentIP := aliveMessages.GetPrimaryIP()
			if variables.MyIP != currentIP && currentIP != "" {
				variables.MyIP = currentIP
				for index := range masterList.Elevators {
					masterList.Elevators[index].Ip = variables.MyIP
				}
			}
			continue
		}
	}
}
