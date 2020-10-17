package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/api"
)

// RefreshOAuth2Token will fetch a new API Key / Access Token from a refresh token
func RefreshOAuth2Token(client *http.Client, channel string, provider string, refreshToken string) (string, error) {
	theurl := sdk.JoinURL(
		api.BackendURL(api.AuthService, channel),
		fmt.Sprintf("oauth2/%s/refresh/%s", provider, url.PathEscape(refreshToken)),
	)
	req, err := http.NewRequest(http.MethodGet, theurl, nil)
	if err != nil {
		return "", err
	}
	api.SetUserAgent(req)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	var res struct {
		AccessToken string `json:"access_token"`
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	if res.AccessToken == "" {
		return "", errors.New("new token not returned, refresh_token might be bad")
	}
	return res.AccessToken, nil
}
