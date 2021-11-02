DROP INDEX ix_compute_tasks_compute_plan_key;
DROP INDEX ix_compute_tasks_test_metric_key;
DROP INDEX ix_models_task;
DROP INDEX ix_performances_compute_task_key;

CREATE INDEX ix_compute_tasks_compute_plan_key ON compute_tasks USING HASH (((asset->>'computePlanKey')::uuid));
CREATE INDEX ix_compute_tasks_test_metric_key ON compute_tasks USING HASH (((asset->'test'->>'metricKey')::uuid));
CREATE INDEX ix_models_task ON models USING HASH (((asset->>'computeTaskKey')::uuid));
CREATE INDEX ix_performances_compute_task_key ON performances USING HASH (((asset->>'computeTaskKey')::uuid));
