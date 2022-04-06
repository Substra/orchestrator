SELECT execute($$
    DROP VIEW expanded_compute_plans;

    ALTER TABLE compute_plans
    ADD COLUMN asset JSONB;

    UPDATE compute_plans
    SET asset = JSONB_BUILD_OBJECT(
        'key', key,
        'owner', owner,
        'deleteIntermediaryModels', delete_intermediary_models,
        'creationDate', to_rfc_3339(creation_date),
        'tag', tag,
        'metadata', metadata
    );

    ALTER TABLE compute_plans
    ALTER COLUMN asset SET NOT NULL;

    ALTER TABLE compute_plans
    RENAME COLUMN key to id;

    DROP INDEX ix_compute_plans_creation_date;
    DROP INDEX ix_compute_plans_owner;

    CREATE INDEX ix_compute_plans_creation ON compute_plans ((asset->>'creationDate'));
    CREATE INDEX ix_compute_plans_owner ON compute_plans ((asset->>'owner'));

    ALTER TABLE compute_plans
    DROP CONSTRAINT compute_plans_owner_channel_fkey,
    DROP COLUMN owner,
    DROP COLUMN delete_intermediary_models,
    DROP COLUMN creation_date,
    DROP COLUMN tag,
    DROP COLUMN metadata;
$$) WHERE column_exists('public', 'compute_plans', 'metadata');
