// Package failprint provides error printing to stderr.
package failprint

import (
	"fmt"
	"os"
)

type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota
	ExitOnError                   // Call os.Exit(2) if error received
)

// PrintFail writes error string to os.Stderr.
func Println(err error) {
	fmt.Fprintln(os.Stderr, err)
}

// PrintFailWorker call PrintFail when error received.
// If ExitOnError is provided and error received,
// can call os.Exit(2) after closing channel.
func PrintFailWorker(errStream <-chan error, errorHandling ErrorHandling) {
	var receiveErr bool
	for err := range errStream {
		receiveErr = true
		Println(err)
	}
	switch errorHandling {
	case ContinueOnError:
	case ExitOnError:
		if receiveErr {
			os.Exit(2)
		}
	}
}
