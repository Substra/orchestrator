SELECT execute($$
    ALTER TABLE events
    ADD COLUMN event JSONB;

    UPDATE events
    SET event = JSONB_BUILD_OBJECT(
        'id', id,
        'assetKey', asset_key,
        'assetKind', asset_kind,
        'eventKind', event_kind,
        'timestamp', to_rfc_3339(timestamp),
        'metadata', metadata
    );

    ALTER TABLE events
    ALTER COLUMN event SET NOT NULL;

    ALTER INDEX ix_events_asset_key RENAME TO ix_event_asset;

    DROP INDEX ix_events_timestamp;
    CREATE INDEX ix_event_creation ON events ((event ->> 'timestamp'));

    ALTER TABLE events
    DROP COLUMN asset_kind,
    DROP COLUMN event_kind,
    DROP COLUMN timestamp,
    DROP COLUMN metadata;

    DROP TABLE event_kinds;
$$) WHERE table_exists('public', 'event_kinds');
