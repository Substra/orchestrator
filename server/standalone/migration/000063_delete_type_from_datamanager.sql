SELECT execute($$
    ALTER TABLE datamanagers
    DROP COLUMN type;

    UPDATE events SET asset = asset - 'type'
    WHERE asset_kind = 'ASSET_DATA_MANAGER';

$$) WHERE column_exists('public', 'datamanagers', 'type');