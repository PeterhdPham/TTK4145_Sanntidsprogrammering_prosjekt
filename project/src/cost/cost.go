package cost

import (
	"Driver-go/elevio"
	"fmt"
	"project/defs"
)

func FindAndAssign(master *defs.MasterList, floor int, button int, fromIP string) {
	if button == int(elevio.BT_Cab) {
		fmt.Println("CAB CALL")
		for elevator := range master.Elevators {
			if master.Elevators[elevator].Ip == fromIP {
				master.Elevators[elevator].Orders[floor][int(elevio.BT_Cab)] = true
			}
		}
	} else {
		bestElevIP := findBestElevIP(master)
		for elevator := range master.Elevators {
			if master.Elevators[elevator].Ip == bestElevIP {
				master.Elevators[elevator].Orders[floor][button] = true
			}
		}
	}
	fmt.Println(master)
}

func findBestElevIP(master *defs.MasterList) string {
	var ipRequests []defs.IpRequestCount
	for _, elevator := range master.Elevators {
		ipRequests = append(ipRequests, defs.IpRequestCount{Ip: elevator.Ip, Requests: 0})
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

	// Find the elevator with the least number of requests
	bestElevIP := ipRequests[0].Ip
	bestElevVal := ipRequests[0].Requests
	for _, ipRequest := range ipRequests {
		if ipRequest.Requests < bestElevVal {
			bestElevVal = ipRequest.Requests
			bestElevIP = ipRequest.Ip
		}
	}

	return bestElevIP
}
