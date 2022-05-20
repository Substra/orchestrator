# Events

Throughout the life of a [ComputePlan](./assets/computeplan.md), the orchestrator will emit events which may be of interest for the clients.

For more details about the inner workings of the event dispatch, refer to [the architecture documentation](./architecture.md).

## Structure of an event

A client can receive events by listening to its dedicated AMQP queue.
All events will come serialized as JSON.

An event will hold the following fields:

- asset_kind: flag the kind of asset which triggered this event, see below for possible values;
- asset_key: the key (UUID) of the relevant asset;
- event_kind: kind of event (see below);
- channel: the channel for which the event has been dispatched;
- asset: a snapshot of the asset referenced by the event in JSON format;
- metadata: a map of keys (string) to values (string);

## Asset Kind

- node
- datasample
- algo
- datamanager
- computetask
- computeplan
- model

## Event Kind

- asset_created: occurs when a new asset has been registered
- asset_updated: occurs when an existing asset has been modified
- asset_disabled: occurs when an asset has been disabled by its owner and is not accessible anymore

## Relevant metadata

Some event will hold additional data.

**Task status change**: when a task status is updated, the *asset_updated* event will have the following metadata:

- status: string representation of the **new** status (see [ComputeTask](./assets/computetask.md))
- reason: cause of the status change, this is a sentence which can be shown or logged
