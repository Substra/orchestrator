# Upgrade guide

This document details operations required to update from a version to another.

## 0.5.0 -> Unreleased

### Manually run migration

Backup your data.
This operation will require a downtime.

1. Edit orchestrator-server deployment to set 0 replicas
1. Manually connect to postgres server
1. Execute [migration 000005](./server/standalone/migrations/000005_improve_compute_tasks_indexes.up.sql)
This may take a while depending on the volume of data (more than 10 minutes with 165k compute tasks).
1. Edit orchestrator-server deployment to set 1 replicas
