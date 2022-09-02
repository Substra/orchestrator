package persistence

import (
	"github.com/substra/orchestrator/lib/asset"
)

type ModelDBAL interface {
	ModelExists(key string) (bool, error)
	GetModel(key string) (*asset.Model, error)
	GetComputeTaskOutputModels(key string) ([]*asset.Model, error)
	AddModel(m *asset.Model, identifier string) error
	UpdateModel(m *asset.Model) error
}

type ModelDBALProvider interface {
	GetModelDBAL() ModelDBAL
}
