SELECT execute($$
    DROP VIEW expanded_models;

    ALTER TABLE models
    ADD COLUMN asset jsonb;

    UPDATE models
    SET asset = JSONB_BUILD_OBJECT(
        'key', key,
        'category', category,
        'computeTaskKey', compute_task_key,
        'permissions', permissions,
        'owner', owner,
        'creationDate', to_rfc_3339(creation_date)
    );

    UPDATE models
    SET asset = asset || JSONB_BUILD_OBJECT(
        'address', build_addressable_jsonb(a.checksum, a.storage_address)
    )
    FROM addressables a
    WHERE models.address = a.storage_address;

    DROP INDEX ix_models_category;
    DROP INDEX ix_models_creation_date;

    CREATE INDEX ix_models_category ON models USING HASH ((asset->>'category'));
    CREATE INDEX ix_models_creation ON models ((asset->>'creationDate'));

    ALTER INDEX ix_models_compute_task_key RENAME TO ix_models_compute_task_id;

    ALTER TABLE models
    RENAME CONSTRAINT models_compute_task_key_fkey TO models_compute_task_id_fkey;

    ALTER TABLE models
    RENAME COLUMN compute_task_key TO compute_task_id;

    ALTER TABLE models
    RENAME COLUMN key TO id;

    ALTER TABLE models
    DROP CONSTRAINT models_address_fkey;

    DELETE FROM addressables
    WHERE storage_address IN (SELECT address FROM models);

    ALTER TABLE models
    DROP CONSTRAINT models_owner_channel_fkey,
    DROP COLUMN category,
    DROP COLUMN address,
    DROP COLUMN permissions,
    DROP COLUMN owner,
    DROP COLUMN creation_date;

    DROP TABLE model_categories;
$$) WHERE table_exists('public', 'model_categories');
