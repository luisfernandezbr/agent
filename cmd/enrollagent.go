package cmd

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pinpt/agent.next/runner"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/agent.next/sysinfo"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/graphql"
	pjson "github.com/pinpt/go-common/v10/json"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/pinpt/integration-sdk/agent"
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
		config.Channel = channel

		url := sdk.JoinURL(api.BackendURL(api.AppService, channel), "/enroll")

		var userID string
		err := waitForRedirect(url, func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			config.APIKey = q.Get("apikey")
			config.CustomerID = q.Get("customer_id")
			userID = q.Get("user_id")
			w.WriteHeader(http.StatusOK)
		})
		if err != nil {
			log.Fatal(logger, "error waiting for browser", "err", err)
		}
		log.Info(logger, "logged in", "customer_id", config.CustomerID)

		log.Info(logger, "enrolling agent...", "customer_id", config.CustomerID)
		client, err := graphql.NewClient(config.CustomerID, "", "", api.BackendURL(api.GraphService, channel))
		if err != nil {
			log.Fatal(logger, "error creating graphql client", "err", err)
		}
		client.SetHeader("Authorization", config.APIKey)
		info, err := sysinfo.GetSystemInfo()
		if err != nil {
			log.Fatal(logger, "error getting system info", "err", err)
		}
		config.SystemID = info.ID
		created := agent.EnrollmentCreatedDate(datetime.NewDateNow())
		enr := agent.Enrollment{
			AgentVersion: commit, // TODO(robin): when we start versioning, switch this to version
			CreatedDate:  created,
			SystemID:     info.ID,
			Hostname:     info.Hostname,
			NumCPU:       info.NumCPU,
			OS:           info.OS,
			Architecture: info.Architecture,
			GoVersion:    info.GoVersion,
			CustomerID:   config.CustomerID,
			UserID:       userID,
		}
		if err := agent.CreateEnrollment(client, enr); err != nil {
			if strings.Contains(err.Error(), "duplicate key error") {
				// TODO(robin): find out how someone might fix this and then update the error message with an action to take
				log.Fatal(logger, "this system has already been enrolled")
			}
			log.Fatal(logger, "error creating enrollment", "err", err)
		}
		fn := configFilename(cmd)
		os.MkdirAll(filepath.Dir(fn), 0700)
		if err := ioutil.WriteFile(fn, []byte(pjson.Stringify(config)), 0644); err != nil {
			log.Fatal(logger, "error writing config file", "err", err)
		}
		log.Info(logger, "agent enrolled 🎉", "customer_id", config.CustomerID)
	},
}

func init() {
	rootCmd.AddCommand(enrollAgentCmd)
	enrollAgentCmd.Flags().String("config", "", "the location of the config file")
	enrollAgentCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
}
