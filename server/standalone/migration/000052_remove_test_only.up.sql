ALTER TABLE datasamples
DROP COLUMN IF EXISTS test_only;

UPDATE events e
SET asset = JSONB_BUILD_OBJECT(
        'key', ds.key,
        'dataManagerKeys', ds.datamanager_keys,
        'owner', ds.owner,
        'checksum', ds.checksum,
        'creationDate', to_rfc_3339(ds.creation_date)
    )
FROM expanded_datasamples ds
WHERE e.asset_kind = 'ASSET_DATA_SAMPLE';