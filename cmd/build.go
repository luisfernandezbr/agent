package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func getBuildCommitForIntegration(integrationDir string) string {
	gensha := exec.Command("git", "rev-parse", "HEAD")
	var shabuf bytes.Buffer
	gensha.Stdout = &shabuf
	gensha.Dir = integrationDir
	gensha.Run()
	return strings.TrimSpace(shabuf.String())
}

func generateMainTemplate(filename, content, descriptor, build, sha string) (string, error) {
	i := strings.Index(content, "runner.Main(&Integration)")
	if i < 0 {
		return "", fmt.Errorf("couldn't find the correct runner.Main func in %s", filename)
	}
	before := content[0:i]
	after := content[i+25:]
	return fmt.Sprintf(`%s
	IntegrationDescriptor := "%s"
	IntegrationBuildDate := "%s"
	IntegrationBuildCommitSHA := "%s"
	runner.Main(&Integration, IntegrationDescriptor, IntegrationBuildDate, IntegrationBuildCommitSHA)
%s
`, before, descriptor, build, sha, after), nil
}

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
		dist := filepath.Join(distDir)
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
			version := getBuildCommitForIntegration(integrationDir)
			bbuf := base64.StdEncoding.EncodeToString(buf)
			tmpl, err := generateMainTemplate(ygofn, string(gobuf), bbuf, datetime.ISODate(), version)
			if err != nil {
				log.Fatal(logger, "error generating build", "err", err)
			}
			ioutil.WriteFile(ygofn, []byte(tmpl), 0644)
			bundleRewriter = func() {
				ioutil.WriteFile(ygofn, gobuf, 0644)
			}
			defer bundleRewriter()
		}
		theenv := os.Environ()
		oses, _ := cmd.Flags().GetStringSlice("os")
		for _, theos := range oses {
			arches, _ := cmd.Flags().GetStringSlice("arch")
			for _, arch := range arches {
				env := append(theenv, []string{"GOOS=" + theos, "GOARCH=" + arch}...)
				outfn := filepath.Join(dist, theos, arch, integration)
				os.MkdirAll(filepath.Dir(outfn), 0700)
				c := exec.Command("go", "build", "-o", outfn)
				c.Stderr = os.Stderr
				c.Stdout = os.Stdout
				c.Stdin = os.Stdin
				c.Dir = integrationDir
				c.Env = env
				if err := c.Run(); err != nil {
					bundleRewriter()
					ioutil.WriteFile(modfp, mod, 0644) // restore original
					os.Exit(1)
				}
				log.Debug(logger, "file built to "+outfn)
			}
		}
		ioutil.WriteFile(modfp, mod, 0644) // restore original
		if bundleRewriter != nil {
			bundleRewriter()
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().String("dir", "dist", "the output directory to place the generated file")
	buildCmd.Flags().Bool("bundle", true, "bundle artifacts into the library")
	buildCmd.Flags().StringSlice("os", []string{pos.Getenv("GOOS", runtime.GOOS)}, "the OS to build for")
	buildCmd.Flags().StringSlice("arch", []string{pos.Getenv("GOARCH", runtime.GOARCH)}, "the architecture to build for")
}
