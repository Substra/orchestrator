# Compute Task

A compute task is a generic structure representing a step in a [compute plan](./computeplan.md).
There are 4 different kind of tasks:

- Train
- Test
- Aggregate
- Composite

While the `ComputeTask` structure holds common fields, type-specific fields are held by dedicated substructures.

Train, Aggregate and Composite tasks can produce one or more [models](./model.md) as output.
Those models will have different [permissions](./permissions.md), which will be determined on task creation.

## Compatibility

Since a compute task will receive models from their parents,
there are some restrictions on which parents are allowed for each task.

Here is the expected cardinality for each task category:

**Note**:

- the asterisk denotes an exclusive link, ie a *Test* task can only have **one** parent at most
- parenthesis denotes optional dependencies
- no parents may be a valid input

| children ↓ / parent → | Train | Test | Aggregate | Composite |
|-----------------------|-------|------|-----------|-----------|
| Train                 | n     | 0    | n         | 0         |
| Test                  | 1*    | 0    | 1*        | 1*        |
| Aggregate             | n     | 0    | n         | n         |
| Composite             | 0     | 0    | (1)       | 1         |

## Rank

A task is executed as part of a [compute plan](./computeplan.md).
Inside the graph of tasks, each task has a rank depending on its depth in the graph.

General rules are:

- A task with no parents has a rank of `0`
- A task with parents has a rank of `max(parentRanks) + 1`

However, for **Test** compute tasks, the rank is set to the one of tested parent.
eg: if a test has an aggregate parent with rank 2, the test will also have a rank 2.

Since parents are set during task definition, the rank is an immutable property.

## Status

A task can have several status (see *States* below for available transitions):

- WAITING: new task waiting for its parents to be DONE. In this state the task cannot be processed yet.
- TODO: all dependencies are built (all parents DONE) so the task can be picked up by a worker and processed.
- DOING: the task is being processed by a worker.
- DONE: task has been successfully completed.
- FAILED: task execution has failed.
- CANCELED: task execution has been interrupted or stopped before completion.

## State

A compute task will go through different state during a compute plan execution.
This is an overview of a task's lifecycle:

![](./schemas/computetask.state.svg)

A task can be created in TODO or WAITING state depending on its parents.

During the ComputePlan execution, as tasks are DONE, their statuses will be reflected to their children.
If all the parents of a child task are DONE, this task enters TODO state.

A task may produces one or more [models](./model.md), they can only be registered when the task in in DOING.
This is to ensure that when a task starts (switch to DOING), all its inputs are available.

### Status change

A status change is a reaction to an action.
Task actions should match the following restrictions:

| action ↓ / sender → | Owner | Worker | Other |
|---------------------|-------|--------|-------|
| DOING               | n     | y      | n     |
| CANCELED            | y     | n      | n     |
| FAILED              | n     | y      | n     |

Basically:

- only the owner can cancel a task
- only the worker can act on a task processing (DOING/DONE/FAILED)

*DONE* status is a special case since the transition is internal:
only registration of all outputs can trigger the transition to DONE.
ie: when registering the last model of a train task, the task will switch to DONE.
The same goes for test tasks once their performance are registered.
