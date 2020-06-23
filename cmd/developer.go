// +build dev

package cmd

import "github.com/pinpt/agent.next/cmd/dev"

func init() {
	rootCmd.AddCommand(dev.BuildCmd)
	rootCmd.AddCommand(dev.DevCmd)
	rootCmd.AddCommand(dev.EnrollCmd)
	rootCmd.AddCommand(dev.GenCmd)
	rootCmd.AddCommand(dev.LoginCmd)
	rootCmd.AddCommand(dev.LogoutCmd)
	rootCmd.AddCommand(dev.PackageCmd)
	rootCmd.AddCommand(dev.PublishCmd)
}
