package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
)

type FailureReportDBAL interface {
	GetFailureReport(computeTaskKey string) (*asset.FailureReport, error)
	AddFailureReport(f *asset.FailureReport) error
}

type FailureReportDBALProvider interface {
	GetFailureReportDBAL() FailureReportDBAL
}
