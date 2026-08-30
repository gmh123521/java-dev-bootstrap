package detection

import (
	"context"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
)

type Status string

const (
	StatusInstalled Status = "installed"
	StatusOutdated  Status = "outdated"
	StatusMissing   Status = "missing"
	StatusError     Status = "error"
)

type Result struct {
	Status  Status
	Version string
	Source  string
	Path    string
	Detail  string
	Err     error
}

type Detector interface {
	Detect(context.Context, model.Package) Result
}
