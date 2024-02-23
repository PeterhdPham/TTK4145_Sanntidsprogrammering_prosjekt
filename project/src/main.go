package main

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
	"time"
)

func main() {

	num_floors := 4

	test := make([]elevData.ElevLight, num_floors)

	fmt.Println(test)

	fmt.Println("Booting elevator") // just to know we're running

	elevio.Init("localhost:15657", num_floors) // connect to elevatorsimulator

	my_status := make(chan elevData.ElevStatus) // need these for testing
	my_dir := make(chan elevio.MotorDirection)
	my_door := make(chan bool)

	go elevData.GetLivedata(my_status, my_dir, my_door) // testing this

	go func() { // these need to be updated when changing motordirection or door open/close()
		for { // changing manually just to observe the statuschange
			time.Sleep(time.Second * 5)

			my_dir <- elevio.MD_Up
			my_door <- true

			time.Sleep(time.Second * 5)

			my_dir <- elevio.MD_Down
			my_door <- false
		}
	}()

	for status_update := range my_status {
		bytestream := elevData.StatusToBytestream(status_update)
		fmt.Println("New updated status: ", string(bytestream))
	}

}
