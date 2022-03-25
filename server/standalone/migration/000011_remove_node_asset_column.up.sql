SELECT execute($$
    ALTER TABLE nodes
    ADD COLUMN creation_date timestamptz;

    UPDATE nodes
    SET creation_date = (asset ->> 'creationDate')::timestamptz;

    ALTER TABLE nodes
    ALTER COLUMN creation_date SET NOT NULL;

    DROP INDEX ix_nodes_creation;

    ALTER TABLE nodes
    DROP COLUMN asset;

    CREATE INDEX ix_nodes_creation_date ON nodes (creation_date);
$$) WHERE NOT column_exists('public', 'nodes', 'creation_date');
