package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pstr "github.com/pinpt/go-common/v10/strings"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <destination> <integration> <version>",
	Short: "download an integration from the registry",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		tmpdir := os.TempDir()
		defer os.RemoveAll(tmpdir)
		// FIXME once we get list from the registry
		destDir := args[0]
		fullIntegration := args[1]
		version := args[2]
		os.MkdirAll(destDir, 0700)
		channel, _ := cmd.Flags().GetString("channel")
		cl, err := api.NewHTTPAPIClientDefault()
		if err != nil {
			log.Fatal(logger, "error creating client", "err", err)
		}
		tok := strings.Split(fullIntegration, "/")
		if len(tok) != 2 {
			log.Fatal(logger, "integration should be in the format: publisher/integration such as pinpt/github")
		}
		integration := tok[1]
		url := pstr.JoinURL(api.BackendURL(api.RegistryService, channel), fmt.Sprintf("/fetch/%s/%s", fullIntegration, version))
		log.Debug(logger, "downloading", "url", url)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Fatal(logger, "error creating request", "err", err)
		}
		resp, err := cl.Do(req)
		if err != nil {
			log.Fatal(logger, "error executing request", "err", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatal(logger, "error downloading request", "err", err, "code", resp.StatusCode)
		}
		defer resp.Body.Close()
		src := filepath.Join(tmpdir, "bundle.zip")
		dest := filepath.Join(tmpdir, "bundle")
		of, err := os.Create(src)
		if err != nil {
			log.Fatal(logger, "error opening download file", "err", err)
		}
		defer of.Close()
		io.Copy(of, resp.Body)
		of.Close()
		resp.Body.Close()
		if err := fileutil.Unzip(src, dest); err != nil {
			log.Fatal(logger, "error performing unzip for integration", "err", err)
		}
		shasum := filepath.Join(dest, "sha512sum.txt.asc")
		if !fileutil.FileExists(shasum) {
			log.Fatal(logger, "error finding integration checksums for bundle", "file", shasum)
		}
		datafn := filepath.Join(dest, "data.zip")
		if err := fileutil.Unzip(datafn, dest); err != nil {
			log.Fatal(logger, "error performing unzip for integration data", "err", err)
		}
		destfn := filepath.Join(dest, runtime.GOOS, runtime.GOARCH, integration)
		if !fileutil.FileExists(destfn) {
			log.Fatal(logger, "error finding integration binary for bundle", "file", destfn)
		}
		shas, err := ioutil.ReadFile(shasum)
		if err != nil {
			log.Fatal(logger, "error reading integration checksums for bundle", "file", shasum, "err", err)
		}
		if len(shas) == 0 {
			log.Fatal(logger, "error reading integration checksums for bundle", "file", shasum, "err", "file was empty")
		}
		for _, line := range strings.Split(string(shas), "\n") {
			tok := strings.Split(strings.TrimSpace(line), "  ")
			if len(tok) == 2 {
				expected := tok[0]
				basepath := tok[1]
				f := filepath.Join(dest, basepath)
				if !fileutil.FileExists(f) {
					log.Fatal(logger, "cannot find file in the bundle checksum but not in the bundle", "file", f)
				}
				cs, err := fileutil.Checksum(f)
				if err != nil {
					log.Fatal(logger, "error performing checksum on file", "file", f, "err", err)
				}
				if cs != expected {
					log.Fatal(logger, "checksum doesn't match from the bundle", "file", f, "expected", expected, "was", cs)
				}
				log.Debug(logger, "validating checksum", "path", basepath, "sha", expected)
			}
		}
		sf, err := os.Open(destfn)
		if err != nil {
			log.Fatal(logger, "error opening file", "file", destfn, "err", err)
		}
		defer sf.Close()
		outfn := filepath.Join(destDir, integration)
		os.Remove(outfn)
		df, err := os.Create(outfn)
		if err != nil {
			log.Fatal(logger, "error opening file", "file", outfn, "err", err)
		}
		defer df.Close()
		io.Copy(df, sf)
		df.Close()
		os.Chmod(outfn, 0500) // make it executable
		outfn, _ = filepath.Abs(outfn)
		log.Info(logger, "platform integration available at "+outfn)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().String("channel", "stable", "the channel which can be set")
}
