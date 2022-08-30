# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Removed

- ModelService.GetComputeTaskInputModels, use ComputeTaskAPI.GetInputAssets instead
- Test task rank special case: rank is not inherited from parent task anymore

### Deprecated

- `NewComputeTask.parent_task_keys` is deprecated since parent tasks are determined from task inputs

## [0.25.0] - 2022-08-29

### Added

- New RPC to disable task outputs

### Deprecated

- ModelService.GetComputeTaskInputModels, use ComputeTaskAPI.GetInputAssets instead

## [0.24.0] - 2022-08-22

### Fixed

- Properly register compute task outputs in distributed mode

### Removed

- (BREAKING) Remove RabbitMQ

## [0.23.0] - 2022-08-17

### Added

- New service methods to update algo, compute_plan and data manager name
- gRPC method to get task input assets

### Changed

- Prevent duplicate model registration based on task output definition
- Switched to zerolog logging library

## [0.22.0] - 2022-08-09

### Added

- Build images with go 1.19
- Add a `Transient` field to the task inputs
- Return an error in distributed mode if a stored event has invalid event or asset kind
- Associate asset with task output on registration

### Removed

- Task counts by status from ComputePlan responses

## [0.21.0] - 2022-08-01

### Added

- Introduce gRPC SubscribeToEvents method in distributed mode

### Changed

- Validate task inputs

### Fixed

- In standalone mode, lock the `events` table when inserting events to prevent
  missing events in `SubscribeToEvents` gRPC stream

### Removed

- Category filter from QueryAlgos rpc
- Legacy compute task permission fields

## [0.20.0] - 2022-07-25

### Added

- Introduce gRPC SubscribeToEvents method in standalone mode
- Dispatch updated asset event on ComputePlan cancellation

### Removed

- Automatic transition to DONE when registering models or performances.

### Changed

- updated grpc healthprobe to 0.4.11 in server image
- updated rabbitmq/amqp091-go lib to 1.4.0

### Fixed

- properly ignore mocks when building image locally

## [0.19.1] - 2022-07-13

### Fixed

- SQL query for organization with null address

## [0.19.0] - 2022-07-11

### Added

- Organization hostname in the organization object
- CancelationDate in the compute plan object

### Fixed

- SQL logging was enabled when `METRICS_ENABLED` flag was passed instead of documented `LOG_SQL_VERBOSE`
- Prevent disabling model if task has only predict or test children
- Don't timeout when canceling a compute plan

## [0.18.0] - 2022-07-05

### Removed

- Metadata set in events

### Changed

- (BREAKING) Removed the `MetricKeys` property of test tasks in favor of the generic `Algo` field

## [0.17.0] - 2022-06-20

### Added

- Enable transition to DONE through ApplyTaskAction

## [0.16.0] - 2022-06-14

### Changed

- (BREAKING) rename node to organization

### Fixed

- allow a worker to cancel a task it does not own

## [0.15.0] - 2022-06-07

### Added

- Introduce Predict task type
- Introduce compute task outputs

### Changed

- use go test to run e2e tests

## [0.14.0] - 2022-05-31

### Added

- Introduce empty compute plan status

### Changed

- base docker image from alpine 3.15 to alpine 3.16

### Fixed

- event asset migration

### Changed

- only update status on task update.

## [0.13.2] - 2022-05-24

### Fixed

- `conn busy` error when querying Tasks

## [0.13.1] - 2022-05-24

### Fixed

- `conn busy` error when querying Algos

## [0.13.0] - 2022-05-23

### Fixed

- In standalone mode, truncate TimeService time to microsecond resolution to match
  PostgreSQL timestamp resolution.

### Changed

- Disable CGO.
- More validation of Algo inputs (data managers / data samples)

### Added

- Introduce compute task inputs **existing tasks won't have any inputs**
- Embed historical assets in the event messages.

## [0.12.0] - 2022-05-16

### Added

- New mandatory name field to compute plan

### Changed

- Remove event column

## [0.11.0] - 2022-05-09

### Added

- Add a new `ALGO_PREDICT` algo category

### Changed

- Validate algo inputs and outputs

## [0.10.0] - 2022-05-03

### Changed

- Remove model asset column
- Remove performance asset column
- Remove datamanager asset column
- Remove datasample asset column
- Algos now have Inputs and Outputs
- The orchestrator-server doesn't run DB migrations on startup anymore

### Removed

- `ASSET_METRIC` kind

## [0.9.2] - 2022-04-15

### Changed

