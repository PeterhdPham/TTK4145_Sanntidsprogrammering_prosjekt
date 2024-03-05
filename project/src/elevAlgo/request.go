package elevalgo

import "project/elevData"

var N_BUTTONS = 3

func requestShouldStop(status elevData.ElevStatus, orders [][]bool, floor int) bool {
	switch status.Direction {
	case -1:
		if orders[floor][1] || orders[floor][2] {
			return true
		}
	case 1:
		if orders[floor][0] || orders[floor][2] {
			return true
		}
	}

	return false
}
