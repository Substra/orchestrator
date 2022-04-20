SELECT execute($$
    ALTER TABLE datamanagers
    ADD COLUMN name            varchar(100),
    ADD COLUMN owner           varchar(100),
    ADD COLUMN permissions     jsonb,
    ADD COLUMN description     varchar(200) REFERENCES addressables (storage_address),
    ADD COLUMN opener          varchar(200) REFERENCES addressables (storage_address),
    ADD COLUMN type            varchar(30),
    ADD COLUMN creation_date   timestamptz,
    ADD COLUMN logs_permission jsonb,
    ADD COLUMN metadata        jsonb,
    ADD CONSTRAINT datamanagers_owner_channel_fkey FOREIGN KEY (owner, channel) REFERENCES nodes (id, channel);

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'description' ->> 'storageAddress',
           asset -> 'description' ->> 'checksum'
    FROM datamanagers;

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'opener' ->> 'storageAddress',
           asset -> 'opener' ->> 'checksum'
    FROM datamanagers;

    UPDATE datamanagers
    SET name            = asset ->> 'name',
        owner           = asset ->> 'owner',
        permissions     = asset -> 'permissions',
        description     = asset -> 'description' ->> 'storageAddress',
        opener          = asset -> 'opener' ->> 'storageAddress',
        type            = asset ->> 'type',
        creation_date   = (asset ->> 'creationDate')::timestamptz,
        logs_permission = COALESCE(asset -> 'logsPermission', '{}'::jsonb),
        metadata        = COALESCE(asset -> 'metadata', '{}'::jsonb);

    ALTER TABLE datamanagers
    ALTER COLUMN name SET NOT NULL,
    ALTER COLUMN owner SET NOT NULL,
    ALTER COLUMN permissions SET NOT NULL,
    ALTER COLUMN description SET NOT NULL,
    ALTER COLUMN opener SET NOT NULL,
    ALTER COLUMN type SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL,
    ALTER COLUMN logs_permission SET NOT NULL,
    ALTER COLUMN metadata SET NOT NULL,
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;

    ALTER TABLE datamanagers
    RENAME COLUMN id TO key;

    DROP INDEX ix_datamanagers_creation;
    CREATE INDEX ix_datamanagers_creation_date ON datamanagers (creation_date);

    ALTER TABLE datamanagers
    DROP COLUMN asset;

    CREATE VIEW expanded_datamanagers AS
    SELECT key,
           name,
           owner,
           channel,
           permissions,
           description         AS description_address,
           desc_add.checksum   AS description_checksum,
           opener              AS opener_address,
           opener_add.checksum AS opener_checksum,
           type,
           creation_date,
           logs_permission,
           metadata
    FROM datamanagers
    JOIN addressables desc_add ON datamanagers.description = desc_add.storage_address
    JOIN addressables opener_add ON datamanagers.opener = opener_add.storage_address;
$$) WHERE NOT view_exists('public', 'expanded_datamanagers');
