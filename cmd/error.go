package cmd

import (
	"fmt"
	"os"
)

const (
	// http://tldp.org/LDP/abs/html/exitcodes.html
	ExitSuccess = iota
	ExitError
	ExitBadConnection
	ExitInvalidInput
	ExitBadFeature
	ExitInterrupted
	ExitIO
	ExitBadArgs = 128
)

func ExitWithError(code int, err error) {
	fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[0m\n", err)
	os.Exit(code)
}
