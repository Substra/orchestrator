ALTER TABLE datamanagers 
ADD COLUMN IF NOT EXISTS archived BOOLEAN DEFAULT False;

DROP VIEW IF EXISTS expanded_datamanagers;

CREATE VIEW expanded_datamanagers AS
SELECT key,
        name,
        owner,
        channel,
        permissions,
        description         AS description_address,
        desc_add.checksum   AS description_checksum,
        opener              AS opener_address,
        opener_add.checksum AS opener_checksum,
        type,
        creation_date,
        logs_permission,
        metadata,
        archived
FROM datamanagers
JOIN addressables desc_add ON datamanagers.description = desc_add.storage_address
JOIN addressables opener_add ON datamanagers.opener = opener_add.storage_address;
