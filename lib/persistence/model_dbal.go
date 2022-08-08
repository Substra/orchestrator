package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

type ModelDBAL interface {
	ModelExists(key string) (bool, error)
	GetModel(key string) (*asset.Model, error)
	QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error)
	GetComputeTaskOutputModels(key string) ([]*asset.Model, error)
	AddModel(m *asset.Model, identifier string) error
	UpdateModel(m *asset.Model) error
}

type ModelDBALProvider interface {
	GetModelDBAL() ModelDBAL
}
