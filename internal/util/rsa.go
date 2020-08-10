package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

// ParsePrivateKey will parse a private key out of pem data in either PKCS1 or PKCS8 form
func ParsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("no pem data in private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		if strings.Contains(err.Error(), "use ParsePKCS8PrivateKey instead for this key format") {
			k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err == nil {
				return k.(*rsa.PrivateKey), nil
			}
		}
		return nil, fmt.Errorf("error parsing key: %w", err)
	}
	return key, nil
}
