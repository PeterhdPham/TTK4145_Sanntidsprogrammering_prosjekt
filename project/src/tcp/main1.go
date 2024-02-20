package tcp

// import (
// 	"fmt"
// 	"os"
// 	"os/exec"
// 	"project/pack" // Assuming your TCP and IP finding logic are in this package
// 	"runtime"
// 	"sync"
// )

// func main1() {
// 	var wg sync.WaitGroup
// 	lowestIPChan := make(chan string) // Channel to receive the lowest IP address

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		living_IPs := make(chan []string)

// 		// Assuming Broadcast_life and Look_for_life start the process to find IPs
// 		go pack.Broadcast_life()          // Start broadcasting
// 		go pack.Look_for_life(living_IPs) // Start looking for life

// 		// Use the findLowestIPAddress logic to find and set the lowest IP
// 		FindLowestIPAddress(living_IPs, lowestIPChan)
// 	}()

// 	// Receive the lowest IP address dynamically
// 	go func() {
// 		for lowestIP := range lowestIPChan {
// 			fmt.Println("Lowest IP Address:", lowestIP)
// 			// Set the serverAddress for the TCP operations
// 			SetServerAddress(lowestIP)
// 			// Start or restart TCP operations with the new server address
// 			NewTCP()
// 		}
// 	}()

// 	wg.Wait() // Wait for all operations to finish
// }

// func clearScreen() {
// 	var cmd *exec.Cmd
// 	switch runtime.GOOS {
// 	case "windows":
// 		cmd = exec.Command("cmd", "/c", "cls")
// 	default:
// 		cmd = exec.Command("clear") // Works on Unix-like systems
// 	}
// 	cmd.Stdout = os.Stdout
// 	cmd.Run()
// }
