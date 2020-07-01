package dev

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

func buildIntegration(logger log.Logger, distDir, integrationDir string) string {
	integrationDir, _ = filepath.Abs(integrationDir)
	integration := strings.Replace(filepath.Base(integrationDir), "agent.next.", "", -1)
	fp := filepath.Join(integrationDir, "integration.go")
	if !fileutil.FileExists(fp) {
		log.Fatal(logger, "couldn't find the integration at "+fp)
	}
	os.MkdirAll(distDir, 0700)
	integrationFile := filepath.Join(distDir, runtime.GOOS, runtime.GOARCH, integration)
	log.Info(logger, "will generate data into temp folder at "+distDir)

	// build our integration
	c := exec.Command(os.Args[0], "build", "--dir", distDir, integrationDir)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		log.Fatal(logger, "error running build", "err", err)
	}
	return integrationFile
}

// DevCmd represents the dev command
var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "run an integration in development mode",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationDir := args[0]
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()

		started := time.Now()
		defer func() {
			log.Info(logger, "completed", "duration", time.Since(started).String())
		}()

		distDir := filepath.Join(os.TempDir(), "agent.next")

		integrationFile := buildIntegration(logger, distDir, integrationDir)

		channel, _ := cmd.Flags().GetString("channel")
		dir, _ := cmd.Flags().GetString("dir")
		if dir != "" {
			dir, _ = filepath.Abs(dir)
		}
		historical, _ := cmd.Flags().GetBool("historical")
		devargs := []string{"dev-export", "--dir", dir, "--channel", channel, "--log-level", "debug"}

		if historical {
			devargs = append(devargs, "--historical=true")
		}

		set, _ := cmd.Flags().GetStringArray("set")
		for _, str := range set {
			devargs = append(devargs, "--set", str)
		}

		c := exec.Command(integrationFile, devargs...)

		pos.OnExit(func(_ int) {
			if c != nil {
				syscall.Kill(-c.Process.Pid, syscall.SIGINT)
				c = nil
			}
		})

		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin
		c.Dir = distDir
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := c.Start(); err != nil {
			os.Exit(1)
		}
		c.Wait()
	},
}

// // webHookCmd represents the dev webhook command
// var webHookCmd = &cobra.Command{
// 	Use:   "webhook",
// 	Short: "run an integration in development mode and feed it a webhook",
// 	Args:  cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		integrationDir := args[0]
// 		logger := log.NewCommandLogger(cmd)
// 		defer logger.Close()

// 		started := time.Now()
// 		defer func() {
// 			log.Info(logger, "completed", "duration", time.Since(started).String())
// 		}()

// 		data, _ := cmd.Flags().GetString("data")
// 		if data == "" {
// 			log.Fatal(logger, "--data is required")
// 		}

// 		distDir := filepath.Join(os.TempDir(), "agent.next")
// 		integrationFile := buildIntegration(logger, distDir, integrationDir)

// 		channel, _ := cmd.Flags().GetString("channel")
// 		dir, _ := cmd.Flags().GetString("dir")
// 		if dir != "" {
// 			dir, _ = filepath.Abs(dir)
// 		}

// 		devargs := []string{
// 			"dev-webhook",
// 			"--dir", dir,
// 			"--channel", channel,
// 			"--data", data,
// 			"--log-level", "debug",
// 		}

// 		headers, _ := cmd.Flags().GetStringArray("headers")
// 		for _, str := range headers {
// 			devargs = append(devargs, "--headers", str)
// 		}

// 		set, _ := cmd.Flags().GetStringArray("set")
// 		for _, str := range set {
// 			devargs = append(devargs, "--set", str)
// 		}

// 		c := exec.Command(integrationFile, devargs...)

// 		pos.OnExit(func(_ int) {
// 			if c != nil {
// 				syscall.Kill(-c.Process.Pid, syscall.SIGINT)
// 				c = nil
// 			}
// 		})

// 		c.Stderr = os.Stderr
// 		c.Stdout = os.Stdout
// 		c.Stdin = os.Stdin
// 		c.Dir = distDir
// 		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
// 		if err := c.Start(); err != nil {
// 			os.Exit(1)
// 		}
// 		c.Wait()
// 	},
// }

func init() {
	// add command to root in ../dev.go
	DevCmd.PersistentFlags().StringArray("set", []string{}, "a config key/value pair such as a=b")
	DevCmd.PersistentFlags().String("dir", "dev_dist", "the directory to output pipe contents")
	DevCmd.PersistentFlags().String("channel", "dev", "the channel which can be set")
	DevCmd.Flags().MarkHidden("channel")
	DevCmd.Flags().Bool("historical", false, "force a historical export")
	// DevCmd.AddCommand(webHookCmd)
	// webHookCmd.Flags().StringArray("headers", []string{}, "headers key/value pair such as a=b")
	// webHookCmd.Flags().String("data", "", "json body of a webhook payload")
}
