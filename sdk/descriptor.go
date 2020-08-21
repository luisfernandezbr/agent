package sdk

import (
	"encoding/base64"
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// Descriptor is metadata about what the integration supports
type Descriptor struct {
	Name           string       `json:"name" yaml:"name"`
	RefType        string       `json:"ref_type" yaml:"ref_type"`
	Description    string       `json:"description" yaml:"description"`
	AvatarURL      string       `json:"avatar_url" yaml:"avatar_url"`
	Capabilities   []string     `json:"capabilities" yaml:"capabilities"`
	Installation   Installation `json:"installation" yaml:"installation"`
	BuildDate      time.Time    `json:"-" yaml:"-"`
	BuildCommitSHA string       `json:"-" yaml:"-"`
}

// InstallationMode is the type of installation
type InstallationMode string

const (
	// InstallationModeCloud is for running the agent in the pinpoint cloud fully managed environment
	InstallationModeCloud InstallationMode = "cloud"
	// InstallationModeSelfManaged is for running the agent in a customers self managed environment
	InstallationModeSelfManaged InstallationMode = "selfmanaged"
)

// AuthorizationType is the type of authorization supported for a given location
type AuthorizationType string

const (
	// OAuth1AuthorizationType is the OAuth1 protocol
	OAuth1AuthorizationType AuthorizationType = "oauth1"
	// OAuth2AuthorizationType is the OAuth2 protocol
	OAuth2AuthorizationType AuthorizationType = "oauth2"
	// BasicAuthorizationType is the basic authentication protocol
	BasicAuthorizationType AuthorizationType = "basic"
	// APIKeyAuthorizationType is an apikey
	APIKeyAuthorizationType AuthorizationType = "apikey"
)

// Installation is metadata about the installation
type Installation struct {
	Modes       []InstallationMode  `json:"modes" yaml:"modes"`
	Cloud       *InstallationConfig `json:"cloud,omitempty" yaml:"cloud"`
	SelfManaged *InstallationConfig `json:"selfmanaged,omitempty" yaml:"selfmanaged"`
}

// InstallationConfig is metadata about a specific installation mode
type InstallationConfig struct {
	Capabilities  []string            `json:"capabilities,omitempty" yaml:"capabilities"`
	Authorization []AuthorizationType `json:"authorizations" yaml:"authorizations"`
}

// LoadDescriptor will load a descriptor from an integration
func LoadDescriptor(descriptorBuf, build, commit string) (*Descriptor, error) {
	buf, err := base64.StdEncoding.DecodeString(descriptorBuf)
	if err != nil {
		return nil, fmt.Errorf("error decoding the IntegrationDescriptor symbol: %w", err)
	}
	var descriptor Descriptor
	if err := yaml.Unmarshal(buf, &descriptor); err != nil {
		return nil, fmt.Errorf("error parsing the IntegrationDescriptor data: %w", err)
	}
	if build != "" {
		tv, err := time.Parse(time.RFC3339, build)
		if err != nil {
			return nil, fmt.Errorf("error parsing the IntegrationBuildDate data: %w", err)
		}
		descriptor.BuildDate = tv
		descriptor.BuildCommitSHA = commit
	}
	return &descriptor, nil
}
