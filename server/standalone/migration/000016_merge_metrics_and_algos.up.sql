SELECT execute($$

    INSERT INTO algo_categories(category)
    VALUES ('ALGO_METRIC');

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'address' ->> 'storageAddress',
           asset -> 'address' ->> 'checksum'
    FROM metrics;

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'description' ->> 'storageAddress',
           asset -> 'description' ->> 'checksum'
    FROM metrics;

    INSERT INTO algos (key, channel, name, category, description, algorithm, permissions, owner, creation_date, metadata)
    SELECT
        id,
        channel,
        asset ->> 'name',
        'ALGO_METRIC',
        asset -> 'description' ->> 'storageAddress',
        asset -> 'address' ->> 'storageAddress',
        asset -> 'permissions',
        asset ->> 'owner',
        (asset ->> 'creationDate')::timestamptz,
        COALESCE(asset -> 'metadata', '{}'::jsonb)
    FROM metrics;

    ALTER TABLE performances
    DROP CONSTRAINT performances_metric_id_fkey;

    ALTER TABLE performances
    ADD CONSTRAINT performances_metric_id_fkey FOREIGN KEY (metric_id) REFERENCES algos (key);

    DROP TABLE metrics;

$$) WHERE table_exists('public', 'metrics');
