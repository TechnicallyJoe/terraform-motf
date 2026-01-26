package main

import (
	"os"

	"github.com/TechnicallyJoe/terraform-motf/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
