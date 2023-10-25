SELECT execute($$
    CREATE TYPE failed_asset_kind AS ENUM (
        'FAILED_ASSET_UNKNOWN',
        'FAILED_ASSET_COMPUTE_TASK',
        'FAILED_ASSET_FUNCTION'
    );
    CREATE TABLE failed_asset_kinds (
        kind VARCHAR(100) PRIMARY KEY
    );

    INSERT INTO failed_asset_kinds(kind)
    VALUES  ('FAILED_ASSET_UNKNOWN'),
            ('FAILED_ASSET_COMPUTE_TASK'),
            ('FAILED_ASSET_FUNCTION');

    ALTER TABLE failure_reports
    RENAME COLUMN compute_task_key TO asset_key;
    ALTER TABLE failure_reports
    ADD COLUMN asset_type VARCHAR(100) DEFAULT 'FAILED_ASSET_COMPUTE_TASK'
    CONSTRAINT asset_kind_fkey REFERENCES failed_asset_kinds (kind);
    ALTER TABLE failure_reports
    ALTER COLUMN asset_type SET DEFAULT 'FAILED_ASSET_UNKNOWN';
    ALTER TABLE failure_reports
    DROP CONSTRAINT failure_reports_compute_task_key_fkey;

    DROP VIEW IF EXISTS expanded_failure_reports;
    CREATE VIEW expanded_failure_reports AS
    SELECT asset_key,
           asset_type,
           error_type,
           logs_address,
           a.checksum AS logs_checksum,
           creation_date,
           owner,
           channel
    FROM failure_reports
    LEFT JOIN addressables a ON failure_reports.logs_address = a.storage_address;

    UPDATE events
    SET asset = jsonb_set(asset, '{assetKey}', asset->'computeTaskKey') - 'computeTaskKey'
    WHERE asset_kind = 'ASSET_FAILURE_REPORT' AND NOT(asset ? 'assetKey');
    UPDATE events e
    SET asset = jsonb_set(asset, '{assetType}', to_jsonb('FAILED_ASSET_COMPUTE_TASK'::text))
    WHERE asset_kind = 'ASSET_FAILURE_REPORT';
    
$$) WHERE NOT column_exists('public', 'failure_reports', 'asset_type');
