SELECT execute($$
    CREATE TYPE function_status AS ENUM (
        'FUNCTION_STATUS_UNKNOWN',
        'FUNCTION_STATUS_WAITING',
        'FUNCTION_STATUS_BUILDING',
        'FUNCTION_STATUS_READY',
        'FUNCTION_STATUS_CANCELED',
        'FUNCTION_STATUS_FAILED'
    );

    ALTER TABLE functions
    ADD COLUMN status function_status DEFAULT 'FUNCTION_STATUS_UNKNOWN';
    ALTER TABLE functions
    ALTER COLUMN status DROP DEFAULT;

    DROP VIEW IF EXISTS expanded_functions;
    CREATE VIEW expanded_functions AS
        SELECT 	key,
                name,
                description             AS description_address,
                desc_add.checksum       AS description_checksum,
                functionAddress         AS function_address,
                function_add.checksum   AS function_checksum,
                permissions,
                owner,
                creation_date,
                metadata,
                channel,
                status
        FROM functions
        JOIN addressables desc_add ON functions.description = desc_add.storage_address
        JOIN addressables function_add ON functions.functionAddress = function_add.storage_address;

    UPDATE events e
    SET asset = jsonb_set(asset, '{status}', to_jsonb('FUNCTION_STATUS_UNKNOWN'::function_status))
    WHERE asset_kind = 'ASSET_FUNCTION';

    CREATE INDEX ix_compute_tasks_function_key_status ON compute_tasks (function_key, status);
    
$$) WHERE NOT column_exists('public', 'functions', 'status');
