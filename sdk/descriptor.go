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
	Publisher      Publisher    `json:"publisher" yaml:"publisher"`
	Capabilities   []string     `json:"capabilities" yaml:"capabilities"`
	Installation   Installation `json:"installation" yaml:"installation"`
	BuildDate      time.Time    `json:"-" yaml:"-"`
	BuildCommitSHA string       `json:"-" yaml:"-"`
}

// Publisher is the metadata about the publisher
type Publisher struct {
	Name      string `json:"name" yaml:"name"`
	URL       string `json:"url" yaml:"url"`
	AvatarURL string `json:"avatar_url" yaml:"avatar_url"`
}

// InstallationMode is the type of installation
type InstallationMode string

const (
	// InstallationModeCloud is for running the agent in the pinpoint cloud fully managed environment
	InstallationModeCloud InstallationMode = "cloud"
	// InstallationModeSelfManaged is for running the agent in a customers self managed environment
	InstallationModeSelfManaged InstallationMode = "selfmanaged"
)

// Installation is metadata about the installation
type Installation struct {
	Modes       []InstallationMode  `json:"modes" yaml:"modes"`
	Cloud       *InstallationConfig `json:"cloud" yaml:"cloud"`
	SelfManaged *InstallationConfig `json:"selfmanaged" yaml:"selfmanaged"`
}

// InstallationConfig is metadata about a specific installation mode
type InstallationConfig struct {
	Network      Network            `json:"network" yaml:"network"`
	Capabilities []string           `json:"capabilities" yaml:"capabilities"`
	Description  string             `json:"description" yaml:"description"`
	Options      InstallationOption `json:"options" yaml:"options"`
}

// Network specific environment details
type Network struct {
	Hostnames []string `json:"hostnames" yaml:"hostnames"`
}

// InstallationOption is metadata for the installation to be captured at configuration time
type InstallationOption struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Type        string `json:"type" yaml:"type"`
	Required    bool   `json:"required" yaml:"required"`
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
	tv, err := time.Parse(time.RFC3339, build)
	if err != nil {
		return nil, fmt.Errorf("error parsing the IntegrationBuildDate data: %w", err)
	}
	descriptor.BuildDate = tv
	descriptor.BuildCommitSHA = commit
	return &descriptor, nil
}
