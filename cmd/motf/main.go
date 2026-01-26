package main

import (
	"os"

	"github.com/TechnicallyJoe/terraform-motf/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
