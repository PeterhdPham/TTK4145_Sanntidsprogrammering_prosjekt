package main

import (
	"Driver-go/elevio"
	"fmt"
	"project/elev_data"
	"time"
)

func main() {

	num_floors := 4

	fmt.Println("Booting elevator")

	elevio.Init("localhost:15657", num_floors)

	my_status := make(chan []int)
	my_dir := make(chan elevio.MotorDirection)
	my_door := make(chan bool)

	go elev_data.Get_livedata(my_status, my_dir, my_door)
	go func() {
		for {
			time.Sleep(time.Second * 5)

			my_dir <- elevio.MD_Up
			my_door <- true

			time.Sleep(time.Second * 5)

			my_dir <- elevio.MD_Down
			my_door <- false
		}
	}()

	for status_update := range my_status {
		fmt.Println("New updated status: ", status_update)
	}

}
