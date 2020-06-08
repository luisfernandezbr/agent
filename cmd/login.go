package cmd

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/99designs/keyring"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/httpmessage"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func getKeyRing() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: "pinpoint",
	})
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to your developer account",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()

		ring, err := getKeyRing()
		if err != nil {
			log.Fatal(logger, "error opening key chain", "err", err)
		}

		channel, _ := cmd.Flags().GetString("channel")

		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal(logger, "error listening to port", "err", err)
		}

		port := listener.Addr().(*net.TCPAddr).Port

		done := make(chan bool, 1)
		var customerid string

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if err := ring.Set(keyring.Item{
				Key:  "apikey",
				Data: []byte(q.Get("apikey")),
			}); err != nil {
				log.Error(logger, "error saving apikey to keychain", "err", err)
			}
			if err := ring.Set(keyring.Item{
				Key:  "customer_id",
				Data: []byte(q.Get("customer_id")),
			}); err != nil {
				log.Error(logger, "error saving customer_id to keychain", "err", err)
			}
			customerid = q.Get("customer_id")
			httpmessage.RenderStatus(w, r, http.StatusOK, "Login Success", "You have logged in successfully and can now close this window")
			done <- true
		})

		server := &http.Server{
			Addr: fmt.Sprintf(":%d", port),
		}
		defer server.Close()
		go server.Serve(listener)

		_ = channel

		baseurl := api.BackendURL(api.AuthService, channel)
		url := sdk.JoinURL(baseurl, "/login?apikey=true&redirect_to="+url.QueryEscape(fmt.Sprintf("http://localhost:%d/", port)))

		if err := browser.OpenURL(url); err != nil {
			log.Fatal(logger, "error opening url", "err", err)
		}

		<-done
		log.Info(logger, "logged in", "customer_id", customerid)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("channel", "stable", "the channel which can be set")
}
