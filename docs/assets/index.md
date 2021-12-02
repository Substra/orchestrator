# Assets

Here we describe the assets and how they are manipulated by the orchestrator.

Asset list:

- [ComputeTask](./computetask.md)
- [DataManager](./datamanager.md)
- [DataSamples](./datasamples.md)
- [Model](./model.md)

Here is an overview of how assets relate to each other:

![](../schemas/entities.svg)

## Assets lifecycle

The assets all serve a single goal: produce models and measure their performances.

It all starts with data registration: a **[DataManager](./datamanager.md)** is created, defining how to open specific samples (opener).

Then, **[DataSamples](./datasamples.md)** are attached to the **DataManager**.
These samples can be either used during both training and testing or restricted to be only used for testing.

Algorithms (named **Algos**) are then created and embed: the actual python algorithm, a Dockerfile to build it and a description file.

**Metrics** reference scripts which evaluate the performances of a model against test data.

Once all those basic assets are defined, a **ComputePlan** can be registered.
A ComputePlan is a way to group tasks [ComputeTasks](./computetask.md) related to a common project.
Tasks are related to one another in a directed acyclic graph: output models are consumed by tasks of higher rank.

The final step is to define the graph of **ComputeTasks**.
Tasks are the core asset of the orchestrator, they represent the execution of a model training or prediction (TestTask).
They glue together the **Algos**, **DataManager**, **DataSamples** and optionally a **Metric**.

When a task is processed, it outputs one or more **Model** or **Performance**.
The produced **[Model](./model.md)** will inherit its permissions from different parents: task definition, data processed, etc.
See more details in [related section of the doc](../permissions.md).

Once all outputs of a task are registered, the task status is switched to `DONE` and its children can be processed.
There are more details in the [compute task documentation](./computetask.md).
