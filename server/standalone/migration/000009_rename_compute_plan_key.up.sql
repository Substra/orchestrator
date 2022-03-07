SELECT execute($$
    ALTER TABLE compute_tasks RENAME COLUMN compute_plan_key TO compute_plan_id;
    ALTER TABLE compute_tasks ADD CONSTRAINT fk_compute_plan FOREIGN KEY(compute_plan_id) REFERENCES compute_plans (id);

    ALTER INDEX ix_compute_tasks_compute_plan_key RENAME TO ix_compute_tasks_compute_plan_id;

$$) WHERE NOT column_exists('public', 'models', 'compute_plan_id');
