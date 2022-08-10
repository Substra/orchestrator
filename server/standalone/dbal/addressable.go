package dbal

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/substra/orchestrator/lib/asset"
)

func (d *DBAL) addAddressable(addressable *asset.Addressable) error {
	stmt := getStatementBuilder().
		Insert("addressables").
		Columns("storage_address", "checksum").
		Values(addressable.StorageAddress, addressable.Checksum)

	return d.exec(stmt)
}

func (d *DBAL) deleteAddressable(storageAddress string) error {
	stmt := getStatementBuilder().
		Delete("addressables").
		Where(sq.Eq{"storage_address": storageAddress})

	return d.exec(stmt)
}
