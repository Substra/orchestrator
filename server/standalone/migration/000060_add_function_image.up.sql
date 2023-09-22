SELECT execute($$
    ALTER TABLE functions
    ADD COLUMN image_address VARCHAR(200) NULL;

    ALTER TABLE functions
    RENAME COLUMN functionAddress to archive_address;

    DROP VIEW IF EXISTS expanded_functions;
    CREATE VIEW expanded_functions AS
        SELECT 	key,
                name,
                description             AS description_address,
                desc_add.checksum       AS description_checksum,
                archive_address,
                archive_add.checksum   AS archive_checksum,
                permissions,
                owner,
                creation_date,
                metadata,
                channel,
                status,
                image_address,
                image_add.checksum   AS image_checksum
        FROM functions
        JOIN addressables desc_add ON functions.description = desc_add.storage_address
        JOIN addressables archive_add ON functions.archive_address = archive_add.storage_address
        LEFT JOIN addressables image_add ON functions.image_address = image_add.storage_address;

    UPDATE events
    SET asset = jsonb_set(asset, '{archiveAddress}', asset->'functionAddress') - 'functionAddress'
    WHERE asset_kind = 'ASSET_FUNCTION' AND NOT(asset ? 'archiveAddress');

    UPDATE events
    SET asset = asset - imageAddress
    WHERE asset_kind = 'ASSET_FUNCTION' AND NOT(asset ? 'imageAddress');
    
$$) WHERE NOT column_exists('public', 'functions', 'image_address');
