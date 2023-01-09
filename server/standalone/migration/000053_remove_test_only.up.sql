SELECT execute($$
        DROP VIEW IF EXISTS expanded_datasamples;
        CREATE VIEW expanded_datasamples AS
        SELECT key,
                owner,
                channel,
                checksum,
                creation_date,
                JSONB_AGG(dd.datamanager_key) AS datamanager_keys
        FROM datasamples
        LEFT JOIN datasample_datamanagers dd ON datasamples.key = dd.datasample_key
        GROUP BY datasamples.key;

        ALTER TABLE datasamples
        DROP COLUMN IF EXISTS test_only;

        UPDATE events SET asset = asset - 'testOnly'
        WHERE asset_kind = 'ASSET_DATA_SAMPLE';
$$) WHERE column_exists('public', 'datasamples', 'test_only');
