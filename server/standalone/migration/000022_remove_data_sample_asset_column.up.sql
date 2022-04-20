SELECT execute($$
    ALTER TABLE datasamples
    ADD COLUMN owner         varchar(100),
    ADD COLUMN test_only     bool,
    ADD COLUMN checksum      varchar(64),
    ADD COLUMN creation_date timestamptz,
    ADD CONSTRAINT datasamples_owner_channel_fkey FOREIGN KEY (owner, channel) REFERENCES nodes (id, channel);

    UPDATE datasamples
    SET owner         = asset ->> 'owner',
        test_only     = COALESCE((asset ->> 'testOnly')::bool, FALSE),
        checksum      = asset ->> 'checksum',
        creation_date = (asset ->> 'creationDate')::timestamptz;

    ALTER TABLE datasamples
    ALTER COLUMN owner SET NOT NULL,
    ALTER COLUMN test_only SET NOT NULL,
    ALTER COLUMN checksum SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL;

    ALTER TABLE datasamples
    RENAME COLUMN id to key;

    CREATE TABLE datasample_datamanagers (
        datasample_key  uuid REFERENCES datasamples (key),
        datamanager_key uuid REFERENCES datamanagers (key),
        PRIMARY KEY (datasample_key, datamanager_key)
    );

    INSERT INTO datasample_datamanagers (datasample_key, datamanager_key)
    SELECT key, JSONB_ARRAY_ELEMENTS_TEXT(asset -> 'dataManagerKeys')::uuid
    FROM datasamples;

    DROP INDEX ix_datasamples_creation;
    CREATE INDEX ix_datasamples_creation_date ON datasamples (creation_date);

    ALTER TABLE datasamples
    DROP COLUMN asset;

    CREATE VIEW expanded_datasamples AS
    SELECT key,
           owner,
           channel,
           test_only,
           checksum,
           creation_date,
           JSONB_AGG(dd.datamanager_key) AS datamanager_keys
    FROM datasamples
    LEFT JOIN datasample_datamanagers dd ON datasamples.key = dd.datasample_key
    GROUP BY datasamples.key;
$$) WHERE NOT view_exists('public', 'expanded_datasamples');
