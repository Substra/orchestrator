SELECT execute($$

    CREATE TABLE metrics (
        id UUID PRIMARY KEY,
        channel varchar(100) NOT NULL,
        asset JSONB NOT NULL
    );

    CREATE INDEX ix_metrics_creation ON metrics ((asset->>'creationDate'));

    INSERT INTO metrics(id, channel, asset)
    SELECT
        algos.key,
        algos.channel,
        JSONB_BUILD_OBJECT(
            'key', algos.key,
            'name', algos.name,
            'owner', algos.owner,
            'address', build_addressable_jsonb(a1.checksum, a1.storage_address),
            'description', build_addressable_jsonb(a2.checksum, a2.storage_address),
            'permissions', algos.permissions,
            'creationDate', to_rfc_3339(algos.creation_date),
            'metadata', algos.metadata
        )
    FROM algos
    JOIN addressables a1 ON a1.storage_address = algos.algorithm
    JOIN addressables a2 ON a2.storage_address = algos.description
    WHERE algos.category = 'ALGO_METRIC';

    ALTER TABLE performances DROP CONSTRAINT performances_metric_id_fkey;

    ALTER TABLE performances ADD CONSTRAINT performances_metric_id_fkey
    FOREIGN KEY (metric_id) REFERENCES metrics (id);

    DELETE FROM algos WHERE category = 'ALGO_METRIC';

    DELETE FROM algo_categories WHERE category = 'ALGO_METRIC';

    DELETE FROM addressables a
    WHERE EXISTS (
        SELECT 1
        FROM metrics m
        WHERE a.storage_address = m.asset -> 'address' ->> 'storageAddress'
        OR a.storage_address = m.asset -> 'description' ->> 'storageAddress'
    );

$$) WHERE NOT table_exists('public', 'metrics');
