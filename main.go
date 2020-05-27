package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:    "majin",
	Hidden: true,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("majin error: %s\n", err)
	}
}
