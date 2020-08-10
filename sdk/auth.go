package sdk

import "crypto/rsa"

// AuthManager is the authentication manager when handling pinpoint auth services
type AuthManager interface {
	// RefreshOAuth2Token will refresh the OAuth2 access token using the provided refreshToken and return a new access token
	RefreshOAuth2Token(refType string, refreshToken string) (string, error)
	// PrivateKey will return a private key stored by the integration UI
	PrivateKey(identifier Identifier) (*rsa.PrivateKey, error)
}
