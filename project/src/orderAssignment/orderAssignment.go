package orderAssignment

import (
	"Driver-go/elevio"
	"project/types"
)

const INIT_MIN_ORDERS = 1000

func FindAndAssign(master *types.MasterList, floor int, button int, fromIP string) {
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
			}
		}
		for elevator := range master.Elevators {
			master.Elevators[elevator].Lights[floor][button] = true
		}
	}
}

func findBestElevIP(master *types.MasterList) string {
	var ipRequests []types.IpRequestCount
	for _, elevator := range master.Elevators {
		ipRequests = append(ipRequests, types.IpRequestCount{Ip: elevator.Ip, Requests: 0, Operative: elevator.Status.Operative, Online: elevator.IsOnline})
	}

	for i, elevator := range master.Elevators {
		for _, floor := range elevator.Orders {
			for _, requested := range floor {
				if requested {
					ipRequests[i].Requests++
				}
			}
		}
	}

	bestElevIP := ""
	bestElevVal := INIT_MIN_ORDERS
	for _, ipRequest := range ipRequests {
		if (ipRequest.Requests < bestElevVal) && ipRequest.Operative && ipRequest.Online {
			bestElevVal = ipRequest.Requests
			bestElevIP = ipRequest.Ip
		}
	}

	return bestElevIP
}
