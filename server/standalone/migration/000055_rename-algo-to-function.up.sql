ALTER TABLE algos
RENAME COLUMN algorithm TO functionAdress;

ALTER TABLE algos
RENAME TO functions;

ALTER TABLE compute_tasks
RENAME COLUMN algo_key TO function_key;

ALTER TABLE algo_outputs
RENAME COLUMN algo_key TO function_key;

ALTER TABLE algo_outputs
RENAME TO function_outputs;

ALTER TABLE algo_inputs
RENAME COLUMN algo_key TO function_key;

ALTER TABLE algo_inputs
RENAME TO function_inputs;

DROP VIEW IF EXISTS expanded_algos;
CREATE VIEW expanded_functions AS
SELECT 	key,
        name,
        description             AS description_address,
        desc_add.checksum       AS description_checksum,
        functionAdress          AS function_address,
        function_add.checksum   AS function_checksum,
	    permissions,
        owner,
        creation_date,
        metadata,
        channel
FROM functions
JOIN addressables desc_add ON functions.description = desc_add.storage_address
JOIN addressables function_add ON functions.functionAdress = function_add.storage_address;

INSERT INTO asset_kinds(kind)
VALUES ('ASSET_FUNCTION');

UPDATE events e
SET asset_kind = 'ASSET_FUNCTION'
WHERE e.asset_kind = 'ASSET_ALGO';

DELETE FROM asset_kinds
WHERE kind = 'ASSET_ALGO';

UPDATE events
SET asset = asset - 'algo' || JSONB_BUILD_OBJECT('function', asset->'algo')
WHERE asset_kind = 'ASSET_COMPUTE_TASK';