- Build with go 1.18

### Fixed

- Update failure report asset column migration to prevent null value error when migrating a populated database
- Parent tasks keys format validation

## [0.9.1] - 2022-04-13

### Fixed

- Order parent tasks keys by task position

## [0.9.0] - 2022-04-13

### Added

- Added ALGO_METRICS Algo category

### Changed

- Remove compute task asset column
- QueryAlgos filter "Category" is now "Categories"
- Remove failure report asset column

### Removed

- Metrics gRPC routes. Use Algo gRPC routes and ALGO_METRICS category instead.

## [0.8.0] - 2022-04-11

### Added

- Allow querying datasamples by keys

### Changed

- Remove node asset column
- Remove algo asset column
- Remove compute plan asset column

### Fixed

- Do not panic on nil filter

## [0.7.0] - 2022-03-29

### Added

- Expose gRPC metrics
- Expose database transaction and events metrics
- Expose task metrics

### Changed

- Log SQL errors regardless of log level
- Remove `asset` column of `nodes` table

### Fixed

- Publish events sequentially, preserving the order

## [0.6.1] - 2022-03-01

### Added

- add support for graceful shutdown on `SIGTERM` signal

### Changed

- removed codegen layer and implicit protojson serialization

### Fixed

- Cancel all tasks when cancelling a compute plan
- Check for compute plan existence on task registration
- Disallow registration of tasks on a compute plan you don't own

## [0.6.0] - 2022-02-18

### Added

- add `Start` and `End` timestamp filters for `EventQueryFilter`
- support composite tasks with two composite parents
- Add migration logs
- add owner field to failure report asset
- Add a new endpoint to register multiple models at the same time

### Changed

- return `datasamples` list in `RegisterDataSamplesResponse`
- return `tasks` list in `RegisterTasksResponse`
- store the error type of a failed compute task in a failure report instead of an event
- improve performance of `compute_tasks` SQL indexes by using dedicated columns instead of JSONB
- improve performance of compute plan queries by leveraging a specific index for status count
- isolation level of read-only queries in standalone mode is now [READ COMMITTED](https://www.postgresql.org/docs/current/transaction-iso.html#XACT-READ-COMMITTED)
- improve performance of model SQL indexes by using dedicated columns instead of JSONB

### Fixed

- set the correct name of the `RegisterFailureReport` service method used in distributed mode
- Return the correct models in `GetComputeTaskInputModels` for composite tasks
- timestamp comparison when performing event sorting and filtering in PostgreSQL
- ComputePlan query now uses correct SQL indexes
- Incorrect sort order when checking parent task compatibility

### Deprecated

- `RegisterModel` gRPC method

## [0.5.0] - 2022-01-16

### Added

- add a `logs_permission` field to the Dataset asset
- add a `GetDataSample` method to the DataSample service

## [0.4.0] - 2022-01-05

### Added

- add filter for compute plan query
- chaincode now properly propagate request ID in every logs
- log events as JSON
- add FailureReport asset to store compute task failure information

## [0.3.0] - 2021-11-30

### Added

- sort queried events
- expose basic metrics from server, chaincode and forwarder behind `METRICS_ENABLED` feature flag
- filter queried events on metadata

## [0.2.0] - 2021-11-02

### Changed

- (BREAKING) Replace objective by metric
- (BREAKING) Multiple metrics and performances per test task
- fail gRPC healthcheck and stop serving on message broker disconnection

### Added

- Get task counts grouped by status when querying compute plans

### Fixed

- Events queried from the gRPC API now have their channel properly set
- Leverage asset_key index when querying events

## [0.1.0] - 2021-10-04

### Fixed

- Stable sorting of tasks

### Added

- Expose the orchestrator version and chaincode version

## [0.0.2] - 2021-09-16

### Added

- Expose worker in task event metadata
- Assets expose a creation date
- Query algo by compute plan
- Handle event backlog
- Retry on fabric timeout
- Add request ID to log context

### Changed

- Do not retry on assets out of sync
- Do not compute plan status on model deletion
- Reuse gateway connection in distributed mode
- Replace readinessProbe by startupProbe
- Do not cascade canceled status

### Fixed

- Properly retry on postgres' serialization error
- Filtering events by asset in distributed mode
- Input models for composite child of aggregate

## [0.0.1] - 2021-06-29


### Added

- Automatic generation of graphviz documentation from *.proto file definition
- asset management
- asset event dispatch
- standalone database (postgresql) support
- distributed ledger (hyperledger-fabric) support
