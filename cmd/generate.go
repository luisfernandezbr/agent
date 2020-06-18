package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/pinpt/agent.next/generator"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	"github.com/spf13/cobra"
)

func banner() {

	fmt.Print(color.New(color.FgHiBlue).Sprint(`
    ____  _                   _       __ 
   / __ \(_)___  ____  ____  (_)___  / /_
  / /_/ / / __ \/ __ \/ __ \/ / __ \/ __/
 / ____/ / / / / /_/ / /_/ / / / / / /_  
/_/   /_/_/ /_/ .___/\____/_/_/ /_/\__/  
             /_/                         
`))
}

// genCmd represents the dev command
var genCmd = &cobra.Command{
	Use:   "generate",
	Short: "generates an integration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		banner()
		fmt.Println("Welcome to the Pinpoint Integration generator!")
		fmt.Println()

		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		var err error
		var result generator.Info

		if err = promptSettings(&result); err != nil {
			log.Error(logger, "error with settings", "err", err)
			os.Exit(1)
		}

		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			var buf strings.Builder
			c := exec.Command("go", "env", "GOPATH")
			c.Stdout = &buf
			if err := c.Run(); err != nil {
				log.Fatal(logger, "error finding your GOPATH. is Golang installed and on your PATH?", "err", err)
			}
			gopath = strings.TrimSpace(buf.String())
		}

		result.Dir = filepath.Join(gopath, "src", result.Pkg)
		if !fileutil.FileExists(result.Dir) {
			os.MkdirAll(result.Dir, 0700)
		}

		if err = generator.Generate(result.Dir, result); err != nil {
			log.Fatal(logger, "error with generator", "err", err)
		}

		appDir := filepath.Join(result.Dir, "app")

		// install the latest websdk and uic.next
		sdkInstall := exec.Command("npm", "install", "@pinpt/agent.websdk", "@pinpt/uic.next", "--save", "--loglevel", "error")
		sdkInstall.Dir = appDir
		sdkInstall.Stderr = os.Stderr
		sdkInstall.Stdin = os.Stdin
		sdkInstall.Stdin = os.Stdin
		sdkInstall.Run()

		// run the npm install
		npmInstall := exec.Command("npm", "install", "--loglevel", "error")
		npmInstall.Dir = appDir
		npmInstall.Stderr = os.Stderr
		npmInstall.Stdin = os.Stdin
		npmInstall.Stdin = os.Stdin
		npmInstall.Run()

		fmt.Println()
		fmt.Println("🎉 project created! open " + result.Dir + " in your editor and start coding!")
		fmt.Println()
	},
}

func promptSettings(result *generator.Info) error {
	validateURL := func(val interface{}, required bool) error {
		str := val.(string)
		if str == "" && !required {
			return nil
		}
		u, err := url.ParseRequestURI(str)
		if err != nil {
			return err
		}
		if u.Scheme == "" {
			return errors.New("missing scheme")
		}
		if u.Host == "" {
			return errors.New("missing host")
		}
		return nil
	}
	return survey.Ask([]*survey.Question{
		{
			Name: "pkg",
			Prompt: &survey.Input{
				Message: "Go Package Name:",
				Help:    "Your Go package such as github.com/pinpt/myintegration",
			},
			Validate: survey.Required,
		},
		{
			Name: "integration_name",
			Prompt: &survey.Input{
				Message: "Name of the integration:",
			},
			Transform: survey.Title,
			Validate: func(val interface{}) error {
				str := strings.TrimSpace(val.(string))
				if str == "" {
					return errors.New("integration name cannot be empty")
				}
				if strings.Contains(str, " ") {
					return errors.New("integration name cannot not contain spaces")
				}
				return nil
			},
		},
		{
			Name: "publisher_name",
			Prompt: &survey.Input{
				Message: "Your company's name:",
			},
			Validate:  survey.Required,
			Transform: survey.Title,
		},
		{
			Name: "publisher_url",
			Prompt: &survey.Input{
				Message: "Your company's url:",
			},
			Validate: func(val interface{}) error {
				return validateURL(val, true)
			},
			Transform: survey.ToLower,
		},
		{
			Name: "identifier",
			Prompt: &survey.Input{
				Message: "Your company's short, unique identifier:",
				Help:    "For example, for Pinpoint we use pinpt. Make sure you choose a unique value",
			},
			Validate:  survey.Required,
			Transform: survey.ToLower,
		},
		{
			Name: "publisher_avatar",
			Prompt: &survey.Input{
				Message: "Your company's avatar url:",
			},
			Validate: func(val interface{}) error {
				return validateURL(val, false)
			},
			Transform: survey.ToLower,
		},
		{
			Name: "integration_types",
			Prompt: &survey.MultiSelect{
				Message: "Choose integration capabilities:",
				Options: []string{
					generator.IntegrationTypeIssueTracking.String(),
					generator.IntegrationTypeSourcecode.String(),
					generator.IntegrationTypeCodeQuality.String(),
					generator.IntegrationTypeCalendar.String(),
				},
			},
		},
	}, result)
}

func init() {
	rootCmd.AddCommand(genCmd)
}
