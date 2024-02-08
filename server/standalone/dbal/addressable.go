package dbal

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/substra/orchestrator/lib/asset"
)

func (d *DBAL) addAddressable(addressable *asset.Addressable, ignore_conflict bool) error {
	stmt := getStatementBuilder().
		Insert("addressables").
		Columns("storage_address", "checksum").
		Values(addressable.StorageAddress, addressable.Checksum)

	if ignore_conflict {
		stmt = stmt.Suffix("ON CONFLICT DO NOTHING")
	}

	return d.exec(stmt)
}

func (d *DBAL) AddressableExists(storageAddress string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(storage_address)").
		From("addressables").
		Where(sq.Eq{"storage_address": storageAddress})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

func (d *DBAL) updateAddressable(addressable *asset.Addressable) error {
	addressableExist, err := d.AddressableExists(addressable.StorageAddress)
	if err != nil {
		return err
	}
	if addressableExist {
		stmt := getStatementBuilder().
			Update("addressables").
			Set("storage_address", addressable.StorageAddress).
			Set("checksum", addressable.Checksum).
			Where(sq.Eq{"storage_address": addressable.StorageAddress})
		return d.exec(stmt)
	}
	return d.addAddressable(addressable, false)
}

func (d *DBAL) deleteAddressable(storageAddress string) error {
	stmt := getStatementBuilder().
		Delete("addressables").
		Where(sq.Eq{"storage_address": storageAddress})

	return d.exec(stmt)
}
