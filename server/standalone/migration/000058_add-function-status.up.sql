SELECT execute($$

    CREATE TABLE function_statuses (
        status VARCHAR(100) PRIMARY KEY
    );

    INSERT INTO function_statuses(status)
    VALUES  ('FUNCTION_STATUS_UNKNOWN'),
            ('FUNCTION_STATUS_WAITING'),
            ('FUNCTION_STATUS_BUILDING'),
            ('FUNCTION_STATUS_READY'),
            ('FUNCTION_STATUS_CANCELED'),
            ('FUNCTION_STATUS_FAILED');


    ALTER TABLE functions
    ADD COLUMN status VARCHAR(100) DEFAULT 'FUNCTION_STATUS_UNKNOWN'
    CONSTRAINT function_status_fkey REFERENCES function_statuses (status);
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
    SET asset = jsonb_set(asset, '{status}', to_jsonb('FUNCTION_STATUS_UNKNOWN'::text))
    WHERE asset_kind = 'ASSET_FUNCTION';

    CREATE INDEX ix_compute_tasks_function_key_status ON compute_tasks (function_key, status);
    
$$) WHERE NOT column_exists('public', 'functions', 'status');
