package dev

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pinpt/agent.next/internal/util"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/httpmessage"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

type devConfig struct {
	CustomerID       string    `json:"customer_id"`
	APIKey           string    `json:"apikey"`
	PrivateKey       string    `json:"private_key"`
	Certificate      string    `json:"certificate"`
	Expires          time.Time `json:"expires"`
	Channel          string    `json:"channel"`
	PublisherRefType string    `json:"publisher_ref_type"`
}

func (c *devConfig) expired() bool {
	return c.Expires.Before(time.Now())
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
	if c.Channel == "" {
		return "", fmt.Errorf("cannot open pinpoint developer file: no channel provided")
	}
	if c.Channel != "stable" {
		return filepath.Join(home, ".pinpoint-developer-"+c.Channel), nil
	}
	return filepath.Join(home, ".pinpoint-developer"), nil
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

func loadDevConfig(channel string) (*devConfig, error) {
	var c devConfig
	c.Channel = channel
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

// LoginCmd represents the login command
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to your developer account",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()

		var config *devConfig

		channel, _ := cmd.Flags().GetString("channel")
		baseurl := api.BackendURL(api.AuthService, channel)
		url := sdk.JoinURL(baseurl, "/login?apikey=true")

		err := util.WaitForRedirect(url, func(w http.ResponseWriter, r *http.Request) {
			config, _ = loadDevConfig(channel)
			q := r.URL.Query()
			customerID := q.Get("customer_id")
			if config != nil {
				if config.CustomerID == customerID {
					log.Info(logger, "refreshing token", "customer_id", customerID)
				} else {
					log.Info(logger, "logging into new customer, you will need to generate a new private key before publishing 🔑", "customer_id", customerID)
				}
			} else {
				config = &devConfig{}
			}
			expires := q.Get("expires")
			if expires != "" {
				e, _ := strconv.ParseInt(expires, 10, 64)
				config.Expires = datetime.DateFromEpoch(e)
			} else {
				config.Expires = time.Now().Add(time.Hour * 23)
			}
			config.APIKey = q.Get("apikey")
			config.CustomerID = customerID
			config.Channel = channel
			if err := config.save(); err != nil {
				log.Error(logger, "error saving config", "err", err)
			}
			httpmessage.RenderStatus(w, r, http.StatusOK, "Login Success", "You have logged in successfully and can now close this window")
		})
		if err != nil {
			log.Fatal(logger, "error waiting for browser", "err", err)
		}

		log.Info(logger, "logged in", "customer_id", config.CustomerID)
	},
}

// testKey represents the test key command
var testKeyCmd = &cobra.Command{
	Use:    "testkey",
	Short:  "testkey will verify that your api key is good",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		channel, _ := cmd.Flags().GetString("channel")
		altApikey, _ := cmd.Flags().GetString("apikey")

		var apikey string
		if altApikey == "" {
			config, err := loadDevConfig(channel)
			if err != nil {
				log.Fatal(logger, "error loading dev config", "err", err)
			}
			if config == nil {
				log.Fatal(logger, "no dev config found")
			}
			apikey = config.APIKey
			fn, _ := config.filename()
			log.Info(logger, "using key from developer config", "config_file", fn, "customer_id", config.CustomerID, "expires", config.Expires)
		} else {
			apikey = altApikey
			log.Info(logger, "using passed in key")
		}
		resp, err := api.Get(cmd.Context(), channel, api.RegistryService, "/validate/1", apikey)
		if err != nil {
			var buf []byte
			if resp != nil {
				buf, _ = ioutil.ReadAll(resp.Body)
			}
			log.Fatal(logger, "error from api", "err", err, "body", string(buf))
		}
		if resp.StatusCode == http.StatusOK {
			log.Info(logger, "key is good! ✅")
		} else {
			log.Warn(logger, "key is bad! 🛑", "status", resp.StatusCode)
		}
	},
}

func init() {
	// add command to root in ../dev.go
	LoginCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
	testKeyCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
	testKeyCmd.Flags().String("apikey", "", "specify a different key")
	LoginCmd.AddCommand(testKeyCmd)
}
