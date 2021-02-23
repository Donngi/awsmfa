package cmd

import (
	"fmt"
	"os"
)

func printErrorRed(err error) {
	fmt.Fprintln(os.Stderr, fmt.Errorf("\x1b[31m[ERROR]: %w\x1b[0m", err))
}

func printBlue(str string) {
	fmt.Printf("\x1b[34m%s\x1b[0m", str)
}

func printCyan(str string) {
	fmt.Printf("\x1b[36m%s\x1b[0m", str)
}
