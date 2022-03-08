# Naming conventions

## Protobuf procedures

Asset retrieval and querying methods should follow these patterns:

- GetXXX should take a key and return a single entity
- GetXXXs should take a slice of keys and return a slice of entities
- GetAllXXXs should take no arguments and return a slice of entities
- QueryXXX should take pagination param and optionally a filter, and return a response with entities (slice) + pagination

This should be the case in both proto and services.

## Protobuf inputs & outputs

Some RPC methods receive specific (i.e. not an asset) input and output.
In those cases, the following convention should be used: the input should be the name of the function suffixed by `Param`, the output suffixed by `Response`.

Example: the function named `QueryComputeTask` takes a `QueryComputeTaskParam` and returns a `QueryComputeTaskResponse`.

## Feature flags

Feature flags passed through environment variables should follow the pattern `FEATURE_ENABLED` and accept a boolean value.

Example: `METRICS_ENABLED`.

## Database

### Tables

Table names are snake_cased and plural

### Columns

- Surrogate keys are named `key` (and not `id`, see [related discussions](https://github.com/owkin/orchestrator/issues/538) on this topic).

### Indexes

Indexes should be prefixed with `ix_` and follow the rule of `ix_<table>_<colums>` where there can be several `<columns>` separated by underscores (`_`).

Example:

```sql
CREATE INDEX IF NOT EXISTS ix_compute_tasks_compute_plan_id_status ON compute_tasks (compute_plan_id, status);
```
