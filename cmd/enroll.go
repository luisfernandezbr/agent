package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pinpt/go-common/v10/api"
	pjson "github.com/pinpt/go-common/v10/json"
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

var identifierRegexp = regexp.MustCompile("[\\w]{3,12}")

func createValidateIdentifier(channel string, apikey string, customerID string) survey.Validator {
	return func(val interface{}) error {
		if sv, ok := val.(string); ok {
			if sv == "" {
				return errors.New("Value is required")
			}
			if identifierRegexp.MatchString(sv) {
				resp, err := api.Get(context.Background(), channel, api.RegistryService, "/validate/"+sv, apikey)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				var rjson struct {
					Found      bool   `json:"found"`
					CustomerID string `json:"customer_id"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&rjson); err != nil {
					return fmt.Errorf("error parsing json response from validate: %w", err)
				}
				if !rjson.Found || rjson.CustomerID == customerID {
					return nil
				}
				return fmt.Errorf("the identifier %s is already taken", sv)
			}
			return fmt.Errorf("identifier must be between 3-12 alphanumeric characters")
		}
		return errors.New("invalid type")
	}
}

func validateAvatar(val interface{}) error {
	if sv, ok := val.(string); ok {
		if sv == "" {
			return errors.New("Value is required")
		}
		resp, err := http.Get(sv)
		if err != nil {
			return fmt.Errorf("error fetching %s", sv)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error fetching %s (status code: %d)", sv, resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if strings.Contains(ct, "png") || strings.Contains(ct, "gif") || strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg") || strings.Contains(ct, "image/") {
			return nil
		}
		return fmt.Errorf("invalid avatar image. must be a GIF, JPEG or PNG")
	}
	return errors.New("invalid type")
}

func validateURL(val interface{}) error {
	if sv, ok := val.(string); ok {
		if sv == "" {
			return errors.New("Value is required")
		}
		resp, err := http.Get(sv)
		if err != nil {
			return fmt.Errorf("error fetching %s", sv)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error fetching %s (status code: %d)", sv, resp.StatusCode)
		}
		return nil
	}
	return errors.New("invalid type")
}

// enrollCmd represents the enroll command
var enrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "enroll to create a developer account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		channel, _ := cmd.Flags().GetString("channel")
		config, err := loadDevConfig()
		if err != nil {
			log.Fatal(logger, "unable to load developer config", "err", err)
		}
		if config.expired() {
			log.Fatal(logger, "your login session has expired. please login again")
		}
		if config.Channel != channel {
			log.Fatal(logger, "your login session was for a different channel. please login again")
		}

		banner()
		fmt.Println()
		fmt.Println("🚀 Time to enroll you in the Pinpoint Developer Program")
		fmt.Println()
		fmt.Println("We need a few bits of information to continue ...")
		fmt.Println()

		var result struct {
			Name        string `json:"name" survey:"name"`
			Identifier  string `json:"identifier" survey:"identifier"`
			Description string `json:"description" survey:"description"`
			Avatar      string `json:"avatar_url" survey:"avatar_url"`
			URL         string `json:"url" survey:"url"`
			Certificate string `json:"csr"`
		}
		if err := survey.Ask([]*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Your Publisher Name:",
					Help:    "Your name such as Pinpoint Software, Inc",
				},
				Validate: survey.Required,
			},
			{
				Name: "identifier",
				Prompt: &survey.Input{
					Message: "Your Publisher Short Identifier:",
					Help:    "Your short identifier must be unique and should contain no spaces or special characters such as pinpt",
				},
				Validate: createValidateIdentifier(channel, config.APIKey, config.CustomerID),
			},
			{
				Name: "avatar_url",
				Prompt: &survey.Input{
					Message: "Your Publisher Avatar:",
					Help:    "A url to your avatar image in PNG, GIF or JPEG format",
				},
				Validate: validateAvatar,
			},
			{
				Name: "url",
				Prompt: &survey.Input{
					Message: "Your Publisher URL:",
					Help:    "The URL to your homepage or URL to your integration",
				},
				Validate: validateURL,
			},
			{
				Name: "description",
				Prompt: &survey.Input{
					Message: "Your Publisher Bio:",
					Help:    "A short description of your bio or company background",
				},
				Validate: survey.ComposeValidators(survey.Required, survey.MaxLength(255), survey.MinLength(20)),
			},
		}, &result); err != nil {
			os.Exit(1)
		}

		fmt.Println()

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
		result.Certificate = string(buf)
		resp, err := api.Put(ctx, channel, api.RegistryService, "enroll", config.APIKey, strings.NewReader(pjson.Stringify(result)))
		if err != nil {
			log.Fatal(logger, "error sending cert request", "err", err)
		}
		respBuf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(logger, "error reading response body", "err", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			log.Fatal(logger, "error from api", "err", string(respBuf))
		}
		log.Info(logger, "recieved certificate from Pinpoint")
		config.Certificate = string(respBuf)
		config.PublisherRefType = result.Identifier
		if err := config.save(); err != nil {
			log.Fatal(logger, "error saving config", "err", err)
		}
		log.Info(logger, "successfully enrolled. you can now publish integrations! go forth and build 🎉", "customer_id", config.CustomerID)
	},
}

func init() {
	rootCmd.AddCommand(enrollCmd)
	enrollCmd.Flags().String("channel", "stable", "the channel which can be set")
}
