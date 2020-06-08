package cmd

import (
	"github.com/pinpt/go-common/v10/log"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout of your developer account",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()

		ring, err := getKeyRing()
		if err != nil {
			log.Fatal(logger, "error opening key chain", "err", err)
		}

		ring.Remove("apkey")
		ring.Remove("customer_id")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
