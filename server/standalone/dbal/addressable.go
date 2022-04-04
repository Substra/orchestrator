package dbal

import "github.com/owkin/orchestrator/lib/asset"

func (d *DBAL) addAddressable(addressable *asset.Addressable) error {
	stmt := getStatementBuilder().
		Insert("addressables").
		Columns("storage_address", "checksum").
		Values(addressable.StorageAddress, addressable.Checksum)

	return d.exec(stmt)
}
