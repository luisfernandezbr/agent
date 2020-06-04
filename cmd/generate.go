package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/pinpt/agent.next/generator"
	"github.com/pinpt/go-common/v10/log"
	"github.com/spf13/cobra"
)

// genCmd represents the dev command
var genCmd = &cobra.Command{
	Use:   "generate <project_name>",
	Short: "generates an integration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Print(color.New(color.FgHiBlue).Sprint(`
    ____  _                   _       __ 
   / __ \(_)___  ____  ____  (_)___  / /_
  / /_/ / / __ \/ __ \/ __ \/ / __ \/ __/
 / ____/ / / / / /_/ / /_/ / / / / / /_  
/_/   /_/_/ /_/ .___/\____/_/_/ /_/\__/  
             /_/                         
`))
		fmt.Println("Welcome to the Pinpoint Integration generator!")
		fmt.Println()

		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		var err error
		var result generator.Info

		if result.PKG, err = checkProjectPath(args[0]); err != nil {
			log.Error(logger, "error with project path", "err", err)
			os.Exit(1)
		}
		if err = promptSettings(&result); err != nil {
			log.Error(logger, "error with settings", "err", err)
			os.Exit(1)
		}

		if err = generator.Generate(args[0], result); err != nil {
			log.Error(logger, "error with generator", "err", err)
			os.Exit(1)
		}
		fmt.Println()
		fmt.Println("🎉 project created! open ./" + args[0] + " in VSCode and start coding!")
		fmt.Println()
	},
}

func checkProjectPath(projname string) (string, error) {
	gopath := os.Getenv("GOPATH") + "/src/"
	pwd, _ := os.Getwd()
	if !strings.HasPrefix(pwd, gopath) {
		return "", errors.New("the project must be in the GOPATH")
	}
	pkg := strings.TrimPrefix(pwd, gopath) + "/" + projname
	if len(strings.Split(pkg, "/")) != 3 {
		return "", errors.New("cd to your org folder path, ie: gopath/src/github.com/{{orgname}}")
	}
	return pkg, nil
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
	err := survey.Ask([]*survey.Question{
		{
			Name: "integration_name",
			Prompt: &survey.Input{
				Message: "Name of the integration",
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
				Message: "Your name, or your company's name",
			},
			Validate:  survey.Required,
			Transform: survey.Title,
		},
		{
			Name: "publisher_url",
			Prompt: &survey.Input{
				Message: "Your company's url",
			},
			Validate: func(val interface{}) error {
				return validateURL(val, true)
			},
			Transform: survey.ToLower,
		},
		{
			Name: "publisher_avatar",
			Prompt: &survey.Input{
				Message: "Your avatar url, or your company's avatar url",
			},
			Validate: func(val interface{}) error {
				return validateURL(val, false)
			},
			Transform: survey.ToLower,
		},
		{
			Name: "integraion_types",
			Prompt: &survey.MultiSelect{
				Message: "Choose capabilities:",
				Options: []string{
					generator.IntegraionTypeIssueTracking.String(),
					generator.IntegraionTypeSourcecode.String(),
					generator.IntegraionTypeCodeQuality.String(),
					generator.IntegraionTypeCalendar.String(),
				},
			},
		},
	}, result)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(genCmd)
}
