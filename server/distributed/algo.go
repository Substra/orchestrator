package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// AlgoAdapter is a grpc server exposing the same algo interface than standalone,
// but relies on a remote chaincode to actually manage the asset.
type AlgoAdapter struct {
	asset.UnimplementedAlgoServiceServer
}

// NewAlgoAdapter creates a Server
func NewAlgoAdapter() *AlgoAdapter {
	return &AlgoAdapter{}
}

// RegisterAlgo will add a new Algo to the network
func (a *AlgoAdapter) RegisterAlgo(ctx context.Context, in *asset.NewAlgo) (*asset.Algo, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.algo:RegisterAlgo"

	response := &asset.Algo{}

	err = invocator.Call(ctx, method, in, response)

	return response, err
}

// GetAlgo returns an algo from its key
func (a *AlgoAdapter) GetAlgo(ctx context.Context, query *asset.GetAlgoParam) (*asset.Algo, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.algo:GetAlgo"

	response := &asset.Algo{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

// QueryAlgos returns all known algos
func (a *AlgoAdapter) QueryAlgos(ctx context.Context, query *asset.QueryAlgosParam) (*asset.QueryAlgosResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.algo:QueryAlgos"

	response := &asset.QueryAlgosResponse{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}
