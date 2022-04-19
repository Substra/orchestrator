SELECT execute($$
    CREATE TABLE model_categories (
        category varchar(50) PRIMARY KEY
    );

    INSERT INTO model_categories(category)
    VALUES ('MODEL_SIMPLE'),
           ('MODEL_HEAD');

    ALTER TABLE models
    ADD COLUMN category      varchar(50) REFERENCES model_categories (category),
    ADD COLUMN address       varchar(200) REFERENCES addressables (storage_address),
    ADD COLUMN permissions   jsonb,
    ADD COLUMN owner         varchar(100),
    ADD COLUMN creation_date timestamptz,
    ADD CONSTRAINT models_owner_channel_fkey FOREIGN KEY (owner, channel) REFERENCES nodes (id, channel);

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'address' ->> 'storageAddress',
           asset -> 'address' ->> 'checksum'
    FROM models
    WHERE (asset -> 'address') IS NOT NULL;

    UPDATE models
    SET category = asset ->> 'category',
        address = asset -> 'address' ->> 'storageAddress',
        permissions = asset -> 'permissions',
        owner = asset ->> 'owner',
        creation_date = (asset ->> 'creationDate')::timestamptz;

    ALTER TABLE models
    ALTER COLUMN category SET NOT NULL,
    ALTER COLUMN permissions SET NOT NULL,
    ALTER COLUMN owner SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL;

    ALTER TABLE models
    RENAME COLUMN id TO key;

    ALTER TABLE models
    RENAME COLUMN compute_task_id TO compute_task_key;

    ALTER TABLE models
    RENAME CONSTRAINT models_compute_task_id_fkey TO models_compute_task_key_fkey;

    ALTER INDEX ix_models_compute_task_id RENAME TO ix_models_compute_task_key;

    DROP INDEX ix_models_category;
    DROP INDEX ix_models_creation;

    CREATE INDEX ix_models_category ON models (category);
    CREATE INDEX ix_models_creation_date ON models (creation_date);

    ALTER TABLE models
    DROP COLUMN asset;

    CREATE VIEW expanded_models AS
    SELECT key,
           compute_task_key,
           category,
           address,
           a.checksum AS checksum,
           permissions,
           owner,
           channel,
           creation_date
    FROM models
    LEFT JOIN addressables a ON models.address = a.storage_address;
$$) WHERE NOT view_exists('public', 'expanded_models');
