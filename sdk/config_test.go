package sdk

import (
	"testing"

	pjson "github.com/pinpt/go-common/v10/json"
	ps "github.com/pinpt/go-common/v10/strings"
	"github.com/stretchr/testify/assert"
)

func TestConfigBasicAuth(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.BasicAuth)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"basic_auth": basicAuth{auth{"url"}, "a", "b"}}))))
	assert.NotNil(cfg.BasicAuth)
	assert.Equal("url", cfg.BasicAuth.URL)
	assert.Equal("a", cfg.BasicAuth.Username)
	assert.Equal("b", cfg.BasicAuth.Password)
}

func TestConfigOAuth2Auth(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.OAuth2Auth)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"oauth2_auth": oauth2Auth{auth{"url"}, "a", ps.Pointer("b")}}))))
	assert.NotNil(cfg.OAuth2Auth)
	assert.Equal("url", cfg.OAuth2Auth.URL)
	assert.Equal("a", cfg.OAuth2Auth.AccessToken)
	assert.Equal("b", *cfg.OAuth2Auth.RefreshToken)
}

func TestConfigAPIKeyAuth(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.APIKeyAuth)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"apikey_auth": apikeyAuth{auth{"url"}, "a"}}))))
	assert.NotNil(cfg.APIKeyAuth)
	assert.Equal("url", cfg.APIKeyAuth.URL)
	assert.Equal("a", cfg.APIKeyAuth.APIKey)
}

func TestConfigExclusions(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.Exclusions)
	exclusions := `# comment
pinpt/test_repo
foo/*
`
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"exclusions": exclusions}))))
	assert.NotNil(cfg.Exclusions)
	assert.True(cfg.Exclusions.Matches("pinpt/test_repo"))
	assert.False(cfg.Exclusions.Matches("pinpt/test_repo2"))
	assert.True(cfg.Exclusions.Matches("foo/test_repo"))
}

func TestConfigInclusions(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.Inclusions)
	inclusions := `# comment
pinpt/test_repo
foo/*
`
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"inclusions": inclusions}))))
	assert.NotNil(cfg.Inclusions)
	assert.True(cfg.Inclusions.Matches("pinpt/test_repo"))
	assert.False(cfg.Inclusions.Matches("pinpt/test_repo2"))
	assert.True(cfg.Inclusions.Matches("foo/test_repo"))
}

func TestUnstructured(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"foo": "bar"}))))
	found, val := cfg.GetString("foo")
	assert.Equal("bar", val)
	assert.True(found)
}
