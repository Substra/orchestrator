CREATE SEQUENCE IF NOT EXISTS seq_events_position AS bigint MINVALUE 1;

SELECT SETVAL('seq_events_position', MAX(position))
FROM events;

ALTER TABLE events
ALTER COLUMN position SET DEFAULT NEXTVAL('seq_events_position');

ALTER SEQUENCE seq_events_position OWNED BY events.position;
