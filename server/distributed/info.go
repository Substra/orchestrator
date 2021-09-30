package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
)

type InfoAdapter struct {
	asset.UnimplementedInfoServiceServer
}

func NewInfoAdapter() *InfoAdapter {
	return &InfoAdapter{}
}

func (a *InfoAdapter) QueryVersion(ctx context.Context, in *asset.QueryVersionParam) (*asset.QueryVersionResponse, error) {
	Invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}

	method := "orchestrator.info:QueryVersion"

	version := &asset.QueryVersionResponse{}

	err = Invocator.Call(ctx, method, in, version)

	version.Orchestrator = common.Version

	return version, err
}
