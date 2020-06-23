package dev

import (
	"github.com/spf13/cobra"
)

// LogoutCmd represents the logout command
var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout of your developer account",
	Run: func(cmd *cobra.Command, args []string) {
		var c devConfig
		c.remove()
	},
}

func init() {
	// add command to root in ../dev.go
}
