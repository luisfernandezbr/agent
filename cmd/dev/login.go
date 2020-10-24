package dev

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pinpt/agent/v4/internal/util"
	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/httpmessage"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

// TODO (Pedro): A lot of this code is very similar to the one in cmd/dev - we should try to consolidate it

type devConfig struct {
	CustomerID       string    `json:"customer_id"`
	APIKey           string    `json:"apikey"`
	RefreshKey       string    `json:"refresh_token"`
	PrivateKey       string    `json:"private_key"`
	Certificate      string    `json:"certificate"`
	Expires          time.Time `json:"expires"`
	Channel          string    `json:"channel"`
	PublisherRefType string    `json:"publisher_ref_type"`
	Logger           sdk.Logger
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
	if !fileutil.FileExists(fn) {
		return nil, fmt.Errorf("config file %s does not exist", fn)
	}
	of, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", fn, err)
	}
	defer of.Close()
	if err := json.NewDecoder(of).Decode(&c); err != nil {
		return nil, fmt.Errorf("error decoding %s: %w", fn, err)
	}

	valid, err := validateConfig(&c, channel)
	if err != nil {
		return &c, fmt.Errorf("error vlidating config file. err %v", err)
	}
	if !valid {
		return &c, errors.New("config file no longer valid")
	}

	return &c, nil
}

func validateConfig(config *devConfig, channel string) (bool, error) {
	var resp struct {
		Expired bool `json:"expired"`
		Valid   bool `json:"valid"`
	}
	res, err := api.Get(context.Background(), channel, api.AuthService, "/validate?customer_id="+config.CustomerID, config.APIKey)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return false, err
	}
	if !resp.Expired {
		return true, nil
	}
	if resp.Expired {
		newToken, err := util.RefreshOAuth2Token(http.DefaultClient, channel, "pinpoint", config.RefreshKey)
		if err != nil {
			return false, err
		}
		config.APIKey = newToken // update the new token
		return true, nil
	}
	if resp.Valid {
		return false, fmt.Errorf("the apikey or refresh token is no longer valid")
	}
	return false, nil
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
		url := sdk.JoinURL(baseurl, "/oauth2/pinpoint/agent/authorize?apikey=true")

		var ok bool
		var errmsg string

		err := util.WaitForRedirect(url, func(w http.ResponseWriter, r *http.Request) {
			config, _ = loadDevConfig(channel)
			q := r.URL.Query()
			err := q.Get("error")
			if err != "" {
				_errdesc := q.Get("error_description")
				if _errdesc == "" || _errdesc == "undefined" {
					_errdesc = err
				}
				errmsg = _errdesc
				httpmessage.RenderStatus(w, r, http.StatusUnauthorized, "Login Failed", "Login failed. "+_errdesc)
				return
			}
			customerID := q.Get("customer_id")
			if customerID == "" {
				log.Fatal(logger, "the authorization server didn't return a valid customer_id")
			}
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
			config.RefreshKey = q.Get("refresh_token")
			config.CustomerID = customerID
			config.Channel = channel
			if err := config.save(); err != nil {
				log.Error(logger, "error saving config", "err", err)
			}
			ok = true
			httpmessage.RenderStatus(w, r, http.StatusOK, "Login Success", "You have logged in successfully and can now close this window")
		})
		if err != nil {
			log.Fatal(logger, "error waiting for browser", "err", err)
		}
		if !ok {
			log.Fatal(logger, "error logging in", "err", errmsg)
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
