package cmd

import (
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout of your developer account",
	Run: func(cmd *cobra.Command, args []string) {
		var c devConfig
		c.remove()
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
