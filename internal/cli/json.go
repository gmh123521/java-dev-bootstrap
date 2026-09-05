package cli

import (
	"encoding/json"

	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/service"
)

type jsonPackage struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonPlanItem struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Skipped   bool          `json:"skipped"`
	Detection jsonDetection `json:"detection"`
	Command   jsonCommand   `json:"command"`
}

type jsonDetection struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	Source  string `json:"source,omitempty"`
	Path    string `json:"path,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Error   string `json:"error,omitempty"`
}

type jsonCommand struct {
	Program string   `json:"program"`
	Args    []string `json:"args"`
}

type jsonInstallReport struct {
	Succeeded          int      `json:"succeeded"`
	Skipped            int      `json:"skipped"`
	Failed             int      `json:"failed"`
	Retried            int      `json:"retried"`
	Verified           int      `json:"verified"`
	VerificationFailed int      `json:"verification_failed"`
	Errors             []string `json:"errors,omitempty"`
	VerificationErrors []string `json:"verification_errors,omitempty"`
}

type jsonDiagnostic struct {
	Level      string `json:"level"`
	Name       string `json:"name"`
	Current    string `json:"current"`
	Suggestion string `json:"suggestion,omitempty"`
}

type jsonPrerequisite struct {
	Name       string `json:"name"`
	Current    string `json:"current"`
	Suggestion string `json:"suggestion,omitempty"`
	OK         bool   `json:"ok"`
}

type jsonSetup struct {
	Ready        bool   `json:"ready"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	Manager      string `json:"manager"`
	ManagerState string `json:"manager_state"`
	Suggestion   string `json:"suggestion,omitempty"`
	NextCommand  string `json:"next_command"`
}

func formatPackagesJSON(packages []model.Package) (string, error) {
	items := make([]jsonPackage, 0, len(packages))
	for _, pkg := range packages {
		items = append(items, jsonPackage{ID: pkg.ID, Name: pkg.Name, Description: pkg.Description})
	}
	return marshalJSON(items)
}

func formatPlanJSON(items []service.PlanItem) (string, error) {
	output := make([]jsonPlanItem, 0, len(items))
	for _, item := range items {
		detection := jsonDetection{
			Status:  string(item.Detection.Status),
			Version: item.Detection.Version,
			Source:  item.Detection.Source,
			Path:    item.Detection.Path,
			Detail:  item.Detection.Detail,
		}
		if item.Detection.Err != nil {
			detection.Error = item.Detection.Err.Error()
		}
		output = append(output, jsonPlanItem{
			ID:        item.Package.ID,
			Name:      item.Package.Name,
			Skipped:   item.Skipped,
			Detection: detection,
			Command:   jsonCommand{Program: item.Command.Program, Args: item.Command.Args},
		})
	}
	return marshalJSON(output)
}

func formatInstallReportJSON(report service.InstallReport) (string, error) {
	return marshalJSON(jsonInstallReport{
		Succeeded:          report.Succeeded,
		Skipped:            report.Skipped,
		Failed:             report.Failed,
		Retried:            report.Retried,
		Verified:           report.Verified,
		VerificationFailed: report.VerificationFailed,
		Errors:             errorStrings(report.Errors),
		VerificationErrors: errorStrings(report.VerificationErrors),
	})
}

func formatDiagnosticsJSON(items []detection.Diagnostic) (string, error) {
	output := make([]jsonDiagnostic, 0, len(items))
	for _, item := range items {
		output = append(output, jsonDiagnostic{Level: string(item.Level), Name: item.Name, Current: item.Current, Suggestion: item.Suggestion})
	}
	return marshalJSON(output)
}

func formatPrerequisitesJSON(items []platform.PrerequisiteItem) (string, error) {
	output := make([]jsonPrerequisite, 0, len(items))
	for _, item := range items {
		output = append(output, jsonPrerequisite{Name: item.Name, Current: item.Current, Suggestion: item.Suggestion, OK: item.OK})
	}
	return marshalJSON(output)
}

func formatSetupJSON(result platform.SetupResult) (string, error) {
	return marshalJSON(jsonSetup{
		Ready:        result.Ready,
		Platform:     result.Platform,
		Architecture: result.Architecture,
		Manager:      result.Manager,
		ManagerState: result.ManagerState,
		Suggestion:   result.Suggestion,
		NextCommand:  result.NextCommand,
	})
}

func errorStrings(errors []error) []string {
	if len(errors) == 0 {
		return nil
	}
	output := make([]string, 0, len(errors))
	for _, err := range errors {
		if err != nil {
			output = append(output, err.Error())
		}
	}
	return output
}

func marshalJSON(value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
