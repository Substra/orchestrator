ALTER TABLE models DROP COLUMN category CASCADE;
DROP TABLE model_categories;

UPDATE events
SET asset = asset #- '{category}'
WHERE asset_kind = 'ASSET_MODEL';
