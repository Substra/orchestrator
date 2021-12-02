# Model

A model is the product of a [compute task](./computetask.md).
Not all compute task produces a model (Test task), and some may produce more than one model (Composite task).

A model inherit its permissions from its parent task.

## Registration

A model can only be registered on a task with `DOING` status.

Registering the last expected output model on a task will also trigger its transition to `DONE` status.

Only the task's worker can register models.

## Compatibility

A model has a category, and it can only be registered for a compatible task (y):

| model category ↓ / task category → | Train | Test | Aggregate | Composite |
|------------------------------------|-------|------|-----------|-----------|
| Simple                             | y     | n    | y         | y         |
| Head                               | n     | n    | n         | y         |

As stated before, a task may produce several models, but can only have one of each category.
i.e. composite task can only have one *head* AND one *simple*.

## Disabling intermediary models

As part of compute plan execution, some models may be *disabled* (depending on the `delete_intermediary_models` flag of the compute plan).
This only occurs for **intermediary** models.

An intermediary model is a model produced by a task which:
- is not a leaf node (i.e. a task with children)
- have all its children in a final state.

When a model is disabled by the orchestrator:
- its address is removed to effectively make it inaccessible
- an "asset_disabled" event is dispatched with the model key

Disabling a model is **not** a suppression of the model, that is up to the backends to remove their models.

Disabling a model is triggered by backends by calling the `ModelService.DisableModel` RPC method with a model key.
This may fail and return an `ErrCannotDisableModel` if one of the model cannot be disabled.
Only the worker which created the model can disable it.

### Disabling strategy

It is worth noting that disabling intermediary models relies entirely on the client backend.
The orchestrator only make necessary data available, but ultimately it is up to the backend to take a decision.

Here is how the chronology to disable models may look like, from a *client* (backend) point of view:

- [event]: task A done -> not a task created by the backend? skip it
- [event]: task B done
- [rpc]: GetInputModels(task A key) -> list of models
- [rpc]: for each model: CanDisableModel(modelKey)? -> if not, skip it
- [rpc]: for each model eligible to be disabled: DisableModel(modelKey)
