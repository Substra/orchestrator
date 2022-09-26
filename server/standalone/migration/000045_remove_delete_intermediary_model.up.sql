ALTER TABLE compute_plans
DROP COLUMN IF EXISTS delete_intermediary_models;

UPDATE events
SET asset = asset::jsonb - 'delete_intermediary_models'
WHERE asset_kind = 'ASSET_COMPUTE_PLAN';
