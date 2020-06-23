package cmd

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/httpmessage"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/pkg/browser"
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

func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("no pem data in private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %w", err)
	}
	return key, nil
}

// waitForRedirect will open a url with a `redirect_to` query string param that gets handled by handler
func waitForRedirect(rawURL string, handler func(w http.ResponseWriter, r *http.Request)) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("error listening to port: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	q := u.Query()
	q.Set("redirect_to", fmt.Sprintf("http://localhost:%d/", port))
	u.RawQuery = q.Encode()

	done := make(chan bool, 1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
		done <- true
	})

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	defer server.Close()
	go server.Serve(listener)

	if err := browser.OpenURL(u.String()); err != nil {
		return fmt.Errorf("error opening url: %w", err)
	}

	<-done
	return nil
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
		baseurl := api.BackendURL(api.AuthService, channel)
		url := sdk.JoinURL(baseurl, "/login?apikey=true")

		err := waitForRedirect(url, func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			config.APIKey = q.Get("apikey")
			config.CustomerID = q.Get("customer_id")
			config.Expires = time.Now().Add(time.Hour * 23)
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

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
}
