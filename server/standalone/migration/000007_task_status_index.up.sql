CREATE INDEX IF NOT EXISTS ix_compute_tasks_compute_plan_key_status ON compute_tasks (compute_plan_key, status);
