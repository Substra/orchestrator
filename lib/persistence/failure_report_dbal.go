package persistence

import (
	"github.com/substra/orchestrator/lib/asset"
)

type FailureReportDBAL interface {
	GetFailureReport(assetKey string) (*asset.FailureReport, error)
	AddFailureReport(f *asset.FailureReport) error
}

type FailureReportDBALProvider interface {
	GetFailureReportDBAL() FailureReportDBAL
}
