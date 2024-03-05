package elevalgo

import (
	"fmt"
	"project/elevData"
)

var N_BUTTONS = 3

func areAllOrdersFalse(orders [][]bool) bool {
	for _, floor := range orders {
		for _, order := range floor {
			if order {
				// Found an order (true), so not all values are false.
				return false
			}
		}
	}
	// If we get here, it means all values were false.
	return true
}
func requestShouldStop(status elevData.ElevStatus, orders [][]bool, floor int) bool {
	if areAllOrdersFalse(orders) {
		fmt.Println("No orders found")
		return true
	}
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
