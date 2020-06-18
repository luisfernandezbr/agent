package cmd

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
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

func openCert(pemFilename string) (*x509.Certificate, error) {
	buf, err := ioutil.ReadFile(pemFilename)
	if err != nil {
		return nil, fmt.Errorf("error openning file: %w", err)
	}
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, fmt.Errorf("no pem data in file %s", pemFilename)
	}
	return x509.ParseCertificate(block.Bytes)
}

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <destination> <integration> <version>",
	Short: "download an integration from the registry",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		tmpdir, err := ioutil.TempDir("", "")
		if err != nil {
			log.Fatal(logger, "error creating temp dir", "err", err)
		}
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
		signature := resp.Header.Get("x-pinpt-signature")
		if signature == "" {
			log.Fatal(logger, "no signature from server, cannot verify bundle")
		}
		sigBuf, err := hex.DecodeString(signature)
		if err != nil {
			log.Fatal(logger, "error decoding signature", "err", err)
		}
		src := filepath.Join(tmpdir, "bundle.zip")
		dest := filepath.Join(tmpdir, "bundle")
		of, err := os.Create(src)
		if err != nil {
			log.Fatal(logger, "error opening download file", "err", err)
		}
		defer of.Close()
		_, err = io.Copy(of, resp.Body)
		if err != nil {
			log.Fatal(logger, "error copying bundle data", "err", err)
		}
		of.Close()
		resp.Body.Close()
		// TODO(robin): figure out why we cant use hash.ChecksumCopy
		sum, err := fileutil.Checksum(src)
		if err != nil {
			log.Fatal(logger, "error taking checksum of downloaded bundle", "err", err)
		}
		checksum, err := hex.DecodeString(sum)
		if err != nil {
			log.Fatal(logger, "error decoding checksum", "err", err)
		}
		if err := fileutil.Unzip(src, dest); err != nil {
			log.Fatal(logger, "error performing unzip for integration", "err", err)
		}
		certfile := filepath.Join(dest, "cert.pem")
		if !fileutil.FileExists(certfile) {
			log.Fatal(logger, "error finding integration developer certificate for bundle", "file", certfile)
		}
		cert, err := openCert(certfile)
		if err != nil {
			log.Fatal(logger, "error opening certificate from bundle", "err", err)
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			log.Fatal(logger, "certificate public key was not rsa")
		}
		if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, checksum, sigBuf); err != nil {
			if err == rsa.ErrVerification {
				log.Fatal(logger, "invalid signature or certificate")
			}
			log.Fatal(logger, "error verifying bundle signature", "err", err)
		}
		datafn := filepath.Join(dest, "data.zip")
		if err := fileutil.Unzip(datafn, dest); err != nil {
			log.Fatal(logger, "error performing unzip for integration data", "err", err)
		}
		destfn := filepath.Join(dest, runtime.GOOS, runtime.GOARCH, integration)
		if !fileutil.FileExists(destfn) {
			log.Fatal(logger, "error finding integration binary for bundle", "file", destfn)
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
