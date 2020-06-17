package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/log"
	"github.com/spf13/cobra"
)

// http://oid-info.com/get/1.3.6.1.4.1.11489.1.7.1.3
var customerOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11489, 1, 7, 1, 3}

func generateCertificateRequest(logger log.Logger, privateKey *rsa.PrivateKey, customerID string) ([]byte, error) {
	subj := pkix.Name{}
	// TODO(robin): add customer id back to subject
	// rawSubj := subj.ToRDNSequence()
	// rawSubj = append(rawSubj, []pkix.AttributeTypeAndValue{
	// 	{Type: customerOID, Value: customerID},
	// })
	// asn1Subj, err := asn1.Marshal(rawSubj)
	// if err != nil {
	// 	return nil, fmt.Errorf("error marshaling subject: %w", err)
	// }
	template := x509.CertificateRequest{
		Subject: subj,
		// RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating cert request: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}), nil
}

// enrollCmd represents the enroll command
var enrollCmd = &cobra.Command{
	Use:   "enroll <publisher>",
	Short: "enroll yourself as a developer for a publisher",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		publisher := args[0]
		ctx := context.Background()
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		channel, _ := cmd.Flags().GetString("channel")

		config, err := loadDevConfig()
		if err != nil {
			log.Fatal(logger, "unable to load developer config", "err", err)
		}

		log.Info(logger, "generating new private key")
		privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			log.Fatal(logger, "error generating private key", "err", err)
		}
		privateKeyBuf := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})
		config.PrivateKey = string(privateKeyBuf)
		if err := config.save(); err != nil {
			log.Fatal(logger, "error saving config", "err", err)
		}
		buf, err := generateCertificateRequest(logger, privateKey, config.CustomerID)
		if err != nil {
			log.Fatal(logger, "error generating certificate request", "err", err)
		}
		basepath := fmt.Sprintf("enroll/%s", publisher)
		resp, err := api.Put(ctx, channel, api.RegistryService, basepath, config.APIKey, bytes.NewReader(buf))
		if err != nil {
			log.Fatal(logger, "error sending cert request", "err", err)
		}
		respBuf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(logger, "error reading response body", "err", err)
		}
		resp.Body.Close()
		if resp.StatusCode != 201 {
			log.Fatal(logger, "error from api", "err", string(respBuf))
		}
		log.Info(logger, "recieved certificate from Pinpoint")
		config.Certificate = string(respBuf)
		if err := config.save(); err != nil {
			log.Fatal(logger, "error saving config", "err", err)
		}
		log.Info(logger, "successfully enrolled", "customer_id", config.CustomerID)
	},
}

func init() {
	rootCmd.AddCommand(enrollCmd)
	enrollCmd.Flags().String("channel", "stable", "the channel which can be set")
}
