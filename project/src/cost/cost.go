package cost

import (
	"Driver-go/elevio"
	"log"
	"project/defs"
)

func FindAndAssign(master *defs.MasterList, floor int, button int, fromIP string) {
	for index := range master.Elevators {
		if button == int(elevio.BT_Cab) {
			continue
		}
		if master.Elevators[index].Orders[floor][button] {
			return
		}
	}

	if button == int(elevio.BT_Cab) {
		for elevator := range master.Elevators {
			if master.Elevators[elevator].Ip == fromIP {
				master.Elevators[elevator].Orders[floor][button] = true
				master.Elevators[elevator].Lights[floor][button] = true
			}
		}
	} else {
		bestElevIP := findBestElevIP(master)
		for elevator := range master.Elevators {
			if master.Elevators[elevator].Ip == bestElevIP {
				master.Elevators[elevator].Orders[floor][button] = true
				log.Printf("Order %d at floor %d assigned to ip %s", floor, button, bestElevIP)
			}
		}
		for elevator := range master.Elevators {
			master.Elevators[elevator].Lights[floor][button] = true
		}
	}
}

func findBestElevIP(master *defs.MasterList) string {
	var ipRequests []defs.IpRequestCount
	for _, elevator := range master.Elevators {
		ipRequests = append(ipRequests, defs.IpRequestCount{Ip: elevator.Ip, Requests: 0, Operative: elevator.Status.Operative, Online: elevator.IsOnline})
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
	bestElevIP := ""
	bestElevVal := 1000
	for _, ipRequest := range ipRequests {
		if (ipRequest.Requests < bestElevVal) && ipRequest.Operative && ipRequest.Online {
			bestElevVal = ipRequest.Requests
			bestElevIP = ipRequest.Ip
		}
	}

	return bestElevIP
}
