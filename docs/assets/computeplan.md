# ComputePlan

A compute plan is a kind of *container* for [compute tasks](./computetask.md).
It does not set any expectation on the topology of tasks and trainings.

It is the entry point to act on all tasks at once.
eg: cancelling a compute plan will cancel all its cancellable (not DOING/FAILED/DONE/CANCEL) tasks.

## Termination statuses

A compute plan can be canceled by the user. In this case, the `cancelation_date` field of the compute plan
will be filled. If any of the tasks of the compute plan fails, the `failure_date` field of the compute plan will
be filled. In any other case, the compute plan won't have a termination status.
