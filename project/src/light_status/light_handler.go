package light_status

import (
	"Driver-go/elevio"
	"fmt"
	"math/rand"
	"time"
)

type LightStatus struct {
	hall_light_up   []bool `json:"hall_light_up"`
	hall_light_down []bool `json:"hall_light_down"`
	cab_light       []bool `json:"cab_light"`
}

func Init_Lights(number_of_floors int) LightStatus {
	hallUp := make([]bool, number_of_floors)
	hallDown := make([]bool, number_of_floors)
	cab := make([]bool, number_of_floors)

	return LightStatus{
		hall_light_up:   hallUp,
		hall_light_down: hallDown,
		cab_light:       cab,
	}
}

func Update_Lights(lightStatus LightStatus) { // Pass LightStatus as an argument
	// Iterating through hall_light_up and changing the lights based on the values
	for i, val := range lightStatus.hall_light_up {
		elevio.SetButtonLamp(0, i, val)
	}

	// Iterating through hall_light_down and changing the lights based on the values
	for i, val := range lightStatus.hall_light_down {
		elevio.SetButtonLamp(1, i, val)
	}

	// Iterating through cab_light and changing the lights based on the values
	for i, val := range lightStatus.cab_light {
		elevio.SetButtonLamp(2, i, val)
	}
}

func RandomizeLights(number_of_floors int, updateChan chan<- LightStatus) {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator

	for {
		lightStatus := Init_Lights(number_of_floors) // Reinitialize or create a new LightStatus

		// Randomly update the light status
		for i := range lightStatus.hall_light_up {
			lightStatus.hall_light_up[i] = rand.Intn(2) == 1
		}
		for i := range lightStatus.hall_light_down {
			lightStatus.hall_light_down[i] = rand.Intn(2) == 1
		}
		for i := range lightStatus.cab_light {
			lightStatus.cab_light[i] = rand.Intn(2) == 1
		}

		updateChan <- lightStatus // Send the new LightStatus through the channel

		time.Sleep(1 * time.Second)
	}
}

func ContinuousUpdate(updateChan <-chan LightStatus) {
	for newLightStatus := range updateChan {
		Update_Lights(newLightStatus)                    // Apply the update to the lights
		fmt.Println("New light status:", newLightStatus) // Optional: print the new status for verification
	}
}
