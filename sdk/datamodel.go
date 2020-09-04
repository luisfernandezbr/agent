package sdk

import "github.com/pinpt/go-common/v10/datamodel"

// Type aliases for our exported datamodel types to create a stable version
// which Integrations depend on instead of directly depending on a specific
// version of the integration-sdk directly

// Model is a data model type
type Model = datamodel.Model

// PartialModel is a partial datamodel type with all optional fields
type PartialModel = datamodel.PartialModel

// IntegrationModel has some extra methods that exist on sourcedata types
type IntegrationModel interface {
	datamodel.Model
	GetIntegrationInstanceID() *string
	SetIntegrationInstanceID(string)
}
