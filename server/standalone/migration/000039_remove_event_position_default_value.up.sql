ALTER TABLE events
ALTER COLUMN position DROP DEFAULT;

DROP SEQUENCE IF EXISTS seq_events_position;