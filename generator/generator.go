package generator

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
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
	Pkg              string   `json:"pkg" survey:"pkg"`

	Capabilities  []datamodel.ModelNameType
	Dir           string
	TitleCaseName string
	LowerCaseName string
	Date          string
	GitTag        string
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
	info.GitTag = gitTag
	for _, name := range AssetNames() {
		// trim off template/
		thename := strings.Replace(name[9:], ".tmpl", "", -1)
		if err := generate(path, thename, info); err != nil {
			return handleError(err)
		}
	}

	return nil
}

func generate(path string, tmplfile string, info Info) error {
	b, err := Asset("template/" + tmplfile + ".tmpl")
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
		// not all files are tmpl so check and see if an as-is file
		b, err = Asset("template/" + tmplfile)
		if err != nil {
			return err
		}
	}
	tmpl, err := template.New(tmplfile).Parse(string(b))
	if err != nil {
		return err
	}
	if tmplfile == "internal/root.go" {
		tmplfile = "internal/" + info.LowerCaseName + ".go"
	}
	fn := filepath.Join(path, tmplfile)
	os.MkdirAll(filepath.Dir(fn), 0700)
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, info); err != nil {
		return err
	}
	if strings.HasSuffix(tmplfile, ".go") {
		b, err = format.Source(tpl.Bytes())
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(fn, b, 0644); err != nil {
			return err
		}
		if err := exec.Command("goimports", "-w", fn).Run(); err != nil {
			return err
		}
	} else {
		if err := ioutil.WriteFile(fn, tpl.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}
