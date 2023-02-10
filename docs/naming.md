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

- Surrogate keys are named `key` and not `id`.

### Indexes

Indexes should be prefixed with `ix_` and follow the rule of `ix_<table>_<columns>` where there can be several `<columns>` separated by underscores (`_`).

Example:

```sql
CREATE INDEX IF NOT EXISTS ix_compute_tasks_compute_plan_id_status ON compute_tasks (compute_plan_id, status);
```

### Check constraints

Check contraints should be prefixed with `ck_` and follow the rule of `ck_<table>_<columns>` where there can be several `<columns>` separated by underscores (`_`).

Example:

```sql
ALTER TABLE failure_reports ADD CONSTRAINT ck_error_type_log_address CHECK (
    (error_type = 'ERROR_TYPE_EXECUTION' AND logs_address IS NOT NULL) OR
    (error_type ! 'ERROR_TYPE_EXECUTION' AND logs_address IS NULL)
);
```

### Sequences

Sequences should be prefixed with `seq_` and follow the naming pattern `seq_<table>_<column>`.

### Views

Views that join together multiple assets by expanding relationships should be named `expanded_<table>`, where `<table>`
refers to the main asset.

Example: Prefer `expanded_functions` to `functions_with_addressables`.
