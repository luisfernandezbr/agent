module github.com/pinpt/agent.next

go 1.14

require (
	github.com/99designs/keyring v1.1.5
	github.com/AlecAivazis/survey/v2 v2.0.7
	github.com/fatih/color v1.9.0
	github.com/go-redis/redis/v8 v8.0.0-beta.2
	github.com/pinpt/go-common/v10 v10.0.3
	github.com/pinpt/integration-sdk v0.0.1004
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.0
	go.opentelemetry.io/otel v0.6.0 // indirect
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

// REMOVE once this PR is merged https://github.com/99designs/keyring/pull/59
replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
