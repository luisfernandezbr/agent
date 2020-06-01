package sdk

import "github.com/pinpt/go-common/datamodel"

// Type aliases for our exported datamodel types to create a stable version
// which Integrations depend on instead of directly depending on a specific
// version of the integration-sdk directly

// Model is a data model type
type Model = datamodel.Model
