SELECT execute($$

    ALTER TABLE compute_tasks DROP CONSTRAINT fk_compute_plan;
    ALTER TABLE compute_tasks RENAME COLUMN compute_plan_id TO compute_plan_key;
    
    ALTER INDEX ix_compute_tasks_compute_plan_id RENAME TO ix_compute_tasks_compute_plan_key;

$$) WHERE NOT column_exists('public', 'models', 'compute_plan_key');
