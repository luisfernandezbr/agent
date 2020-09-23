package sdk

import "crypto/rsa"

// AuthManager is the authentication manager when handling pinpoint auth services
type AuthManager interface {
	// RefreshOAuth2Token will refresh the OAuth2 access token using the provided refreshToken and return a new access token
	RefreshOAuth2Token(refType string, refreshToken string) (string, error)
	// PrivateKey will return a private key stored by the integration UI
	PrivateKey(identifier Identifier) (*rsa.PrivateKey, error)
}

// OAuth1Identity returns an identity for an OAuth1 integration
type OAuth1Identity struct {
	Name      string  `json:"name"`
	RefID     string  `json:"ref_id"`
	AvatarURL *string `json:"avatar_url"`
	Email     *string `json:"email"`
}

// OAuth1Integration is implemented by integrations that support OAuth1 identity
type OAuth1Integration interface {
	// IdentifyOAuth1User should be implemented to get an identity for a user tied to the private key
	IdentifyOAuth1User(identifier Identifier, url string, privateKey *rsa.PrivateKey, consumerKey string, consumerSecret string, token string, tokenSecret string) (*OAuth1Identity, error)
}
