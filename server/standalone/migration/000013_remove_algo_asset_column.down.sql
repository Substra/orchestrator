SELECT execute($$
    DROP VIEW expanded_algos;

    ALTER TABLE algos
    ADD COLUMN asset JSONB;

    UPDATE algos
    SET asset = build_algo_jsonb(
        key,
        name,
        category,
        desc_add.checksum,
        description,
        algo_add.checksum,
        algorithm,
        permissions,
        owner,
        creation_date,
        metadata
        )
    FROM addressables desc_add, addressables algo_add
    WHERE algos.description = desc_add.storage_address AND algos.algorithm = algo_add.storage_address;

    ALTER TABLE algos
    ALTER COLUMN asset SET NOT NULL;

    ALTER TABLE algos
    RENAME COLUMN key to id;

    DROP INDEX ix_algos_category;
    DROP INDEX ix_algos_creation_date;

    CREATE INDEX ix_algos_category ON algos USING HASH ((asset->>'category'));
    CREATE INDEX ix_algos_creation ON algos ((asset->>'creationDate'));

    ALTER TABLE algos
    DROP CONSTRAINT algos_owner_channel_fkey,
    DROP COLUMN name,
    DROP COLUMN category,
    DROP COLUMN description,
    DROP COLUMN algorithm,
    DROP COLUMN permissions,
    DROP COLUMN owner,
    DROP COLUMN creation_date,
    DROP COLUMN metadata;

    DROP TABLE algo_categories;
    DROP TABLE addressables;
$$) WHERE table_exists('public', 'addressables');
