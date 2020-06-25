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
	pos "github.com/pinpt/go-common/v10/os"
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

func downloadIntegration(logger log.Logger, channel string, toDir string, publisher string, integration string, version string) (string, error) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	cl, err := api.NewHTTPAPIClientDefault()
	if err != nil {
		return "", fmt.Errorf("error creating client: %w", err)
	}
	url := pstr.JoinURL(api.BackendURL(api.RegistryService, channel), fmt.Sprintf("/fetch/%s/%s/%s", publisher, integration, version))
	log.Debug(logger, "downloading", "url", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	resp, err := cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error downloading request status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	signature := resp.Header.Get("x-pinpt-signature")
	if signature == "" {
		return "", fmt.Errorf("no signature from server, cannot verify bundle")
	}
	sigBuf, err := hex.DecodeString(signature)
	if err != nil {
		return "", fmt.Errorf("error decoding signature: %w", err)
	}
	src := filepath.Join(tmpdir, "bundle.zip")
	dest := filepath.Join(tmpdir, "bundle")
	of, err := os.Create(src)
	if err != nil {
		return "", fmt.Errorf("error opening download file: %w", err)
	}
	defer of.Close()

	if _, err := io.Copy(of, resp.Body); err != nil {
		return "", fmt.Errorf("error copying bundle data: %w", err)
	}
	if err := of.Close(); err != nil {
		return "", fmt.Errorf("error closing bundle.zip: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return "", fmt.Errorf("error closing response body: %w", err)
	}
	// TODO(robin): figure out why we cant use hash.ChecksumCopy
	sum, err := fileutil.Checksum(src)
	if err != nil {
		return "", fmt.Errorf("error taking checksum of downloaded bundle: %w", err)
	}
	checksum, err := hex.DecodeString(sum)
	if err != nil {
		return "", fmt.Errorf("error decoding checksum: %w", err)
	}
	if err := fileutil.Unzip(src, dest); err != nil {
		return "", fmt.Errorf("error performing unzip for integration: %w", err)
	}
	certfile := filepath.Join(dest, "cert.pem")
	if !fileutil.FileExists(certfile) {
		return "", fmt.Errorf("error finding integration developer certificate (%s) for bundle", certfile)
	}
	cert, err := openCert(certfile)
	if err != nil {
		return "", fmt.Errorf("error opening certificate from bundle: %w", err)
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("certificate public key was not rsa")
	}
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, checksum, sigBuf); err != nil {
		if err == rsa.ErrVerification {
			return "", fmt.Errorf("invalid signature or certificate")
		}
		return "", fmt.Errorf("error verifying bundle signature: %w", err)
	}
	datafn := filepath.Join(dest, "data.zip")
	if err := fileutil.Unzip(datafn, dest); err != nil {
		return "", fmt.Errorf("error performing unzip for integration data: %w", err)
	}
	destfn := filepath.Join(dest, runtime.GOOS, runtime.GOARCH, integration)
	if !fileutil.FileExists(destfn) {
		return "", fmt.Errorf("error finding integration binary (%s) in bundle", destfn)
	}
	sf, err := os.Open(destfn)
	if err != nil {
		return "", fmt.Errorf("error opening file (%s): %w", destfn, err)
	}
	defer sf.Close()
	outfn := filepath.Join(toDir, integration)
	os.Remove(outfn)
	df, err := os.Create(outfn)
	if err != nil {
		return "", fmt.Errorf("error creating file (%s): %w", outfn, err)
	}
	defer df.Close()
	if _, err := io.Copy(df, sf); err != nil {
		return "", fmt.Errorf("error copying binary data: %w", err)
	}
	if err := df.Close(); err != nil {
		return "", fmt.Errorf("error closing output file (%s): %w", outfn, err)
	}
	os.Chmod(outfn, 0755) // make it executable
	outfn, _ = filepath.Abs(outfn)
	return outfn, nil
}

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <destination> <integration> <version>",
	Short: "download an integration from the registry",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		destDir := args[0]
		fullIntegration := args[1]
		version := args[2]
		os.MkdirAll(destDir, 0700)
		channel, _ := cmd.Flags().GetString("channel")
		tok := strings.Split(fullIntegration, "/")
		if len(tok) != 2 {
			log.Fatal(logger, "integration should be in the format: publisher/integration such as pinpt/github")
		}
		publisher := tok[0]
		integration := tok[1]
		outfn, err := downloadIntegration(logger, channel, destDir, publisher, integration, version)
		if err != nil {
			log.Fatal(logger, "error downloading integration", "err", err)
		}
		log.Info(logger, "platform integration available at "+outfn)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
}
