package sdk

// User is an integration user that has been linked to the integration instance for a customer
type User struct {
	ID                  string   `json:"_id"`
	Name                string   `json:"name"`
	Emails              []string `json:"emails"`
	RefID               string   `json:"ref_id"`
	OAuth1Authorization *struct {
		Date        int64  `json:"date_ts"`
		ConsumerKey string `json:"consumer_key"`
		Token       string `json:"oauth_token"`
		TokenSecret string `json:"oauth_token_secret"`
	} `json:"oauth1_authorization"`
	OAuth2Authorization *struct {
		Date         int64   `json:"date_ts"`
		AccessToken  string  `json:"token"`
		RefreshToken *string `json:"refresh_token"`
		Scopes       string  `json:"scopes"`
	} `json:"oauth2_authorization"`
}

// UserManager is a control interface for getting users
type UserManager interface {
	// Users will return the integration users for a given integration instance
	Users(control Control) ([]User, error)
}
