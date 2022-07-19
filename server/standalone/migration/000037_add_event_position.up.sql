SELECT execute($$
    ALTER TABLE events
    ADD COLUMN position bigint UNIQUE;

    UPDATE events
    SET position = e.position
    FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY timestamp) AS position FROM events) AS e
    WHERE events.id = e.id;

    CREATE SEQUENCE seq_events_position AS bigint MINVALUE 1;
    SELECT SETVAL('seq_events_position', MAX(position))
    FROM events;

    ALTER TABLE events
    ALTER COLUMN position SET NOT NULL,
    ALTER COLUMN position SET DEFAULT NEXTVAL('seq_events_position');

    ALTER SEQUENCE seq_events_position OWNED BY events.position;
$$) WHERE NOT column_exists('public', 'events', 'position');
