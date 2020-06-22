package cmd

import (
	"net/http"

	"github.com/pinpt/agent.next/runner"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

// enrollAgentCmd represents the login command
var enrollAgentCmd = &cobra.Command{
	Use:   "enroll-agent", // TODO(robin): move dev commands into their own module and rename this enroll
	Short: "connect this agent to Pinpoint's backend",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		channel, _ := cmd.Flags().GetString("channel")

		var config runner.ConfigFile

		url := sdk.JoinURL(api.BackendURL(api.AppService, channel), "/enroll")

		err := waitForRedirect(url, func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			config.APIKey = q.Get("apikey")
			config.CustomerID = q.Get("customer_id")
			w.WriteHeader(http.StatusOK)
		})
		if err != nil {
			log.Fatal(logger, "error waiting for browser", "err", err)
		}
		log.Info(logger, "logged in", "customer_id", config.CustomerID)

	},
}

func init() {
	rootCmd.AddCommand(enrollAgentCmd)
	enrollAgentCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
}
