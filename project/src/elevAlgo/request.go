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

func requestClearAtFloor(myStatus elevData.ElevStatus, myOrders [][]bool, floor int) (elevData.ElevStatus, [][]bool) {
	fmt.Println("Request Clear")

	switch myStatus.Direction {
	case 1:
		// if !requestsAbove(e) && !myOrders[floor][0] {
		// 	myOrders[floor][1] = false
		// }
		myOrders[floor][0] = false

	case -1:
		// if !requestsBelow(e) && !myOrders[floor][1] {
		// 	myOrders[floor][0] = false
		// }
		myOrders[floor][1] = false

	case 0:
		fallthrough
	default:
		myOrders[floor][0] = false
		myOrders[floor][1] = false
	}

	return myStatus, myOrders

}
