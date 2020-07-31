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
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"basic_auth": basicAuth{auth{"url", 0}, "a", "b"}}))))
	assert.NotNil(cfg.BasicAuth)
	assert.Equal("url", cfg.BasicAuth.URL)
	assert.Equal("a", cfg.BasicAuth.Username)
	assert.Equal("b", cfg.BasicAuth.Password)
}

func TestConfigOAuth2Auth(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.OAuth2Auth)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"oauth2_auth": oauth2Auth{auth{"url", 0}, "a", ps.Pointer("b"), nil}}))))
	assert.NotNil(cfg.OAuth2Auth)
	assert.Equal("url", cfg.OAuth2Auth.URL)
	assert.Equal("a", cfg.OAuth2Auth.AccessToken)
	assert.Equal("b", *cfg.OAuth2Auth.RefreshToken)
}

func TestConfigOAuth1Auth(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.OAuth1Auth)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"oauth1_auth": oauth1Auth{auth{"url", 0}, "consumer", "token", "secret"}}))))
	assert.NotNil(cfg.OAuth1Auth)
	assert.Equal("url", cfg.OAuth1Auth.URL)
	assert.Equal("consumer", cfg.OAuth1Auth.ConsumerKey)
	assert.Equal("token", cfg.OAuth1Auth.Token)
	assert.Equal("secret", cfg.OAuth1Auth.Secret)
}

func TestConfigOAuth2AuthWithScopes(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.OAuth2Auth)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"oauth2_auth": oauth2Auth{auth{"url", 0}, "a", ps.Pointer("b"), ps.Pointer("scope")}}))))
	assert.NotNil(cfg.OAuth2Auth)
	assert.Equal("url", cfg.OAuth2Auth.URL)
	assert.Equal("a", cfg.OAuth2Auth.AccessToken)
	assert.Equal("b", *cfg.OAuth2Auth.RefreshToken)
	assert.Equal("scope", *cfg.OAuth2Auth.Scopes)
}

func TestConfigAPIKeyAuth(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.APIKeyAuth)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"apikey_auth": apikeyAuth{auth{"url", 0}, "a"}}))))
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
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"exclusions": map[string]string{"pinpt": exclusions}}))))
	assert.NotNil(cfg.Exclusions)
	assert.True(cfg.Exclusions.Matches("pinpt", "pinpt/test_repo"))
	assert.False(cfg.Exclusions.Matches("pinpt", "pinpt/test_repo2"))
	assert.True(cfg.Exclusions.Matches("pinpt", "foo/test_repo"))
}

func TestConfigInclusions(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.Nil(cfg.Inclusions)
	inclusions := `# comment
pinpt/test_repo
foo/*
`
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"inclusions": map[string]string{"pinpt": inclusions}}))))
	assert.NotNil(cfg.Inclusions)
	assert.True(cfg.Inclusions.Matches("pinpt", "pinpt/test_repo"))
	assert.False(cfg.Inclusions.Matches("pinpt", "pinpt/test_repo2"))
	assert.True(cfg.Inclusions.Matches("pinpt", "foo/test_repo"))
}

func TestUnstructured(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	assert.NoError(cfg.Parse([]byte(pjson.Stringify(map[string]interface{}{"foo": "bar"}))))
	found, val := cfg.GetString("foo")
	assert.Equal("bar", val)
	assert.True(found)
}

func TestConfigFromString(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	configstr := `{"integration_type":"CLOUD","api_key":"785f10bc83f703a2363d1d9809eec9b117a8926a","profile":{"Provider":"github","ID":"MDQ6VXNlcjYwMjc=","DisplayName":"Jeff Haynie","EmailAddress":"jhaynie@gmail.com","AvatarURL":"https://avatars1.githubusercontent.com/u/6027?v=4","Company":"pinpt","Integration":{"emails":[{"email":"jhaynie@gmail.com","validated":true,"primary":true},{"email":"jhaynie@pinpt.co","validated":true,"primary":false},{"email":"jhaynie@pinpt.com","validated":true,"primary":false},{"email":"jhaynie@pinpoint.com","validated":true,"primary":false}],"username":"jhaynie","id":"MDQ6VXNlcjYwMjc=","avatars":"https://avatars1.githubusercontent.com/u/6027?v=4","name":"Jeff Haynie","location":"Austin, TX","url":"https://github.com/jhaynie","refType":"github","auth":{"accessToken":"785f10bc83f703a2363d1d9809eec9b117a8926a","scopes":"repo,user:email,read:user,read:org","created":1593008668174}}},"accounts":{"pinpt":{"id":"MDEyOk9yZ2FuaXphdGlvbjI0NDAwNTI2","name":"Pinpoint","description":"Pinpoint uses data science to advance the way people and teams deliver software","avatarUrl":"https://avatars3.githubusercontent.com/u/24400526?v=4","login":"pinpt","repositories":{"totalCount":193},"type":"ORG","public":false},"jhaynie":{"id":"MDQ6VXNlcjYwMjc=","name":"Jeff Haynie","login":"jhaynie","description":"open source developer, co-founder/CEO of pinpoint.com, previous co-founder/CEO of Appcelerator","avatarUrl":"https://avatars1.githubusercontent.com/u/6027?u=1dfa46ade92e7c8b202d761a5344aa9c8630b70c&v=4","repositories":{"totalCount":63},"organizations":{"nodes":[{"id":"MDEyOk9yZ2FuaXphdGlvbjI0NDAwNTI2","name":"Pinpoint","description":"Pinpoint uses data science to advance the way people and teams deliver software","avatarUrl":"https://avatars3.githubusercontent.com/u/24400526?v=4","login":"pinpt","repositories":{"totalCount":193}}]},"type":"USER","public":false},"facebook":{"id":"facebook","name":"Facebook","description":"We are working to build community through open source technology. NB: members must have two-factor auth.","avatarUrl":"https://avatars3.githubusercontent.com/u/69631?v=4","repositories":{"totalCount":125},"type":"ORG","public":true}},"exclusions":{"pinpt":"pinpt/soc2*"}}`
	assert.NoError(cfg.Parse([]byte(configstr)))
	assert.False(cfg.Exclusions.Matches("pinpt", "pinpt/foo"))
	assert.False(cfg.Exclusions.Matches("foobar", "pinpt/foo"))
	assert.True(cfg.Exclusions.Matches("pinpt", "pinpt/soc2_foo"))
	assert.Equal(CloudIntegration, cfg.IntegrationType)
}

