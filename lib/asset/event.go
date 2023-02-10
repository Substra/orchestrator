package asset

import (
	"fmt"

	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MarshalEventAsset returns the JSON encoding of the Asset
// of the provided Event.
func MarshalEventAsset(event *Event) ([]byte, error) {
	var m proto.Message
	switch a := event.Asset.(type) {
	case *Event_Function:
		m = event.GetFunction()
	case *Event_ComputePlan:
		m = event.GetComputePlan()
	case *Event_ComputeTask:
		m = event.GetComputeTask()
	case *Event_DataManager:
		m = event.GetDataManager()
	case *Event_DataSample:
		m = event.GetDataSample()
	case *Event_FailureReport:
		m = event.GetFailureReport()
	case *Event_Model:
		m = event.GetModel()
	case *Event_Organization:
		m = event.GetOrganization()
	case *Event_Performance:
		m = event.GetPerformance()
	case *Event_ComputeTaskOutputAsset:
		m = event.GetComputeTaskOutputAsset()
	default:
		return nil, orcerrors.NewInternal(fmt.Sprintf("unsupported asset %T", a))
	}

	return protojson.Marshal(m)
}

// UnmarshalEventAsset parses the JSON-encoded data according to the assetKind
// and stores the result in the Asset field of the provided event.
func UnmarshalEventAsset(b []byte, event *Event, assetKind AssetKind) error {
	switch assetKind {
	case AssetKind_ASSET_FUNCTION:
		function := new(Function)
		if err := protojson.Unmarshal(b, function); err != nil {
			return err
		}
		event.Asset = &Event_Function{Function: function}
	case AssetKind_ASSET_COMPUTE_PLAN:
		plan := new(ComputePlan)
		if err := protojson.Unmarshal(b, plan); err != nil {
			return err
		}
		event.Asset = &Event_ComputePlan{ComputePlan: plan}
	case AssetKind_ASSET_COMPUTE_TASK:
		task := new(ComputeTask)
		if err := protojson.Unmarshal(b, task); err != nil {
			return err
		}
		event.Asset = &Event_ComputeTask{ComputeTask: task}
	case AssetKind_ASSET_DATA_MANAGER:
		manager := new(DataManager)
		if err := protojson.Unmarshal(b, manager); err != nil {
			return err
		}
		event.Asset = &Event_DataManager{DataManager: manager}
	case AssetKind_ASSET_DATA_SAMPLE:
		sample := new(DataSample)
		if err := protojson.Unmarshal(b, sample); err != nil {
			return err
		}
		event.Asset = &Event_DataSample{DataSample: sample}
	case AssetKind_ASSET_FAILURE_REPORT:
		report := new(FailureReport)
		if err := protojson.Unmarshal(b, report); err != nil {
			return err
		}
		event.Asset = &Event_FailureReport{FailureReport: report}
	case AssetKind_ASSET_MODEL:
		model := new(Model)
		if err := protojson.Unmarshal(b, model); err != nil {
			return err
		}
		event.Asset = &Event_Model{Model: model}
	case AssetKind_ASSET_ORGANIZATION:
		organization := new(Organization)
		if err := protojson.Unmarshal(b, organization); err != nil {
			return err
		}
		event.Asset = &Event_Organization{Organization: organization}
	case AssetKind_ASSET_PERFORMANCE:
		perf := new(Performance)
		if err := protojson.Unmarshal(b, perf); err != nil {
			return err
		}
		event.Asset = &Event_Performance{Performance: perf}
	case AssetKind_ASSET_COMPUTE_TASK_OUTPUT_ASSET:
		out := new(ComputeTaskOutputAsset)
		if err := protojson.Unmarshal(b, out); err != nil {
			return err
		}
		event.Asset = &Event_ComputeTaskOutputAsset{ComputeTaskOutputAsset: out}
	default:
		return orcerrors.NewInternal(fmt.Sprintf("unsupported asset kind %T", assetKind))
	}

	return nil
}
