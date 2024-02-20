package main

import (
	"Driver-go/elevio"
	"fmt"
	// "time"
)

func main() {

	num_floors := 4

	fmt.Println("Booting elevator")

	elevio.Init("localhost:15657", num_floors)

	elev_data.get_livedata()

}
