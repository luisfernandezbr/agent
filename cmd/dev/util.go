package dev

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pinpt/go-common/v10/log"
	"github.com/spf13/cobra"
)

// UtilCmd is for dev utilities
var UtilCmd = &cobra.Command{
	Use:   "util",
	Short: "base for all util commands",
}

func unpackPipeData(data string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewBuffer(raw))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error creating gzip reader: %w", err)
	}
	buf, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("error reading gzip data: %w", err)
	}
	if err := zr.Close(); err != nil {
		return nil, fmt.Errorf("error closing gzip reader: %w", err)
	}
	return buf, nil
}

var decodePipeBodyCmd = &cobra.Command{
	Use:   "decodePipeBody <file>",
	Short: "decode the Objects of an agent.Data model",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		search, _ := cmd.Flags().GetString("search")
		for _, fn := range args {
			f, err := ioutil.ReadFile(fn)
			if err != nil {
				log.Fatal(logger, "error reading file", "file", fn, "err", err)
			}
			var objects map[string]string
			if err := json.Unmarshal(f, &objects); err != nil {
				log.Fatal(logger, "error decoding body", "err", err)
			}
			var found, searched int
			for modelName, data := range objects {
				log.Info(logger, modelName)
				buf, err := unpackPipeData(data)
				if err != nil {
					log.Fatal(logger, "error decoding pipe data", "err", err)
				}
				for _, line := range bytes.Split(buf, []byte("\n")) {
					searched++
					if search != "" {
						if bytes.Contains(line, []byte(search)) {
							found++
							fmt.Println(string(line))
						}
					} else {
						fmt.Println(string(line))
					}
				}
			}
			if search != "" {
				log.Info(logger, fmt.Sprintf("found %d matches to search query '%s'", found, search), "total", searched)
			}
		}
	},
}

func init() {
	UtilCmd.AddCommand(decodePipeBodyCmd)
	decodePipeBodyCmd.Flags().String("search", "", "search all data for given string")
}
