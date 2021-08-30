# ComputePlan

A compute plan is a kind of *container* for [compute tasks](./computetask.md).
It does not set any expectation on the topology of tasks and trainings.

It is the entry point to act on all tasks at once.
eg: cancelling a compute plan will cancel all its cancellable (not DOING/FAILED/DONE/CANCEL) tasks.

## Statuses

ComputePlan status is determined from its tasks and follow the rules below (by order of evaluation):

- if any task is FAILED, the compute plan is FAILED
- if any task is CANCELED, the compute plan is CANCELED
- if all tasks are DONE, the compute plan is DONE
- if all tasks are WAITING, the compute plan is WAITING
- if no tasks are DOING or DONE and one task is TODO, the compute plan is TODO
- in any other case the compute plan is DOING
