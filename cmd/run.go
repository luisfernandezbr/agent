package cmd

import (
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

func startIntegration(logger log.Logger, channel string, dir string, publisher string, integration string, version string) (*exec.Cmd, error) {
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
	cm := exec.Command(integrationExecutable, "--channel", channel)
	cm.Stdout = os.Stdout
	cm.Stderr = os.Stderr
	cm.Stdin = os.Stdin
	err := cm.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting integration %s: %w", longName, err)
	}
	log.Debug(logger, "started integration", "pid", cm.Process.Pid)
	return cm, nil
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <integration> <version>",
	Short: "run a published integration",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		fullIntegration := args[0]
		version := args[1]
		tok := strings.Split(fullIntegration, "/")
		if len(tok) != 2 {
			log.Fatal(logger, "integration should be in the format: publisher/integration such as pinpt/github")
		}
		publisher := tok[0]
		integration := tok[1]
		channel, _ := cmd.Flags().GetString("channel")
		dir := "."

		c, err := startIntegration(logger, channel, dir, publisher, integration, version)
		if err != nil {
			log.Fatal(logger, "error starting integration", "err", err)
		}

		done := make(chan bool)
		pos.OnExit(func(_ int) {
			log.Info(logger, "shutdown")
			c.Process.Kill()
			done <- true
		})
		<-done
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("channel", "stable", "the channel which can be set")
	runCmd.Flags().StringP("dir", "d", "", "directory inside of which to run the integration")
}
