/* This function is helpful to run idempotent SQL migrations */
/* See https://www.depesz.com/2008/06/18/conditional-ddl/ */

CREATE OR REPLACE FUNCTION execute(TEXT) RETURNS VOID AS $$
BEGIN EXECUTE $1; END;
$$ LANGUAGE plpgsql STRICT;

/* These helper functions are useful when writing idempotent SQL migrations */
/* See https://www.depesz.com/2008/06/18/conditional-ddl/ */

CREATE OR REPLACE FUNCTION schema_exists(TEXT) RETURNS bool as $$
    SELECT exists(SELECT 1 FROM information_schema.schemata WHERE schema_name = $1);
$$ language sql STRICT;

CREATE OR REPLACE FUNCTION table_exists(TEXT, TEXT) RETURNS bool as $$
    SELECT exists(SELECT 1 FROM information_schema.tables WHERE (table_schema, table_name, table_type) = ($1, $2, 'BASE TABLE'));
$$ language sql STRICT;

CREATE OR REPLACE FUNCTION view_exists(TEXT, TEXT) RETURNS bool as $$
    SELECT exists(SELECT 1 FROM information_schema.views WHERE (table_schema, table_name) = ($1, $2));
$$ language sql STRICT;

CREATE OR REPLACE FUNCTION column_exists(TEXT, TEXT, TEXT) RETURNS bool as $$
    SELECT exists(SELECT 1 FROM information_schema.columns WHERE (table_schema, table_name, column_name) = ($1, $2, $3));
$$ language sql STRICT;

CREATE OR REPLACE FUNCTION index_exists(TEXT) RETURNS bool as $$
    SELECT exists(SELECT 1 FROM pg_class t WHERE relkind = 'i' and relname = $1);
$$ language sql STRICT;
