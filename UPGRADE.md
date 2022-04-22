# Upgrade guide

This document details operations required to update from a version to another.

## 0.6.0 -> 0.6.1

The SQL migration `server/standalone/migration/000009_rename_compute_plan_key.up.sql` might fail if the database contains compute tasks which are not bound to any compute plan.
If that is the case, create a new compute plan in the DB for each orphan compute task. This is a very unlikely scenario which shouldn't happen in practice.

## 0.5.0 -> 0.6.0

### Manually run migration

Backup your data.
This operation will require a downtime.

1. Edit orchestrator-server deployment to set 0 replicas
1. Manually connect to postgres server
1. Execute [migration 000005](./server/standalone/migrations/000005_improve_compute_tasks_indexes.up.sql)
This may take a while depending on the volume of data (more than 10 minutes with 165k compute tasks).
1. Edit orchestrator-server deployment to set 1 replicas
