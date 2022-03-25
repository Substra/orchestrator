-- Convert a timestamp with TZ to its RFC 3339 string representation.
-- The approach followed is similar to:
-- https://github.com/protocolbuffers/protobuf-go/blob/3992ea83a23c00882339f33511074d251e19822c/encoding/protojson/well_known_types.go#L782.
-- Contrary to the above, nanoseconds are never output, as PostgreSQL timestamps only have
-- microsecond resolution.
CREATE OR REPLACE FUNCTION to_rfc_3339(ts timestamptz) RETURNS text AS
$$
DECLARE
    s text;
BEGIN
    s := TO_CHAR(ts AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US');
    s := RTRIM(s, '000');
    s := RTRIM(s, '.000');
    RETURN s || 'Z';
END;
$$ LANGUAGE plpgsql;
