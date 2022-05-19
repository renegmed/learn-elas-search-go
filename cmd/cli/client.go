package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client", // this is a sub-command
	Short: "Client services",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Client services")
	},
}
