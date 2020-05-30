package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/pinpt/agent.next/internal/export"
	"github.com/pinpt/agent.next/internal/manager/dev"
	"github.com/pinpt/agent.next/internal/pipe/console"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/fileutil"
	"github.com/pinpt/go-common/log"
	"github.com/spf13/cobra"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:  "dev",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integration := args[0]
		_logger := log.NewCommandLogger(cmd)
		defer _logger.Close()
		logger := log.With(_logger, "pkg", integration)
		cwd, _ := os.Getwd()
		fp := filepath.Join(cwd, "integration", integration, "integration.go")
		if !fileutil.FileExists(fp) {
			log.Fatal(logger, "couldn't find the integration at "+fp)
		}
		os.MkdirAll(filepath.Join(cwd, "dist"), 0655)
		dist := filepath.Join(cwd, "dist", integration+".so")
		c := exec.Command("go", "build", "-buildmode=plugin", "-o", dist, fp)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			os.Exit(1)
		}

		plug, err := plugin.Open(dist)
		if err != nil {
			log.Fatal(logger, "couldn't open integration plugin", "err", err)
		}
		sym, err := plug.Lookup("Integration")
		if err != nil {
			log.Fatal(logger, "couldn't integration plugin entrypoint", "err", err)
		}
		instance := sym.(sdk.Integration)
		config := make(sdk.Config)
		arr, _ := cmd.Flags().GetStringSlice("config")
		for _, val := range arr {
			tok := strings.Split(val, "=")
			config[strings.TrimSpace(tok[0])] = strings.TrimSpace(tok[1])
		}
		log.Info(_logger, "starting")
		mgr := dev.New(logger)
		if err := instance.Start(logger, config, mgr); err != nil {
			log.Fatal(logger, "failed to start", "err", err)
		}
		conPipe := console.New(logger)
		jobid, _ := cmd.Flags().GetString("jobid")
		customerid, _ := cmd.Flags().GetString("customerid")
		completion := make(chan export.Completion, 1)
		export, err := export.New(logger, config, jobid, customerid, conPipe, completion)
		if err != nil {
			log.Fatal(logger, "export failed", "err", err)
		}
		started := time.Now()
		if err := instance.Export(export); err != nil {
			log.Fatal(logger, "export failed", "err", err)
		}
		done := <-completion
		if done.Error != nil {
			log.Error(logger, "error running export", "err", done.Error)
		} else {
			log.Info(logger, "export finished", "duration", time.Since(started))
		}
		if err := instance.Stop(); err != nil {
			log.Fatal(logger, "failed to stop", "err", err)
		}
		log.Info(_logger, "stopped")
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
	devCmd.Flags().StringSlice("config", []string{}, "a config key/value pair such as a=b")
	devCmd.Flags().String("jobid", "999", "job id")
	devCmd.Flags().String("customerid", "000", "customer id")
}
