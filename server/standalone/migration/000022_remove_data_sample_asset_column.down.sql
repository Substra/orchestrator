SELECT execute($$
    ALTER TABLE datasamples
    ADD COLUMN asset jsonb;

    UPDATE datasamples d
    SET asset = JSONB_BUILD_OBJECT(
        'key', d.key,
        'dataManagerKeys', e.datamanager_keys,
        'owner', d.owner,
        'testOnly', d.test_only,
        'checksum', d.checksum,
        'creationDate', to_rfc_3339(d.creation_date)
    )
    FROM expanded_datasamples e
    WHERE d.key = e.key;

    DROP VIEW expanded_datasamples;

    DROP TABLE datasample_datamanagers;

    DROP INDEX ix_datasamples_creation_date;
    CREATE INDEX ix_datasamples_creation ON datasamples ((asset->>'creationDate'));

    ALTER TABLE datasamples
    RENAME COLUMN key TO id;

    ALTER TABLE datasamples
    DROP CONSTRAINT datasamples_owner_channel_fkey,
    DROP COLUMN owner,
    DROP COLUMN test_only,
    DROP COLUMN checksum,
    DROP COLUMN creation_date;
$$) WHERE column_exists('public', 'datasamples', 'creation_date');
