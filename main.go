package main

import (
	"os"

	"github.com/mikowitz/greyskull/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}