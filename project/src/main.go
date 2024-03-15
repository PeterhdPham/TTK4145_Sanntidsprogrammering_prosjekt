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
	elevio.Init("localhost:15657", constants.N_FLOORS) // connect to elevatorsimulator
	elevator = elevatorData.InitElevator()             // initialize the elevator

	masterElevator.Elevators = append(masterElevator.Elevators, elevator) // append the elevator to the master list of elevators

	myStatus := make(chan types.ElevStatus)  // channel to receive status updates
	myOrders := make(chan [][]bool)          // channel to receive order updates
	go elevatorData.InitOrdersChan(myOrders) // initialize the orders channel

	variables.MyIP = aliveMessages.GetPrimaryIP()

	ticker := time.NewTicker(5 * time.Second)

	go roleConfiguration.Config_Roles(&elevator, &masterElevator) // initialize the server and client connections

	go elevatorAlgorithm.ElevatorControlLoop(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role) // initialize the elevator algorithm

	for {
		select {
		case newStatus := <-myStatus:
			// log.Println("status update: ", string(utility.MarshalJson(newStatus)))

			//Sends message to server
			if roleConfiguration.ServerConnection != nil && elevator.Role == constants.SLAVE {
				if !reflect.DeepEqual(elevator.Status, newStatus) {
					elevator.Status = newStatus
					// Convert message to byte slice
					err := communication.SendMessage(roleConfiguration.ServerConnection, newStatus, "") // Assign the error value to "err"
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			} else if elevator.Role == constants.MASTER {
				elevator.Status = newStatus
				elevatorData.UpdateStatusMasterList(&masterElevator, elevator.Status, variables.MyIP)
				communication.BroadcastMessage(nil, &masterElevator)
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
					// Convert message to byte slice
					err := communication.SendMessage(roleConfiguration.ServerConnection, elevator, "") // Assign the error value to "err"
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			}
		case <-ticker.C:
			bytes := utility.MarshalJson(masterElevator)
			log.Println("")
			log.Println("MasterList: ", string(bytes))
			log.Println("\nActive ips: ", roleConfiguration.ActiveIPs)
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
