package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/httpmessage"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

type devConfig struct {
	CustomerID string `json:"customer_id"`
	APIKey     string `json:"apikey"`
}

func (c *devConfig) remove() {
	fn, err := c.filename()
	if err == nil {
		os.Remove(fn)
	}
}

func (c *devConfig) filename() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fn := filepath.Join(home, ".pinpoint-developer")
	return fn, nil
}

func (c *devConfig) save() error {
	fn, err := c.filename()
	if err != nil {
		return err
	}
	of, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer of.Close()
	return json.NewEncoder(of).Encode(c)
}

func loadDevConfig() (*devConfig, error) {
	var c devConfig
	fn, err := c.filename()
	if err != nil {
		return nil, err
	}
	if fileutil.FileExists(fn) {
		of, err := os.Open(fn)
		if err != nil {
			return nil, fmt.Errorf("error opening %s: %w", fn, err)
		}
		defer of.Close()
		if err := json.NewDecoder(of).Decode(&c); err != nil {
			return nil, fmt.Errorf("error decoding %s: %w", fn, err)
		}
		return &c, nil
	}
	return &c, nil
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to your developer account",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()

		var config devConfig

		channel, _ := cmd.Flags().GetString("channel")

		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal(logger, "error listening to port", "err", err)
		}

		port := listener.Addr().(*net.TCPAddr).Port

		done := make(chan bool, 1)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			config.APIKey = q.Get("apikey")
			config.CustomerID = q.Get("customer_id")
			if err := config.save(); err != nil {
				log.Error(logger, "error saving config", "err", err)
			}
			httpmessage.RenderStatus(w, r, http.StatusOK, "Login Success", "You have logged in successfully and can now close this window")
			done <- true
		})

		server := &http.Server{
			Addr: fmt.Sprintf(":%d", port),
		}
		defer server.Close()
		go server.Serve(listener)

		baseurl := api.BackendURL(api.AuthService, channel)
		url := sdk.JoinURL(baseurl, "/login?apikey=true&redirect_to="+url.QueryEscape(fmt.Sprintf("http://localhost:%d/", port)))

		if err := browser.OpenURL(url); err != nil {
			log.Fatal(logger, "error opening url", "err", err)
		}

		<-done
		log.Info(logger, "logged in", "customer_id", config.CustomerID)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("channel", "stable", "the channel which can be set")
}
