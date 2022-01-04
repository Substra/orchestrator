package asset

// Kind represent the type of assets handled by the orchestrator
type Kind = string

var (
	// NodeKind is the type of Node assets
	NodeKind Kind = "node"
	// MetricKind is the type of Metric assets
	MetricKind = "metric"
	// DataSampleKind is the type of DataSample assets
	DataSampleKind = "datasample"
	// AlgoKind is the type of Algo assets
	AlgoKind = "algo"
	// DataManagerKind is the type of DataManager assets
	DataManagerKind = "datamanager"
	// ComputeTaskKind is the type of ComputeTask assets
	ComputeTaskKind = "computetask"
	// ComputePlanKind is the type of ComputePlan assets
	ComputePlanKind = "computeplan"
	// ModelKind is the type of Model assets
	ModelKind = "model"
	// PerformanceKind is the type of Performance assets
	PerformanceKind = "performance"
	// FailureReportKind is the type of FailureReport assets
	FailureReportKind = "failurereport"
)
