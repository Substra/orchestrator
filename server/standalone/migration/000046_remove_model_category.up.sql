SELECT execute($$
    ALTER TABLE models DROP COLUMN category CASCADE;
    DROP TABLE model_categories;

    CREATE VIEW expanded_models AS
    SELECT key,
           compute_task_key,
           address,
           a.checksum AS checksum,
           permissions,
           owner,
           channel,
           creation_date
    FROM models
    LEFT JOIN addressables a ON models.address = a.storage_address;

    UPDATE events
    SET asset = asset #- '{category}'
    WHERE asset_kind = 'ASSET_MODEL';
$$) WHERE table_exists('public', 'model_categories');
