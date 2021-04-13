# Compute Task

A compute task is a generic structure representing a compute task in a compute plan.
There are 4 different kind of tasks:
- Train
- Test
- Aggregate
- Composite

While the `ComputeTask` structure holds common fields, type-specific fields are held by dedicated substructures.

## Compatibility

Since a compute task will receive models from their parents,
there are some restrictions on which parents are allowed for each task.

Here is the expected cardinality for each task category:

**Note**:
- the asterisk denotes an exclusive link, ie a *Train* task can only have **one** parent at most
- parenthesis denotes optional dependencies
- no parents is a valid input

| children ↓ / parent → | Train | Test | Aggregate | Composite |
|-----------------------|-------|------|-----------|-----------|
| Train                 | n     | 0    | n         | 0         |
| Test                  | 1*    | 0    | 0         | 1*        |
| Aggregate             | n     | 0    | n         | n         |
| Composite             | 0     | 0    | (1)       | 1         |
