// +build dev

package cmd

import "github.com/pinpt/agent/cmd/dev"

func init() {
	rootCmd.AddCommand(
		dev.BuildCmd,
		dev.DevCmd,
		dev.EnrollCmd,
		dev.GenCmd,
		dev.LoginCmd,
		dev.LogoutCmd,
		dev.PackageCmd,
		dev.PublishCmd,
		dev.UtilCmd,
	)
}
