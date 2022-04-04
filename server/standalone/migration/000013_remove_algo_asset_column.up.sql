SELECT execute($$
    CREATE TABLE addressables (
        storage_address varchar(200) PRIMARY KEY,
        checksum        varchar(64)
    );

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'description' ->> 'storageAddress',
           asset -> 'description' ->> 'checksum'
    FROM algos;

    INSERT INTO addressables(storage_address, checksum)
    SELECT asset -> 'algorithm' ->> 'storageAddress',
           asset -> 'algorithm' ->> 'checksum'
    FROM algos;

    CREATE TABLE algo_categories (
        category varchar(50) PRIMARY KEY
    );

    INSERT INTO algo_categories(category)
    VALUES ('ALGO_SIMPLE'),
           ('ALGO_AGGREGATE'),
           ('ALGO_COMPOSITE');

    ALTER TABLE algos
    ADD COLUMN name          varchar(100),
    ADD COLUMN category      varchar(50) REFERENCES algo_categories (category),
    ADD COLUMN description   varchar(200) REFERENCES addressables (storage_address),
    ADD COLUMN algorithm     varchar(200) REFERENCES addressables (storage_address),
    ADD COLUMN permissions   jsonb,
    ADD COLUMN owner         varchar(100),
    ADD COLUMN creation_date timestamptz,
    ADD COLUMN metadata      jsonb,
    ADD CONSTRAINT algos_owner_channel_fkey FOREIGN KEY (owner, channel) REFERENCES nodes (id, channel);

    UPDATE algos
    SET name = asset ->> 'name',
        category = asset ->> 'category',
        description = asset -> 'description' ->> 'storageAddress',
        algorithm = asset -> 'algorithm' ->> 'storageAddress',
        permissions = asset -> 'permissions',
        owner = asset ->> 'owner',
        creation_date = (asset ->> 'creationDate')::timestamptz,
        metadata = COALESCE(asset -> 'metadata', '{}'::jsonb);

    ALTER TABLE algos
    ALTER COLUMN name SET NOT NULL,
    ALTER COLUMN category SET NOT NULL,
    ALTER COLUMN description SET NOT NULL,
    ALTER COLUMN algorithm SET NOT NULL,
    ALTER COLUMN permissions SET NOT NULL,
    ALTER COLUMN owner SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL,
    ALTER COLUMN metadata SET NOT NULL,
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;

    DROP INDEX ix_algos_category;
    DROP INDEX ix_algos_creation;

    CREATE INDEX ix_algos_category ON algos (category);
    CREATE INDEX ix_algos_creation_date ON algos (creation_date);

    ALTER TABLE algos
    RENAME COLUMN id to key;

    ALTER TABLE algos
    DROP COLUMN asset;

    CREATE VIEW expanded_algos AS
    SELECT key,
           name,
           category,
           description       AS description_address,
           desc_add.checksum AS description_checksum,
           algorithm         AS algorithm_address,
           algo_add.checksum AS algorithm_checksum,
           permissions,
           owner,
           creation_date,
           metadata,
           channel
    FROM algos
    JOIN addressables desc_add ON algos.description = desc_add.storage_address
    JOIN addressables algo_add ON algos.algorithm = algo_add.storage_address;
$$) WHERE NOT view_exists('public', 'expanded_algos');
