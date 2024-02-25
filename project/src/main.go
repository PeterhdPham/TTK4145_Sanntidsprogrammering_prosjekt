package main

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"project/movement"
	"time"
)

func main() {

	num_floors := 4

	requests := []int{3, 0, 2, 1}

	fmt.Println("Current requests: ", requests)

	fmt.Println("Booting elevator") // just to know we're running

	elevio.Init("localhost:15657", num_floors) // connect to elevatorsimulator

	my_status := make(chan elevData.ElevStatus) // need these for testing
	my_dir := make(chan elevio.MotorDirection)
	my_door := make(chan bool)

	go elevData.GetLivedata(my_status, my_dir, my_door) // testing this

	for {
		go movement.FulfillRequest(requests, my_status, my_dir, my_door)
		time.Sleep(time.Second * 10)
		requests = requests[1:]
		fmt.Println("Current requests: ", requests)
		time.Sleep(time.Second * 5)
	}

}
