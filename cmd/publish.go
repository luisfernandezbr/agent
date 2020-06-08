package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pnum "github.com/pinpt/go-common/v10/number"
	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish <integration dir>",
	Short: "publish an integration to the registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationDir := args[0]
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		tmpdir := os.TempDir()
		defer os.RemoveAll(tmpdir)
		log.Info(logger, "building package")
		c := exec.Command(os.Args[0], "package", integrationDir, "--dir", tmpdir)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Run(); err != nil {
			os.Exit(1)
		}
		bundle := filepath.Join(tmpdir, "bundle.zip")
		if !fileutil.FileExists(bundle) {
			os.Exit(1)
		}
		of, err := os.Open(bundle)
		if err != nil {
			log.Fatal(logger, "error opening bundle", "err", err)
		}
		defer of.Close()
		stat, _ := os.Stat(bundle)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		channel, _ := cmd.Flags().GetString("channel")
		opts := []api.WithOption{
			api.WithContentType("application/zip"),
			api.WithHeader("Content-Length", strconv.Itoa(int(stat.Size()))),
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
			c, err := loadDevConfig()
			if err != nil {
				log.Fatal(logger, "error opening developer config", "err", err)
			}
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
		basepath := fmt.Sprintf("publish/%s/%s/%s", descriptor.Publisher.Identifier, descriptor.RefType, version)
		log.Info(logger, "uploading", "size", pnum.ToBytesSize(stat.Size()))
		resp, err := api.Put(ctx, channel, api.RegistryService, basepath, apikey, of, opts...)
		if err != nil || resp.StatusCode != http.StatusAccepted {
			buf, _ := ioutil.ReadAll(resp.Body)
			log.Fatal(logger, "error publishing your bundle", "err", err, "body", string(buf))
		}
		log.Info(logger, "🚀 published", "integration", descriptor.RefType, "version", version)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().String("channel", "stable", "the channel which can be set")
	publishCmd.Flags().String("apikey", "", "api key")
	publishCmd.Flags().String("secret", "", "internal shared secret")
	publishCmd.Flags().MarkHidden("secret")
}
