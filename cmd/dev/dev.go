package dev

import (
	"io/ioutil"
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
	log.Debug(logger, "will generate integration into folder at "+distDir)

	// build our integration
	c := exec.Command(os.Args[0], "build", "--dir", distDir, integrationDir)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		log.Fatal(logger, "error running build", "err", err)
	}
	return integrationFile
}

type commandCallback func(cmd *cobra.Command, args []string) []string

func createDevCommand(name string, cmdname string, short string, inputRequired bool, callback commandCallback) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
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

			devargs := callback(cmd, []string{
				cmdname,
				"--dir", dir,
				"--channel", channel,
				"--log-level", "debug",
			})
			if inputRequired {
				data, _ := cmd.Flags().GetString("input")
				if data == "" {
					log.Fatal(logger, "--input is required")
				}
				if fileutil.FileExists(data) {
					log.Info(logger, "loading webhook from file", "file", data)
					buf, err := ioutil.ReadFile(data)
					if err != nil {
						log.Fatal(logger, "error reading data file", "err", err)
					}
					data = string(buf)
				}
				devargs = append(devargs, "--input", data)
			}

			headers, _ := cmd.Flags().GetStringArray("header")
			for _, str := range headers {
				devargs = append(devargs, "--header", str)
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
}

// DevCmd represents the dev command
var DevCmd = createDevCommand("dev", "dev-export", "run an integration in development mode", false, func(cmd *cobra.Command, devargs []string) []string {
	historical, _ := cmd.Flags().GetBool("historical")
	if historical {
		devargs = append(devargs, "--historical=true")
	}

	webhookEnabled, _ := cmd.Flags().GetBool("webhook")
	if webhookEnabled {
		devargs = append(devargs, "--webhook")
	}

	record, _ := cmd.Flags().GetString("record")
	replay, _ := cmd.Flags().GetString("replay")

	if record != "" {
		record, _ = filepath.Abs(record)
		devargs = append(devargs, "--record", record)
	}
	if replay != "" {
		replay, _ = filepath.Abs(replay)
		devargs = append(devargs, "--replay", replay)
	}
	return devargs
})

var webHookCmd = createDevCommand("webhook", "dev-webhook", "run an integration in development mode and feed it a webhook", true, func(cmd *cobra.Command, args []string) []string {
	refID, _ := cmd.Flags().GetString("ref-id")
	webhookURL, _ := cmd.Flags().GetString("webhook-url")
	return append(args, "--ref-id", refID, "--webhook-url", webhookURL)
})

var mutationCmd = createDevCommand("mutation", "dev-mutation", "run an integration in development mode and feed it a mutation", true, func(cmd *cobra.Command, args []string) []string {
	return args
})

func init() {
	// add command to root in ../dev.go
	DevCmd.PersistentFlags().StringArray("set", []string{}, "a config key/value pair such as a=b")
	DevCmd.PersistentFlags().String("dir", "dev_dist", "the directory to output pipe contents")
	DevCmd.PersistentFlags().String("channel", "dev", "the channel which can be set")
	DevCmd.Flags().Bool("webhook", false, "enable webhook registration")
	DevCmd.Flags().MarkHidden("channel")
	DevCmd.Flags().Bool("historical", false, "force a historical export")
	DevCmd.Flags().String("record", "", "record all interactions to directory specified")
	DevCmd.Flags().String("replay", "", "replay all interactions from directory specified")
	DevCmd.AddCommand(webHookCmd)
	DevCmd.AddCommand(mutationCmd)
	webHookCmd.Flags().StringArray("header", []string{}, "headers key/value pair such as a=b")
	webHookCmd.Flags().String("input", "", "json body of a webhook payload, as a string or file")
	webHookCmd.Flags().String("ref-id", "9999", "the ref_id value")
	webHookCmd.Flags().String("webhook-url", "http://example.com/hook/123456", "the webhook url value")

	mutationCmd.Flags().String("input", "", "json body of a mutation payload, as a string or file")
}
