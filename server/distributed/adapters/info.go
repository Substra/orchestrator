package adapters

import (
	"context"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/server/common"
	"github.com/substra/orchestrator/server/distributed/interceptors"
)

type InfoAdapter struct {
	asset.UnimplementedInfoServiceServer
}

func NewInfoAdapter() *InfoAdapter {
	return &InfoAdapter{}
}

func (a *InfoAdapter) QueryVersion(ctx context.Context, in *asset.QueryVersionParam) (*asset.QueryVersionResponse, error) {
	Invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}

	method := "orchestrator.info:QueryVersion"

	version := &asset.QueryVersionResponse{}

	err = Invocator.Call(ctx, method, in, version)

	version.Orchestrator = common.Version

	return version, err
}
