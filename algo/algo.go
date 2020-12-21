package algo

import (
	"github.com/substrafoundation/substra-orchestrator/asset"
)

// Algo is the representation of one of the element type stored in the ledger
type Algo struct {
	asset.Asset
	Key            string                 `json:"key"`
	Name           string                 `json:"name"`
	Checksum       string                 `json:"checksum"`
	StorageAddress string                 `json:"storage_address"`
	Description    *asset.ChecksumAddress `json:"description"`
	Owner          string                 `json:"owner"`
	Metadata       map[string]string      `json:"metadata"`
}

// CompositeAlgo is the representation of one of the element type stored in the ledger
type CompositeAlgo struct {
	Algo
}

// AggregateAlgo is the representation of one of the element type stored in the ledger
type AggregateAlgo struct {
	Algo
}
