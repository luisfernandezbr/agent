package dev

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pinpt/agent/internal/util"
	"github.com/pinpt/agent/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/hash"
	"github.com/pinpt/go-common/v10/log"
	pnum "github.com/pinpt/go-common/v10/number"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

func signFile(filename string, privateKey *rsa.PrivateKey) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("error openning bundle: %w", err)
	}
	defer f.Close()
	sum, err := hash.Sha256Checksum(f)
	if err != nil {
		return "", fmt.Errorf("error creating checksum: %w", err)
	}
	sigBuf, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, sum)
	if err != nil {
		return "", fmt.Errorf("error signing checksum of bundle: %w", err)
	}
	return hex.EncodeToString(sigBuf), nil
}

// PublishCmd represents the publish command
var PublishCmd = &cobra.Command{
	Use:   "publish <integration dir>",
	Short: "publish an integration to the registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationDir := args[0]
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		channel, _ := cmd.Flags().GetString("channel")
		tmpdir, err := ioutil.TempDir("", "")
		if err != nil {
			log.Fatal(logger, "error creating temp dir", "err", err)
		}
		defer os.RemoveAll(tmpdir)
		c, err := loadDevConfig(channel)
		if err != nil {
			log.Fatal(logger, "error opening developer config", "err", err)
		}
		if c.PrivateKey == "" {
			log.Fatal(logger, "missing private key in config, please enroll before publishing")
		}
		if c.expired() {
			log.Fatal(logger, "your login session has expired. please login again")
		}
		privateKey, err := util.ParsePrivateKey(c.PrivateKey)
		if err != nil {
			log.Fatal(logger, "unable to parse private key in config")
		}
		log.Info(logger, "building package")
		cm := exec.Command(os.Args[0], "package", integrationDir, "--dir", tmpdir, "--channel", channel)
		cm.Stdout = os.Stdout
		cm.Stderr = os.Stderr
		cm.Stdin = os.Stdin
		if err := cm.Run(); err != nil {
			log.Fatal(logger, "error running package command", "err", err)
		}
		bundle := filepath.Join(tmpdir, "bundle.zip")
		if !fileutil.FileExists(bundle) {
			log.Fatal(logger, "error bundle does not exist", "err", err)
		}
		signature, err := signFile(bundle, privateKey)
		if err != nil {
			log.Fatal(logger, "error getting signature for bundle", "err", err)
		}
		of, err := os.Open(bundle)
		if err != nil {
			log.Fatal(logger, "error opening bundle", "err", err)
		}
		defer of.Close()
		stat, _ := os.Stat(bundle)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		opts := []api.WithOption{
			api.WithContentType("application/zip"),
			api.WithHeader("x-pinpt-signature", signature),
			func(req *http.Request) error {
				req.ContentLength = stat.Size()
				return nil
			},
		}
		apikey, _ := cmd.Flags().GetString("apikey")
		secret, _ := cmd.Flags().GetString("secret")
		if secret != "" {
			opts = append(opts, api.WithHeader("x-api-key", secret))
		} else if apikey == "" {
			apikey = c.APIKey
			if apikey == "" {
				log.Fatal(logger, "you must login or provide the apikey using --apikey before continuing")
			}
		}
		descriptorFn := filepath.Join(integrationDir, "integration.yaml")
		descriptorBuf, err := ioutil.ReadFile(descriptorFn)
		if err != nil {
			log.Fatal(logger, "error reading descriptor", "err", err, "file", descriptorFn)
		}
		descriptor, err := sdk.LoadDescriptor(base64.StdEncoding.EncodeToString(descriptorBuf), "", "")
		if err != nil {
			log.Fatal(logger, "error loading descriptor", "err", err, "file", descriptorFn)
		}
		version := getBuildCommitForIntegration(integrationDir)
		basepath := fmt.Sprintf("publish/%s/%s/%s", c.PublisherRefType, descriptor.RefType, version)
		log.Info(logger, "uploading", "size", pnum.ToBytesSize(stat.Size()))
		resp, err := api.Put(ctx, channel, api.RegistryService, basepath, apikey, of, opts...)
		if err != nil || resp.StatusCode != http.StatusAccepted {
			var buf []byte
			if resp != nil {
				buf, _ = ioutil.ReadAll(resp.Body)
			}
			log.Fatal(logger, "error publishing your bundle", "err", err, "body", string(buf))
		}
		log.Info(logger, "🚀 published", "integration", descriptor.RefType, "version", version)
	},
}

func init() {
	PublishCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
	PublishCmd.Flags().String("apikey", "", "api key")
	PublishCmd.Flags().String("secret", "", "internal shared secret")
	PublishCmd.Flags().MarkHidden("secret")
}