func TestConfigExclusionsNegate(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	configstr := `{"exclusions":{"pinpt":"!*\npinpt/robin"}}`
	assert.NoError(cfg.Parse([]byte(configstr)))
	assert.False(cfg.Exclusions.Matches("pinpt", "pinpt/foo"))
	assert.False(cfg.Exclusions.Matches("pinpt", "pinpt/soc2_foo"))
	assert.False(cfg.Exclusions.Matches("pinpt", "pinpt/bar"))
	assert.True(cfg.Exclusions.Matches("pinpt", "pinpt/robin"))
}

func TestConfigFrom(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	configstr := `{"integration_type":"CLOUD","api_key":"785f10bc83f703a2363d1d9809eec9b117a8926a","profile":{"Provider":"github","ID":"MDQ6VXNlcjYwMjc=","DisplayName":"Jeff Haynie","EmailAddress":"jhaynie@gmail.com","AvatarURL":"https://avatars1.githubusercontent.com/u/6027?v=4","Company":"pinpt","Integration":{"emails":[{"email":"jhaynie@gmail.com","validated":true,"primary":true},{"email":"jhaynie@pinpt.co","validated":true,"primary":false},{"email":"jhaynie@pinpt.com","validated":true,"primary":false},{"email":"jhaynie@pinpoint.com","validated":true,"primary":false}],"username":"jhaynie","id":"MDQ6VXNlcjYwMjc=","avatars":"https://avatars1.githubusercontent.com/u/6027?v=4","name":"Jeff Haynie","location":"Austin, TX","url":"https://github.com/jhaynie","refType":"github","auth":{"accessToken":"785f10bc83f703a2363d1d9809eec9b117a8926a","scopes":"repo,user:email,read:user,read:org","created":1593008668174}}},"accounts":{"pinpt":{"id":"MDEyOk9yZ2FuaXphdGlvbjI0NDAwNTI2","name":"Pinpoint","description":"Pinpoint uses data science to advance the way people and teams deliver software","avatarUrl":"https://avatars3.githubusercontent.com/u/24400526?v=4","login":"pinpt","repositories":{"totalCount":193},"type":"ORG","public":false},"jhaynie":{"id":"MDQ6VXNlcjYwMjc=","name":"Jeff Haynie","login":"jhaynie","description":"open source developer, co-founder/CEO of pinpoint.com, previous co-founder/CEO of Appcelerator","avatarUrl":"https://avatars1.githubusercontent.com/u/6027?u=1dfa46ade92e7c8b202d761a5344aa9c8630b70c&v=4","repositories":{"totalCount":63},"organizations":{"nodes":[{"id":"MDEyOk9yZ2FuaXphdGlvbjI0NDAwNTI2","name":"Pinpoint","description":"Pinpoint uses data science to advance the way people and teams deliver software","avatarUrl":"https://avatars3.githubusercontent.com/u/24400526?v=4","login":"pinpt","repositories":{"totalCount":193}}]},"type":"USER","public":false},"facebook":{"id":"facebook","name":"Facebook","description":"We are working to build community through open source technology. NB: members must have two-factor auth.","avatarUrl":"https://avatars3.githubusercontent.com/u/69631?v=4","repositories":{"totalCount":125},"type":"ORG","public":true}},"exclusions":{"pinpt":"pinpt/soc2*"}}`
	assert.NoError(cfg.Parse([]byte(configstr)))
	var profile struct {
		Provider string
	}
	assert.NoError(cfg.From("profile", &profile))
	assert.Equal("github", profile.Provider)
}

func TestConfigAccounts(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig(nil)
	configstr := `{"accounts":{"bitbucket":{"id":"bitbucket", "type":"ORG", "public":true}, "john_smith":{"id":"john_smith", "type":"USER", "public":false}}}`
	assert.NoError(cfg.Parse([]byte(configstr)))
	var accounts ConfigAccounts
	assert.NoError(cfg.From("accounts", &accounts))

	assert.Equal("bitbucket", accounts["bitbucket"].ID)
	assert.Equal(ConfigAccountTypeOrg, accounts["bitbucket"].Type)
	assert.Equal(true, accounts["bitbucket"].Public)

	assert.Equal("john_smith", accounts["john_smith"].ID)
	assert.Equal(ConfigAccountTypeUser, accounts["john_smith"].Type)
	assert.Equal(false, accounts["john_smith"].Public)
}
