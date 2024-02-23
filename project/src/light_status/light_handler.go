package light_status

import (
	"Driver-go/elevio"
	"fmt"
	"math/rand"
	"time"
)

type LightStatus struct {
	HallLightUp   []bool `json:"HallLightUp"`
	HallLightDown []bool `json:"HallLightDown"`
	CabLight      []bool `json:"CabLight"`
}

func InitLights(NumberOfFloors int) LightStatus {
	hallUp := make([]bool, NumberOfFloors)
	hallDown := make([]bool, NumberOfFloors)
	cab := make([]bool, NumberOfFloors)

	return LightStatus{
		HallLightUp:   hallUp,
		HallLightDown: hallDown,
		CabLight:      cab,
	}
}

func UpdateLights(lightStatus LightStatus) { // Pass LightStatus as an argument
	// Iterating through hall_light_up and changing the lights based on the values
	for i, val := range lightStatus.HallLightUp {
		elevio.SetButtonLamp(0, i, val)
	}

	// Iterating through hall_light_down and changing the lights based on the values
	for i, val := range lightStatus.HallLightDown {
		elevio.SetButtonLamp(1, i, val)
	}

	// Iterating through cab_light and changing the lights based on the values
	for i, val := range lightStatus.CabLight {
		elevio.SetButtonLamp(2, i, val)
	}
}

func RandomizeLights(NumberOfFloors int, updateChan chan<- LightStatus) {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator

	for {
		lightStatus := InitLights(NumberOfFloors) // Reinitialize or create a new LightStatus

		// Randomly update the light status
		for i := range lightStatus.HallLightUp {
			lightStatus.HallLightUp[i] = rand.Intn(2) == 1
		}
		for i := range lightStatus.HallLightDown {
			lightStatus.HallLightDown[i] = rand.Intn(2) == 1
		}
		for i := range lightStatus.CabLight {
			lightStatus.CabLight[i] = rand.Intn(2) == 1
		}

		updateChan <- lightStatus // Send the new LightStatus through the channel

		time.Sleep(1 * time.Second)
	}
}

func ContinuousUpdate(updateChan <-chan LightStatus) {
	for newLightStatus := range updateChan {
		UpdateLights(newLightStatus)                     // Apply the update to the lights
		fmt.Println("New light status:", newLightStatus) // Optional: print the new status for verification
	}
}
