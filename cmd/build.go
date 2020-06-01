package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/datetime"
	"github.com/pinpt/go-common/fileutil"
	"github.com/pinpt/go-common/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type rewriteFunc func()

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

		bundle, _ := cmd.Flags().GetBool("bundle")
		var bundleRewriter rewriteFunc
		if bundle {
			yfn := filepath.Join(integrationDir, "integration.yaml")
			if !fileutil.FileExists(yfn) {
				log.Fatal(logger, "missing required file at "+yfn)
			}
			ygofn := filepath.Join(integrationDir, "integration.go")
			if !fileutil.FileExists(ygofn) {
				log.Fatal(logger, "missing required file at "+ygofn)
			}
			buf, err := ioutil.ReadFile(yfn)
			if err != nil {
				log.Fatal(logger, "error opening required file at "+yfn, "err", err)
			}
			gobuf, err := ioutil.ReadFile(ygofn)
			if err != nil {
				log.Fatal(logger, "error opening required file at "+ygofn, "err", err)
			}
			var descriptor sdk.Descriptor
			if err := yaml.Unmarshal(buf, &descriptor); err != nil {
				log.Fatal(logger, "error parsing config file at "+yfn, "err", err)
			}
			gensha := exec.Command("git", "rev-parse", "HEAD")
			var shabuf bytes.Buffer
			gensha.Stdout = &shabuf
			gensha.Dir = integrationDir
			gensha.Run()
			bbuf := base64.StdEncoding.EncodeToString(buf)
			ioutil.WriteFile(ygofn, []byte(fmt.Sprintf("%s\n\nvar IntegrationDescriptor = \"%s\"\nvar IntegrationBuildDate = \"%s\"\nvar IntegrationBuildCommitSHA = \"%s\"", gobuf, bbuf, datetime.ISODate(), strings.TrimSpace(shabuf.String()))), 0644)
			bundleRewriter = func() {
				ioutil.WriteFile(ygofn, gobuf, 0644)
			}
			defer bundleRewriter()
		}
		c := exec.Command("go", "build", "-buildmode=plugin", "-o", dist, fp)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Dir = integrationDir
		if err := c.Run(); err != nil {
			ioutil.WriteFile(modfp, mod, 0644) // restore original
			os.Exit(1)
		}
		ioutil.WriteFile(modfp, mod, 0644) // restore original
		if bundleRewriter != nil {
			bundleRewriter()
		}
		log.Info(logger, "file built to "+dist)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().String("dir", "dist", "the output directory to place the generated file")
	buildCmd.Flags().Bool("bundle", true, "bundle artifacts into the library")
}
