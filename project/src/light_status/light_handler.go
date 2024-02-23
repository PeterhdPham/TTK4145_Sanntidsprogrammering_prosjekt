package light_status

import (
	"Driver-go/elevio"
	"fmt"
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
	// Iterating through hall_light_up and printing its values
	fmt.Println("Hall Light Up:")
	for i, val := range lightStatus.hall_light_up {
		elevio.SetButtonLamp(0, i, val)
	}

	// Iterating through hall_light_down and printing its values
	fmt.Println("Hall Light Down:")
	for i, val := range lightStatus.hall_light_down {
		elevio.SetButtonLamp(1, i, val)
	}

	// Iterating through cab_light and printing its values
	fmt.Println("Cab Light:")
	for i, val := range lightStatus.cab_light {
		elevio.SetButtonLamp(2, i, val)
	}
}

func Light_Testing() {
	fmt.Println("Light Testing:")
	lightStatus := Init_Lights(4) // Store the returned LightStatus for use
	Update_Lights(lightStatus)    // Pass the LightStatus to Update_Lights
}
