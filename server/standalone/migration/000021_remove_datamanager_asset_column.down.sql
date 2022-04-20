SELECT execute($$
    ALTER TABLE datamanagers
    ADD COLUMN asset jsonb;

    UPDATE datamanagers d
    SET asset = JSONB_BUILD_OBJECT(
        'key', d.key,
        'name', d.name,
        'owner', d.owner,
        'permissions', d.permissions,
        'description', build_addressable_jsonb(e.description_checksum, e.description_address),
        'opener', build_addressable_jsonb(e.opener_checksum, e.opener_address),
        'type', d.type,
        'creationDate', to_rfc_3339(d.creation_date),
        'logsPermission', d.logs_permission,
        'metadata', d.metadata
    )
    FROM expanded_datamanagers e
    WHERE d.key = e.key;

    DROP VIEW expanded_datamanagers;

    DROP INDEX ix_datamanagers_creation_date;
    CREATE INDEX ix_datamanagers_creation ON datamanagers ((asset->>'creationDate'));

    ALTER TABLE datamanagers
    RENAME COLUMN key TO id;

    ALTER TABLE datamanagers
    DROP CONSTRAINT datamanagers_description_fkey,
    DROP CONSTRAINT datamanagers_opener_fkey;

    DELETE FROM addressables
    WHERE storage_address IN (SELECT description FROM datamanagers);

    DELETE FROM addressables
    WHERE storage_address IN (SELECT opener FROM datamanagers);

    ALTER TABLE datamanagers
    DROP CONSTRAINT datamanagers_owner_channel_fkey,
    DROP COLUMN name,
    DROP COLUMN owner,
    DROP COLUMN permissions,
    DROP COLUMN description,
    DROP COLUMN opener,
    DROP COLUMN type,
    DROP COLUMN creation_date,
    DROP COLUMN logs_permission,
    DROP COLUMN metadata;
$$) WHERE column_exists('public', 'datamanagers', 'metadata');
