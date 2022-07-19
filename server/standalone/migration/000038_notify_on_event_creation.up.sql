CREATE OR REPLACE FUNCTION notify_event()
    RETURNS trigger AS
$$
DECLARE
BEGIN
    PERFORM pg_notify('events', JSON_BUILD_OBJECT(
            'event_position', NEW.position,
            'channel', NEW.channel
        )::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS notify_event ON events;

CREATE TRIGGER notify_event
    AFTER INSERT
    ON events
    FOR EACH ROW
EXECUTE PROCEDURE notify_event();
