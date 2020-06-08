package generator

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/integration-sdk/calendar"
	"github.com/pinpt/integration-sdk/codequality"
	"github.com/pinpt/integration-sdk/sourcecode"
	"github.com/pinpt/integration-sdk/work"
)

// IntegrationType used in the generator survey
type IntegrationType string

func (i IntegrationType) String() string {
	return string(i)
}

const (
	// IntegrationTypeSourcecode type source code
	IntegrationTypeSourcecode IntegrationType = "Source Code"
	// IntegrationTypeIssueTracking type work
	IntegrationTypeIssueTracking IntegrationType = "Issue Tracking"
	// IntegrationTypeCalendar type calendar
	IntegrationTypeCalendar IntegrationType = "Calendar"
	// IntegrationTypeCodeQuality type code quality
	IntegrationTypeCodeQuality IntegrationType = "Code Quality"
)

// Info needed info to generate the integration
type Info struct {
	Name             string   `json:"integration_name" survey:"integration_name"`
	PublisherName    string   `json:"publisher_name" survey:"publisher_name"`
	PublisherURL     string   `json:"publisher_url" survey:"publisher_url"`
	PublisherAvatar  string   `json:"publisher_avatar" survey:"publisher_avatar"`
	Identifier       string   `json:"identifier" survey:"identifier"`
	IntegrationTypes []string `json:"integration_types" survey:"integration_types"`

	Capabilities  []datamodel.ModelNameType
	PKG           string
	TitleCaseName string
	LowerCaseName string
	Date          string
}

// Generate generates a new project
func Generate(path string, info Info) error {

	handleError := func(err error) error {
		os.RemoveAll(path)
		return err
	}
	if err := os.MkdirAll(filepath.Join(path, "internal"), 0755); err != nil {
		return err
	}

	for _, t := range info.IntegrationTypes {
		switch IntegrationType(t) {
		case IntegrationTypeSourcecode:
			info.Capabilities = append(info.Capabilities,
				sourcecode.PullRequestCommitModelName,
				sourcecode.PullRequestCommentModelName,
			)
		case IntegrationTypeIssueTracking:
			info.Capabilities = append(info.Capabilities,
				work.UserModelName,
				work.ProjectModelName,
				work.IssueModelName,
				work.IssueCommentModelName,
				work.SprintModelName,
			)
		case IntegrationTypeCalendar:
			info.Capabilities = append(info.Capabilities,
				calendar.UserModelName,
				calendar.CalendarModelName,
				calendar.EventModelName,
			)
		case IntegrationTypeCodeQuality:
			info.Capabilities = append(info.Capabilities,
				codequality.MetricModelName,
				codequality.ProjectModelName,
			)
		}
	}

	info.TitleCaseName = strings.Title(info.Name)
	info.LowerCaseName = strings.ToLower(info.Name)
	info.Date = time.Now().String()

	if err := generate(path, "integration.go", info); err != nil {
		return handleError(err)
	}
	if err := generate(path, "go.mod", info); err != nil {
		return handleError(err)
	}
	if err := generate(path, "README.md", info); err != nil {
		return handleError(err)
	}
	if err := generate(path, "integration.yaml", info); err != nil {
		return handleError(err)
	}
	if err := generate(path, "internal/root.go", info); err != nil {
		return handleError(err)
	}
	return nil
}

func generate(path string, tmplfile string, info Info) error {
	b, err := Asset("template/" + tmplfile + ".tmpl")
	if err != nil {
		return err
	}
	tmpl, err := template.New(tmplfile).Parse(string(b))
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(path, tmplfile))
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, info)

}
