package cmd

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pinpt/go-common/fileutil"
	"github.com/pinpt/go-common/log"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build an integration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationDir := args[0]
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		integrationDir, _ = filepath.Abs(integrationDir)
		integration := strings.Replace(filepath.Base(integrationDir), "agent.next.", "", -1)
		fp := filepath.Join(integrationDir, "integration.go")
		if !fileutil.FileExists(fp) {
			log.Fatal(logger, "couldn't find the integration at "+fp)
		}
		distDir, _ := cmd.Flags().GetString("dir")
		distDir, _ = filepath.Abs(distDir)
		os.MkdirAll(distDir, 0700)
		dist := filepath.Join(distDir, integration+".so")
		// local dev issue with plugins: https://github.com/golang/go/issues/31354
		modfp := filepath.Join(integrationDir, "go.mod")
		mod, err := ioutil.ReadFile(modfp)
		if err != nil {
			log.Fatal(logger, "error reading plugin go.mod", "err", err)
		}
		ioutil.WriteFile(modfp, []byte(string(mod)+"\nreplace github.com/pinpt/agent.next => ../agent.next"), 0644)
		c := exec.Command("go", "build", "-buildmode=plugin", "-o", dist, fp)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Dir = integrationDir
		if err := c.Run(); err != nil {
			ioutil.WriteFile(modfp, mod, 0644) // restore original
			os.Exit(1)
		}
		ioutil.WriteFile(modfp, mod, 0644) // restore original
		log.Info(logger, "file built to "+dist)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().String("dir", "dist", "the output directory to place the generated file")
}
