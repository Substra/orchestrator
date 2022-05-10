SELECT execute($$
    CREATE TABLE event_kinds (
        kind varchar(50) PRIMARY KEY
    );

    INSERT INTO event_kinds(kind)
    VALUES ('EVENT_ASSET_CREATED'),
           ('EVENT_ASSET_UPDATED'),
           ('EVENT_ASSET_DISABLED');

    ALTER TABLE events
    ADD COLUMN asset_kind varchar(50) REFERENCES asset_kinds (kind),
    ADD COLUMN event_kind varchar(50) REFERENCES event_kinds (kind),
    ADD COLUMN timestamp  timestamptz,
    ADD COLUMN metadata   jsonb;

    UPDATE events
    SET asset_kind = event ->> 'assetKind',
        event_kind = event ->> 'eventKind',
        timestamp  = (event ->> 'timestamp')::timestamptz,
        metadata   = COALESCE(event -> 'metadata', '{}'::jsonb);

    ALTER TABLE events
    ALTER COLUMN asset_kind SET NOT NULL,
    ALTER COLUMN event_kind SET NOT NULL,
    ALTER COLUMN timestamp SET NOT NULL,
    ALTER COLUMN metadata SET NOT NULL,
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;

    ALTER INDEX ix_event_asset RENAME TO ix_events_asset_key;

    DROP INDEX ix_event_creation;
    CREATE INDEX ix_events_timestamp ON events(timestamp);

    ALTER TABLE events
    DROP COLUMN event;
$$) WHERE column_exists('public', 'events', 'event');
