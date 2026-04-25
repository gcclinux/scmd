package util

import (
	"fmt"
	"sync"
	"time"
)

var (
	spinnerMu   sync.Mutex
	spinnerStop chan struct{}
	isSpinning  bool
)

// StartSpinner starts an animated "Thinking..." loading indicator.
func StartSpinner() {
	spinnerMu.Lock()
	defer spinnerMu.Unlock()
	if isSpinning {
		return
	}
	isSpinning = true
	spinnerStop = make(chan struct{})

	go func() {
		state := 0
		for {
			select {
			case <-spinnerStop:
				// Clear the line when stopping
				fmt.Print("\r\033[K")
				return
			default:
				var msg string
				var sleep time.Duration
				switch state {
				case 0:
					msg = "\rThinking.  "
					sleep = 500 * time.Millisecond
				case 1:
					msg = "\rThinking.. "
					sleep = 500 * time.Millisecond
				case 2:
					msg = "\rThinking..."
					sleep = 1000 * time.Millisecond
				}
				fmt.Print(msg)
				state = (state + 1) % 3

				select {
				case <-spinnerStop:
					fmt.Print("\r\033[K")
					return
				case <-time.After(sleep):
				}
			}
		}
	}()
}

// StopSpinner stops the animated loading indicator.
func StopSpinner() {
	spinnerMu.Lock()
	defer spinnerMu.Unlock()
	if !isSpinning {
		return
	}
	close(spinnerStop)
	isSpinning = false
}
