package main

import (
	"os"

	"github.com/TechnicallyJoe/sturdy-parakeet/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
