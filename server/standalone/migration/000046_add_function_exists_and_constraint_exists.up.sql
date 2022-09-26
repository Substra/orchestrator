CREATE OR REPLACE FUNCTION function_exists(TEXT) RETURNS bool as $$
    SELECT EXISTS(SELECT * FROM pg_proc WHERE proname = $1);
$$ language sql STRICT;

CREATE OR REPLACE FUNCTION constraint_exists(TEXT, TEXT) RETURNS bool as $$
    SELECT EXISTS(SELECT * FROM information_schema.table_constraints WHERE table_name=$1 AND constraint_name=$2);
$$ language sql STRICT;
