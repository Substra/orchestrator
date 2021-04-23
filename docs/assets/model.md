# Model

A model is the product of a [compute task](./computetask.md).
Not all compute task produces a model (Test task), and some may produce more than one model (Composite task).

A model inherit its permissions from its parent task.

## Registration

A model can only be registered on a task with `DOING` status.
It should be the last step before marking the task as `DONE`.

## Compatibility

A model has a category, and it can only be registered for a compatible task (y):

| model category ↓ / task category → | Train | Test | Aggregate | Composite |
|------------------------------------|-------|------|-----------|-----------|
| Simple                             | y     | n    | y         | n         |
| Head                               | n     | n    | n         | y         |
| Trunk                              | n     | n    | n         | y         |

As stated before, a task may produce several models, but can only have one of each category.
ie: composite task can only have one *head* AND one *trunk*.
