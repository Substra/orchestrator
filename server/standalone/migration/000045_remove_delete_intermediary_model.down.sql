ALTER TABLE compute_plans
ADD COLUMN IF NOT EXISTS delete_intermediary_models bool DEFAULT false;

ALTER TABLE compute_plans
ALTER COLUMN delete_intermediary_models SET NOT NULL; 

UPDATE events
SET asset = asset::jsonb || '{"delete_intermediary_models": false}'::jsonb
WHERE asset_kind = 'ASSET_COMPUTE_PLAN';
