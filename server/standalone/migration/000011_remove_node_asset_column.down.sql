SELECT execute($$
    DROP INDEX ix_nodes_creation_date;

    ALTER TABLE nodes
    ADD COLUMN asset JSONB;

    UPDATE nodes
    SET asset = json_build_object(
        'id', id,
        'creationDate', to_rfc_3339(creation_date)
    );

    CREATE INDEX ix_nodes_creation ON nodes ((asset->>'creationDate'));

    ALTER TABLE nodes
    DROP COLUMN creation_date;
$$) WHERE NOT column_exists('public', 'nodes', 'asset');
