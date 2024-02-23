package main

import (
	"Driver-go/elevio"
	"fmt"
	"project/light_status"
)

func main() {
	num_floors := 4

	fmt.Println("Booting elevator")
	elevio.Init("localhost:15657", num_floors)

	updateLigthChan := make(chan light_status.LightStatus) // Create a channel for light status updates

	// Start continuously updating light status based on channel updates
	go light_status.ContinuousUpdate(updateLigthChan)

	// Send the initial light status with all lights off through the channel
	updateLigthChan <- light_status.InitLights(num_floors)
	
	select {}
}
