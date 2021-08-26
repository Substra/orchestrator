DROP TABLE IF EXISTS nodes;
DROP INDEX IF EXISTS ix_nodes_creation;

DROP TABLE IF EXISTS objectives;
DROP INDEX IF EXISTS ix_objectives_creation;

DROP TABLE IF EXISTS datasamples;
DROP INDEX IF EXISTS ix_datasamples_creation;

DROP TABLE IF EXISTS algos;
DROP INDEX IF EXISTS ix_algos_category;
DROP INDEX IF EXISTS ix_algos_creation;

DROP TABLE IF EXISTS datamanagers;
DROP INDEX IF EXISTS ix_datamanagers_creation;

DROP TABLE IF EXISTS compute_tasks;
DROP INDEX IF EXISTS ix_compute_tasks_parents;
DROP INDEX IF EXISTS ix_compute_tasks_status;
DROP INDEX IF EXISTS ix_compute_tasks_category;
DROP INDEX IF EXISTS ix_compute_tasks_compute_plan_key;
DROP INDEX IF EXISTS ix_compute_tasks_worker;
DROP INDEX IF EXISTS ix_compute_tasks_test_objective_key;
DROP INDEX IF EXISTS ix_compute_tasks_creation;

DROP TABLE IF EXISTS models;
DROP INDEX IF EXISTS ix_models_task;
DROP INDEX IF EXISTS ix_models_category;
DROP INDEX IF EXISTS ix_models_creation;

DROP TABLE IF EXISTS compute_plans;
DROP INDEX IF EXISTS ix_compute_plans_creation;

DROP TABLE IF EXISTS performances;
DROP INDEX IF EXISTS ix_performances_compute_task_key;
DROP INDEX IF EXISTS ix_performances_creation;

DROP TABLE IF EXISTS events;
DROP INDEX IF EXISTS ix_event_asset;
DROP INDEX IF EXISTS ix_event_creation;
