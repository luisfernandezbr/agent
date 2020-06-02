package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	export "github.com/pinpt/agent.next/internal/export/dev"
	manager "github.com/pinpt/agent.next/internal/manager/dev"
	"github.com/pinpt/agent.next/internal/pipe/console"
	"github.com/pinpt/agent.next/internal/pipe/file"
	state "github.com/pinpt/agent.next/internal/state/file"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/fileutil"
	"github.com/pinpt/go-common/log"
	"github.com/spf13/cobra"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "run an integration in development mode",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationDir := args[0]
		_logger := log.NewCommandLogger(cmd)
		defer _logger.Close()
		integrationDir, _ = filepath.Abs(integrationDir)
		integration := strings.Replace(filepath.Base(integrationDir), "agent.next.", "", -1)
		fp := filepath.Join(integrationDir, "integration.go")
		if !fileutil.FileExists(fp) {
			log.Fatal(_logger, "couldn't find the integration at "+fp)
		}
		distDir := filepath.Join(os.TempDir(), "agent.next")
		os.MkdirAll(distDir, 0700)
		dist := filepath.Join(distDir, integration+".so")
		logger := log.With(_logger, "pkg", integration)

		// build our integration
		c := exec.Command(os.Args[0], "build", "--bundle=false", "--dir", distDir, integrationDir)
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
		configkv := make(map[string]interface{})
		arr, _ := cmd.Flags().GetStringSlice("config")
		for _, val := range arr {
			tok := strings.Split(val, "=")
			configkv[strings.TrimSpace(tok[0])] = strings.TrimSpace(tok[1])
		}
		config := sdk.NewConfig(configkv)
		log.Info(_logger, "starting")
		mgr := manager.New(logger)
		if err := instance.Start(logger, config, mgr); err != nil {
			log.Fatal(logger, "failed to start", "err", err)
		}
		var pipe sdk.Pipe
		outdir, _ := cmd.Flags().GetString("dir")
		if outdir != "" {
			os.MkdirAll(outdir, 0700)
			pipe = file.New(logger, outdir)
		} else {
			pipe = console.New(logger)
		}
		jobid, _ := cmd.Flags().GetString("jobid")
		customerid, _ := cmd.Flags().GetString("customerid")
		statedir, _ := cmd.Flags().GetString("state")
		if statedir == "" {
			statedir = outdir
		}
		statefn := filepath.Join(statedir, "state.json")
		stateobj, err := state.New(statefn)
		if err != nil {
			log.Fatal(logger, "error opening state file", "err", err)
		}
		completion := make(chan export.Completion, 1)
		exp, err := export.New(logger, config, stateobj, jobid, customerid, pipe, completion)
		if err != nil {
			log.Fatal(logger, "export failed", "err", err)
		}
		started := time.Now()
		if err := instance.Export(exp); err != nil {
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
	devCmd.Flags().String("dir", "", "the directory to output pipe contents")
	devCmd.Flags().String("state", "", "the state file directory")
}
