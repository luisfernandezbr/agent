package dev

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cirruslabs/echelon"
	"github.com/cirruslabs/echelon/renderers"
	"github.com/pinpt/agent/v4/internal/util"
	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/hash"
	"github.com/pinpt/go-common/v10/log"
	pnum "github.com/pinpt/go-common/v10/number"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

func signFile(filename string, privateKey *rsa.PrivateKey) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("error openning bundle: %w", err)
	}
	defer f.Close()
	sum, err := hash.Sha256Checksum(f)
	if err != nil {
		return "", fmt.Errorf("error creating checksum: %w", err)
	}
	sigBuf, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, sum)
	if err != nil {
		return "", fmt.Errorf("error signing checksum of bundle: %w", err)
	}
	return hex.EncodeToString(sigBuf), nil
}

// PublishCmd represents the publish command
var PublishCmd = &cobra.Command{
	Use:   "publish <integration dir>",
	Short: "publish an integration to the registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationDir := args[0]
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		channel, _ := cmd.Flags().GetString("channel")
		chunkSize, _ := cmd.Flags().GetInt64("chunk")

		tmpdir, err := ioutil.TempDir("", "")
		if err != nil {
			log.Fatal(logger, "error creating temp dir", "err", err)
		}
		defer os.RemoveAll(tmpdir)

		c, err := loadDevConfig(channel)
		if err != nil {
			log.Fatal(logger, "error opening developer config", "err", err)
		}
		if c.PrivateKey == "" {
			log.Fatal(logger, "missing private key in config, please enroll before publishing")
		}
		if c.expired() {
			log.Fatal(logger, "your login session has expired. please login again")
		}
		privateKey, err := util.ParsePrivateKey(c.PrivateKey)
		if err != nil {
			log.Fatal(logger, "unable to parse private key in config")
		}

		log.Info(logger, "building package")
		cm := exec.Command(os.Args[0], "package", integrationDir, "--dir", tmpdir, "--channel", channel)
		cm.Stdout = os.Stdout
		cm.Stderr = os.Stderr
		cm.Stdin = os.Stdin
		if err := cm.Run(); err != nil {
			log.Fatal(logger, "error running package command", "err", err)
		}
		bundle := filepath.Join(tmpdir, "bundle.zip")
		if !fileutil.FileExists(bundle) {
			log.Fatal(logger, "error bundle does not exist", "err", err)
		}
		signature, err := signFile(bundle, privateKey)
		if err != nil {
			log.Fatal(logger, "error getting signature for bundle", "err", err)
		}

		rd, nopts, fiSize, errs, err := uploadPogress(logger, bundle, chunkSize)
		if err != nil {
			log.Fatal(logger, "error on upload", "err", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		opts := []api.WithOption{
			api.WithHeader("x-pinpt-signature", signature),
		}

		opts = append(opts, nopts...)

		apikey, _ := cmd.Flags().GetString("apikey")
		secret, _ := cmd.Flags().GetString("secret")
		if secret != "" {
			opts = append(opts, api.WithHeader("x-api-key", secret))
		} else if apikey == "" {
			apikey = c.APIKey
			if apikey == "" {
				log.Fatal(logger, "you must login or provide the apikey using --apikey before continuing")
			}
		}
		descriptorFn := filepath.Join(integrationDir, "integration.yaml")
		descriptorBuf, err := ioutil.ReadFile(descriptorFn)
		if err != nil {
			log.Fatal(logger, "error reading descriptor", "err", err, "file", descriptorFn)
		}
		descriptor, err := sdk.LoadDescriptor(base64.StdEncoding.EncodeToString(descriptorBuf), "", "")
		if err != nil {
			log.Fatal(logger, "error loading descriptor", "err", err, "file", descriptorFn)
		}
		version := getBuildCommitForIntegration(integrationDir)
		basepath := fmt.Sprintf("publish/%s/%s/%s", c.PublisherRefType, descriptor.RefType, version)
		log.Info(logger, "uploading", "size", pnum.ToBytesSize(fiSize))
		resp, err := api.Put(ctx, channel, api.RegistryService, basepath, apikey, rd, opts...)
		if err := <-errs; err != nil {
			log.Warn(logger, "error on upload", "err", err)
		}
		if err != nil || resp.StatusCode != http.StatusAccepted {
			var buf []byte
			if resp != nil {
				buf, _ = ioutil.ReadAll(resp.Body)
			}
			log.Fatal(logger, "error publishing your bundle", "err", err, "body", string(buf))
		}
		log.Info(logger, "🚀 published", "integration", descriptor.RefType, "version", version)
	},
}

func uploadPogress(logger sdk.Logger, bundle string, chunkSize int64) (io.Reader, []api.WithOption, int64, chan error, error) {

	errs := make(chan error, 1)

	file, err := os.Open(bundle)
	if err != nil {
		return nil, nil, 0, errs, fmt.Errorf("error opening bundle %s", err)
	}
	fi, _ := file.Stat()

	//buffer for storing multipart data
	byteBuf := &bytes.Buffer{}

	//part: parameters
	mpWriter := multipart.NewWriter(byteBuf)

	_, err = mpWriter.CreateFormFile("file", fi.Name())
	if err != nil {
		return nil, nil, 0, errs, fmt.Errorf("error creating form file %s", err)
	}

	contentType := mpWriter.FormDataContentType()

	nmulti := byteBuf.Len()
	multi := make([]byte, nmulti)
	_, err = byteBuf.Read(multi)
	if err != nil {
		return nil, nil, 0, errs, fmt.Errorf("error reading from buffer %s", err)
	}

	//part: latest boundary
	//when multipart closed, latest boundary is added
	err = mpWriter.Close()
	if err != nil {
		return nil, nil, 0, errs, fmt.Errorf("error closing writter %s", err)
	}
	nboundary := byteBuf.Len()
	lastBoundary := make([]byte, nboundary)
	_, err = byteBuf.Read(lastBoundary)
	if err != nil {
		return nil, nil, 0, errs, fmt.Errorf("error reading last boundary %s", err)
	}

	//calculate content length
	totalSize := int64(nmulti) + fi.Size() + int64(nboundary)

	rd, wr := io.Pipe()

	renderer := renderers.NewInteractiveRenderer(os.Stdout, nil)
	go renderer.StartDrawing()
	progressLog := echelon.NewLogger(echelon.InfoLevel, renderer)

	go func() {

		//write multipart
		_, err = wr.Write(multi)
		if err != nil {
			errs <- fmt.Errorf("error writting multipart %s", err)
			return
		}

		scoped := progressLog.Scoped("Uploading bundle")

		//write file
		buf := make([]byte, chunkSize)
		inc := 0
		lastProgress := 0
		for {
			n, err := file.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				scoped.Finish(false)
				progressLog.Finish(false)
				errs <- fmt.Errorf("error reading buffer %s", err)
				return
			}
			inc += len(buf[:n])
			progress := int(math.Round(float64(inc) / float64(totalSize) * 100))
			if progress != lastProgress {
				scoped.Infof("Progress %s/%s %d%%", pnum.ToBytesSize(int64(inc)), pnum.ToBytesSize(totalSize), progress)
			}
			lastProgress = progress

			_, err = wr.Write(buf[:n])
			if err != nil {
				scoped.Finish(false)
				progressLog.Finish(false)
				errs <- fmt.Errorf("error reading buffer chunk %s", err)
				return
			}
		}
		//write boundary
		progress := int(math.Round(float64(inc) / float64(totalSize) * 100))
		scoped.Infof("Progress %s/%s %d%%", pnum.ToBytesSize(int64(inc)), pnum.ToBytesSize(totalSize), progress)
		_, err = wr.Write(lastBoundary)
		if err != nil {
			scoped.Finish(false)
			progressLog.Finish(false)
			errs <- fmt.Errorf("error writting last boundary %s", err)
			return
		}
		if err := wr.Close(); err != nil {
			scoped.Finish(false)
			progressLog.Finish(false)
			errs <- fmt.Errorf("error closing writter %s", err)
			return
		}

		scoped.Finish(true)
		progressLog.Finish(true)
		renderer.StopDrawing()

		errs <- nil
	}()

	opts := []api.WithOption{
		api.WithContentType(contentType),
		api.WithHeader("x-file-length", strconv.FormatInt(fi.Size(), 10)),
		func(req *http.Request) error {
			req.ContentLength = totalSize
			return nil
		},
	}

	return rd, opts, fi.Size(), errs, nil
}

func init() {
	PublishCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "stable"), "the channel which can be set")
	PublishCmd.Flags().String("apikey", "", "api key")
	PublishCmd.Flags().String("secret", "", "internal shared secret")
	PublishCmd.Flags().Int64("chunk", 800, "chunk size")
	PublishCmd.Flags().MarkHidden("secret")
}
