package elevalgo

import (
	"time"
)

var (
	timerEndTime time.Time
	timerActive  bool
	timerChannel chan bool // Channel to signal the timeout event
)

func init() {
	// Initialize the timerChannel
	timerChannel = make(chan bool, 1) // Buffered channel to prevent blocking on send
}

// timerStart starts the timer with a specified duration in seconds.
func timerStart(duration time.Duration) {
	timerEndTime = time.Now().Add(duration)
	timerActive = true
	go func() {
		time.Sleep(duration)
		if timerActive && time.Now().After(timerEndTime) {
			timerChannel <- true // Signal that the timer has expired
		}
	}()
}

func timerStop() {
	timerActive = false
	// Optionally, clear the channel if stopping prematurely
	select {
	case <-timerChannel:
	default:
	}
}

// timerTimedOut checks if the timer has timed out.
func timerTimedOut() bool {
	return timerActive && time.Now().After(timerEndTime)
}
