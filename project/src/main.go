package main

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"project/movement"
	// "time"
)

func main() {

	num_floors := 4

	requests := make(chan []int)

	var topRequest int

	fmt.Println("Current requests: ", requests)

	fmt.Println("Booting elevator") // just to know we're running

	elevio.Init("localhost:15657", num_floors) // connect to elevatorsimulator

	my_status := make(chan elevData.ElevStatus) // need these for testing
	my_dir := make(chan elevio.MotorDirection)
	my_door := make(chan bool)

	go func(){
		for door := range my_door {
			fmt.Println(door)
		}
	}()


	go elevData.GetLivedata(my_status, my_dir, my_door) // testing this

	go movement.FulfillRequests(requests, my_status, my_dir, my_door)

	for {
		fmt.Scan(&topRequest)
		if topRequest == 0{
			fmt.Println(<-my_status)
		}else{
			slice := []int{topRequest}
			requests <- slice
		}
	}
}
