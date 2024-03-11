package cost

import (
	"Driver-go/elevio"
	"fmt"
	"project/variable"
)

func FindAndAssign(master *variable.MasterList, floor int, button int, fromIP string) {
	fmt.Printf("Floor: %d and Button: %d\n", floor, button)
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

}

func findBestElevIP(master *variable.MasterList) string {
	if len(master.Elevators) == 0 {
		fmt.Println("Empty")
		return ""
	} else {
		fmt.Println("Length: ", len(master.Elevators))
	}

	// Initialize a slice to hold the IP and request count pairs
	var ipRequests []variable.IpRequestCount

	// Populate ipRequests with all elevators and initialize request counts to 0
	for _, elevator := range master.Elevators {
		ipRequests = append(ipRequests, variable.IpRequestCount{Ip: elevator.Ip, Requests: 0})
	}

	// Count requests for each elevator
	for i, elevator := range master.Elevators {
		for _, floor := range elevator.Orders {
			for _, requested := range floor {
				if requested {
					ipRequests[i].Requests++
				}
			}
		}
	}

	// Print the number of requests per IP for debugging
	for _, ipRequest := range ipRequests {
		fmt.Printf("Elevator IP: %s, Number of Orders: %d\n", ipRequest.Ip, ipRequest.Requests)
	}

	// Find the elevator with the least number of requests
	bestElevIP := ipRequests[0].Ip
	bestElevVal := ipRequests[0].Requests
	for _, ipRequest := range ipRequests {
		if ipRequest.Requests < bestElevVal {
			bestElevVal = ipRequest.Requests
			bestElevIP = ipRequest.Ip
		}
	}

	fmt.Printf("Best IP: %s with %d orders\n", bestElevIP, bestElevVal)
	return bestElevIP
}
