package elevalgo

import (
	"Driver-go/elevio"
	"project/elevData"
)

func findAndAssign(master *elevData.MasterList, floor int, button int, fromIP string) {
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
	// jsonToSend, err := json.Marshal(master)
	// if err != nil {
	// 	print("Error marshalling master: ", err)
	// }
	// tcp.BroadcastMessage(string(jsonToSend), nil)
}

func findBestElevIP(master *elevData.MasterList) string {
	numRequests := make(map[string]int, len(master.Elevators))
	for _, elevator := range master.Elevators {
		for _, floor := range elevator.Orders {
			for _, requested := range floor {
				if requested {
					numRequests[elevator.Ip]++
				}
			}
		}
	}
	var bestElevIP string = master.Elevators[0].Ip
	var bestElevVal int = 1e10
	for ip, value := range numRequests {
		if value > bestElevVal {
			bestElevVal = value
			bestElevIP = ip
		}
	}
	return bestElevIP
}
