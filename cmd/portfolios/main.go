package main

import (
	"os"

	"github.com/lenon/portfolios/cmd/portfolios/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
