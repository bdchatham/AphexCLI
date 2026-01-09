package progress

import (
	"fmt"
	"time"
)

// Spinner represents a simple text-based spinner
type Spinner struct {
	message string
	frames  []string
	stop    chan bool
	done    chan bool
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"|", "/", "-", "\\"},
		stop:    make(chan bool),
		done:    make(chan bool),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r")
				s.done <- true
				return
			default:
				fmt.Printf("\r%s %s", s.frames[i%len(s.frames)], s.message)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.stop <- true
	<-s.done
	fmt.Print("\r\033[K") // Clear the line
}

// WithSpinner runs a function with a spinner
func WithSpinner(message string, fn func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()
	defer spinner.Stop()
	
	return fn()
}
