package sdk

import "github.com/pinpt/integration-sdk/cicd"

// Type aliases for our exported datamodel types to create a stable version
// which Integrations depend on instead of directly depending on a specific
// version of the integration-sdk directly

// CICDBuild is the build
type CICDBuild = cicd.Build

// CICDBuildStartDate is the build start date
type CICDBuildStartDate = cicd.BuildStartDate

// CICDBuildEndDate is the build end date
type CICDBuildEndDate = cicd.BuildEndDate

// CICDBuildEnvironment is the build environment
type CICDBuildEnvironment = cicd.BuildEnvironment

// CICDBuildStatus is the build status
type CICDBuildStatus = cicd.BuildStatus

// CICDBuildStatusRunning is the enumeration value for running
const CICDBuildStatusRunning = cicd.BuildStatusRunning

// CICDBuildStatusPass is the enumeration value for pass
const CICDBuildStatusPass = cicd.BuildStatusPass

// CICDBuildStatusFail is the enumeration value for fail
const CICDBuildStatusFail = cicd.BuildStatusFail

// CICDBuildStatusCancel is the enumeration value for cancel
const CICDBuildStatusCancel = cicd.BuildStatusCancel

// CICDBuildEnvironmentProduction is the enumeration value for production
const CICDBuildEnvironmentProduction = cicd.BuildEnvironmentProduction

// CICDBuildEnvironmentDevelopment is the enumeration value for development
const CICDBuildEnvironmentDevelopment = cicd.BuildEnvironmentDevelopment

// CICDBuildEnvironmentBeta is the enumeration value for beta
const CICDBuildEnvironmentBeta = cicd.BuildEnvironmentBeta

// CICDBuildEnvironmentStaging is the enumeration value for staging
const CICDBuildEnvironmentStaging = cicd.BuildEnvironmentStaging

// CICDBuildEnvironmentTest is the enumeration value for test
const CICDBuildEnvironmentTest = cicd.BuildEnvironmentTest

// CICDBuildEnvironmentOther is the enumeration value for other
const CICDBuildEnvironmentOther = cicd.BuildEnvironmentOther

// CICDDeployment is the deployment
type CICDDeployment = cicd.Deployment

// CICDDeploymentStartDate is the start date
type CICDDeploymentStartDate = cicd.DeploymentStartDate

// CICDDeploymentEndDate is the end date
type CICDDeploymentEndDate = cicd.DeploymentEndDate

// CICDDeploymentEnvironment is the environment
type CICDDeploymentEnvironment = cicd.DeploymentEnvironment

// CICDDeploymentStatus is status
type CICDDeploymentStatus = cicd.DeploymentStatus

// CICDDeploymentEnvironmentProduction is the enumeration value for production
const CICDDeploymentEnvironmentProduction = cicd.DeploymentEnvironmentProduction

// CICDDeploymentEnvironmentDevelopment is the enumeration value for development
const CICDDeploymentEnvironmentDevelopment = cicd.DeploymentEnvironmentDevelopment

// CICDDeploymentEnvironmentBeta is the enumeration value for beta
const CICDDeploymentEnvironmentBeta = cicd.DeploymentEnvironmentBeta

// CICDDeploymentEnvironmentStaging is the enumeration value for staging
const CICDDeploymentEnvironmentStaging = cicd.DeploymentEnvironmentStaging

// CICDDeploymentEnvironmentTest is the enumeration value for test
const CICDDeploymentEnvironmentTest = cicd.DeploymentEnvironmentTest

// CICDDeploymentEnvironmentOther is the enumeration value for other
const CICDDeploymentEnvironmentOther = cicd.DeploymentEnvironmentOther

// CICDDeploymentStatusRunning is the enumeration value for running
const CICDDeploymentStatusRunning = cicd.DeploymentStatusRunning

// CICDDeploymentStatusPass is the enumeration value for pass
const CICDDeploymentStatusPass = cicd.DeploymentStatusPass

// CICDDeploymentStatusFail is the enumeration value for fail
const CICDDeploymentStatusFail = cicd.DeploymentStatusFail

// CICDDeploymentStatusCancel is the enumeration value for cancel
const CICDDeploymentStatusCancel = cicd.DeploymentStatusCancel

// NewCICDBuildID returns a new CICD.Build ID
func NewCICDBuildID(customerID string, refType string, refID string) string {
	return cicd.NewBuildID(customerID, refType, refID)
}

// NewCICDDeploymentID returns a new CICD.Deployment ID
func NewCICDDeploymentID(customerID string, refType string, refID string) string {
	return cicd.NewDeploymentID(customerID, refType, refID)
}
