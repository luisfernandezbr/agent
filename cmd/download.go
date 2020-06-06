package cmd

import (
	"archive/zip"
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

func unzip(logger log.Logger, src, dest string) error {
	log.Debug(logger, "unzipping", "src", src, "dest", dest)
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		// Store filename/path for returning and using later on
		/* #nosec */
		fpath := filepath.Join(dest, f.Name)
		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}
		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		n, err := io.Copy(outFile, rc)
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
		log.Debug(logger, "unzipped", "file", outFile.Name(), "size", n)
	}
	return nil
}

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
		integration := args[1]
		version := args[2]
		publisher := "pinpt"
		os.MkdirAll(destDir, 0700)
		channel, _ := cmd.Flags().GetString("channel")
		cl, err := api.NewHTTPAPIClientDefault()
		if err != nil {
			log.Fatal(logger, "error creating client", "err", err)
		}
		url := pstr.JoinURL(api.BackendURL(api.RegistryService, channel), fmt.Sprintf("/fetch/%s/%s/%s", publisher, integration, version))
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
		if err := unzip(logger, src, dest); err != nil {
			log.Fatal(logger, "error performing unzip for integration", "err", err)
		}
		shasum := filepath.Join(dest, "sha512sum.txt.asc")
		if !fileutil.FileExists(shasum) {
			log.Fatal(logger, "error finding integration checksums for bundle", "file", shasum)
		}
		datafn := filepath.Join(dest, "data.zip")
		if err := unzip(logger, datafn, dest); err != nil {
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
				cs, err := checksum(f)
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
		log.Debug(logger, "platform integration available at "+outfn)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().String("channel", "stable", "the channel which can be set")
}
