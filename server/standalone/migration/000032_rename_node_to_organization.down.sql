SELECT execute($$

    ALTER TABLE organizations
    RENAME TO nodes;
    ALTER INDEX ix_organizations_creation_date RENAME TO ix_nodes_creation_date;

    INSERT INTO asset_kinds(kind)
    VALUES ('ASSET_NODE');

    UPDATE events e
    SET asset_kind = 'ASSET_NODE'
    WHERE e.asset_kind = 'ASSET_ORGANIZATION';

    DELETE FROM asset_kinds
    WHERE kind = 'ASSET_ORGANIZATION';

$$) WHERE not table_exists('public', 'nodes');
