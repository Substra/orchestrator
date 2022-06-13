SELECT execute($$

    ALTER TABLE nodes
    RENAME TO organizations;
    ALTER INDEX ix_nodes_creation_date RENAME TO ix_organizations_creation_date;

    INSERT INTO asset_kinds(kind)
    VALUES ('ASSET_ORGANIZATION');

    UPDATE events e
    SET asset_kind = 'ASSET_ORGANIZATION'
    WHERE e.asset_kind = 'ASSET_NODE';

    DELETE FROM asset_kinds
    WHERE kind = 'ASSET_NODE' ;

$$) WHERE not table_exists('public', 'organizations');
