package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func printErrorRed(err error) {
	red := color.New(color.FgRed).FprintfFunc()
	red(os.Stderr, fmt.Errorf("\x1b[31m[ERROR]: %w\x1b[0m", err).Error())
}

func printBlue(str string) {
	color.Blue(str)
}

func printCyan(str string) {
	color.Cyan(str)
}
