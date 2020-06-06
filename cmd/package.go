package cmd

import (
	"archive/zip"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	"github.com/spf13/cobra"
)

func zipDir(filename string, dir string, pattern *regexp.Regexp) (int, error) {
	filenames, err := fileutil.FindFiles(dir, pattern)
	if err != nil {
		return 0, err
	}
	newZipFile, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range filenames {
		stat, _ := os.Stat(file)
		if !stat.IsDir() {
			if err = addFileToZip(zipWriter, dir, file); err != nil {
				return 0, err
			}
		}
	}
	return len(filenames), nil
}

func addFileToZip(zipWriter *zip.Writer, dir string, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name, _ = filepath.Rel(dir, filename)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func checksum(fn string) (string, error) {
	hasher := sha512.New()
	of, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	for {
		buf := make([]byte, 8096)
		n, err := of.Read(buf)
		if err == io.EOF || n == 0 {
			break
		}
		hasher.Write(buf[0:n])
	}
	of.Close()
	sha := hex.EncodeToString(hasher.Sum(nil))
	return sha, nil
}

func shaFiles(dir string, outfile string, re *regexp.Regexp) error {
	filenames, err := fileutil.FindFiles(dir, re)
	if err != nil {
		return err
	}
	var shas strings.Builder
	for _, fn := range filenames {
		stat, _ := os.Stat(fn)
		if !stat.IsDir() {
			sha, err := checksum(fn)
			if err != nil {
				return err
			}
			relfn, _ := filepath.Rel(dir, fn)
			shas.WriteString(sha + "  " + relfn + "\n")
		}
	}
	return ioutil.WriteFile(outfile, []byte(shas.String()), 0644)
}

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "package an integration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationDir := args[0]
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		integrationDir, _ = filepath.Abs(integrationDir)
		distDir, _ := cmd.Flags().GetString("dir")
		distDir, _ = filepath.Abs(distDir)
		bundleDir := filepath.Join(distDir, "bundle")
		dataDir := filepath.Join(bundleDir, "data")
		os.MkdirAll(bundleDir, 0700)
		os.MkdirAll(dataDir, 0700)

		buf, err := ioutil.ReadFile(filepath.Join(integrationDir, "integration.yaml"))
		if err != nil {
			log.Fatal(logger, "error loading integration.yaml", "err", err)
		}

		ioutil.WriteFile(filepath.Join(dataDir, "integration.yaml"), buf, 0644)

		dataFn := filepath.Join(bundleDir, "data.zip")
		bundleFn := filepath.Join(distDir, "bundle.zip")

		oss, _ := cmd.Flags().GetStringSlice("os")
		arches, _ := cmd.Flags().GetStringSlice("arch")

		cargs := []string{"build", integrationDir, "--dir", dataDir}
		for _, o := range oss {
			cargs = append(cargs, "--os", o)
		}
		for _, a := range arches {
			cargs = append(cargs, "--arch", a)
		}
		c := exec.Command(os.Args[0], cargs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Run(); err != nil {
			os.Exit(1)
		}

		sha := getBuildCommitForIntegration(integrationDir)

		// write out our version file
		ioutil.WriteFile(filepath.Join(bundleDir, "version.txt"), []byte(sha), 0644)

		// write out the sha sum512 for each file in the zip for integrity checking
		shafilename := filepath.Join(bundleDir, "sha512sum.txt.asc")
		if err := shaFiles(dataDir, shafilename, regexp.MustCompile(".*")); err != nil {
			log.Fatal(logger, "error generating sha sums", "err", err)
		}

		if _, err := zipDir(dataFn, dataDir, regexp.MustCompile(".*")); err != nil {
			log.Fatal(logger, "error building zip file", "err", err)
		}
		if _, err := zipDir(bundleFn, bundleDir, regexp.MustCompile(".(zip|asc|txt)$")); err != nil {
			log.Fatal(logger, "error building zip file", "err", err)
		}
		os.RemoveAll(bundleDir)
		log.Info(logger, "bundle packaged to "+bundleFn)
	},
}

func init() {
	rootCmd.AddCommand(packageCmd)
	packageCmd.Flags().String("dir", "dist", "the output directory to place the generated file")
	packageCmd.Flags().StringSlice("os", []string{"darwin", "linux"}, "the OS to build for")
	packageCmd.Flags().StringSlice("arch", []string{"amd64"}, "the architecture to build for")
}
