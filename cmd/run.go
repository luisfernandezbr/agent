package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

func getIntegration(ctx context.Context, logger log.Logger, channel string, dir string, publisher string, integration string, version string) (*exec.Cmd, error) {
	longName := fmt.Sprintf("%s/%s/%s", publisher, integration, version)
	integrationExecutable, _ := filepath.Abs(filepath.Join(dir, integration))
	if !fileutil.FileExists(integrationExecutable) {
		log.Info(logger, "need to download integration", "integration", longName)
		var err error
		integrationExecutable, err = downloadIntegration(logger, channel, dir, publisher, integration, version)
		if err != nil {
			return nil, fmt.Errorf("error downloading integration %s: %w", longName, err)
		}
		log.Info(logger, "downloaded", "integration", integrationExecutable)
	}
	// NOTE: We probably dont have to inject channel
	cm := exec.CommandContext(ctx, integrationExecutable, "--channel", channel)
	cm.Stdout = os.Stdout
	cm.Stderr = os.Stderr
	cm.Stdin = os.Stdin
	return cm, nil
}

func serveCommand(logger log.Logger, c *exec.Cmd) error {
	if err := c.Start(); err != nil {
		return fmt.Errorf("error starting integration %s: %w", c.String(), err)
	}
	log.Debug(logger, "started integration", "pid", c.Process.Pid)
	go func(logger log.Logger, c *exec.Cmd) {
		status, err := c.Process.Wait()
		if err != nil {
			log.Fatal(logger, "error waiting for process to finish", "err", err)
		}
		if status.Success() {
			log.Info(logger, "integration exited sucessfully")
		} else {
			log.Error(logger, "integration exited with error code", "code", status.ExitCode())
		}
		os.Exit(status.ExitCode())
	}(logger, c)
	return nil
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <integration> <version>",
	Short: "run a published integration",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		ctx, cancel := context.WithCancel(context.Background())
		fullIntegration := args[0]
		var version string
		if len(args) > 1 {
			version = args[1]
		}
		tok := strings.Split(fullIntegration, "/")
		if len(tok) != 2 {
			log.Fatal(logger, "integration should be in the format: publisher/integration such as pinpt/github")
		}
		publisher := tok[0]
		integration := tok[1]
		channel, _ := cmd.Flags().GetString("channel")
		dir, _ := cmd.Flags().GetString("dir")

		c, err := getIntegration(ctx, logger, channel, dir, publisher, integration, version)
		if err != nil {
			log.Fatal(logger, "error creating integration exec", "err", err)
		}
		if err := serveCommand(logger, c); err != nil {
			log.Fatal(logger, "error serving command", "err", err)
		}

		done := make(chan bool)
		pos.OnExit(func(_ int) {
			log.Info(logger, "shutdown")
			cancel()
			done <- true
		})
		<-done
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
	runCmd.Flags().StringP("dir", "d", "", "directory inside of which to run the integration")
}
