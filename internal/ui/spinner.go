// Package ui provides terminal UI helpers.
package ui

import (
	"fmt"
	"os"
	"time"
)

var frames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// Spin displays an animated spinner with msg on stderr.
// Call the returned stop func with the operation error when done:
// on success the line is cleared; on error it morphs into the error.
func Spin(msg string) func(error) {
	done := make(chan struct{})
	go func() {
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%c %s...", frames[i%len(frames)], msg)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	return func(err error) {
		close(done)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\r✗ %s\n", err)
		} else {
			fmt.Fprint(os.Stderr, "\r\033[K")
		}
	}
}
