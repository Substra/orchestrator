CREATE INDEX IF NOT EXISTS ix_compute_plans_owner ON compute_plans ((asset->>'owner'));
