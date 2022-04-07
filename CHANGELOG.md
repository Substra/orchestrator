# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Remove node asset column (#604)
- Remove algo asset column (#612)
- Remove compute plan asset column (#618)

### Fixed
- Do not panic on nil filter (#510)

## [0.7.0] - 2022-03-29

### Added
- Expose gRPC metrics (#584)
- Expose database transaction and events metrics (#589)
- Expose task metrics (#590)

### Changed
- Log SQL errors regardless of log level (#587)
- Remove `asset` column of `nodes` table (#604)

### Fixed
- Publish events sequentially, preserving the order (#600)

## [0.6.1] - 2022-03-01

### Added
- add support for graceful shutdown on `SIGTERM` signal (#557)

### Changed
- removed codegen layer and implicit protojson serialization (#535)

### Fixed
- Cancel all tasks when cancelling a compute plan (#546)
- Check for compute plan existence on task registration (#554)
- Disallow registration of tasks on a compute plan you don't own (#566)

## [0.6.0] - 2022-02-18

### Added
- add `Start` and `End` timestamp filters for `EventQueryFilter` (#482)
- support composite tasks with two composite parents (#464)
- Add migration logs (#501)
- add owner field to failure report asset (#531)
- Add a new endpoint to register multiple models at the same time (#530,#541)

### Changed
- return `datasamples` list in `RegisterDataSamplesResponse` (#486)
- return `tasks` list in `RegisterTasksResponse` (#493)
- store the error type of a failed compute task in a failure report instead of an event (#487)
- improve performance of `compute_tasks` SQL indexes by using dedicated columns instead of JSONB (#503)
- improve performance of compute plan queries by leveraging a specific index for status count (#509)
- isolation level of read-only queries in standalone mode is now [READ COMMITTED](https://www.postgresql.org/docs/current/transaction-iso.html#XACT-READ-COMMITTED) (#492)
- improve performance of model SQL indexes by using dedicated columns instead of JSONB (#539)

### Fixed
- set the correct name of the `RegisterFailureReport` service method used in distributed mode (#485)
- Return the correct models in `GetComputeTaskInputModels` for composite tasks (#499)
- timestamp comparison when performing event sorting and filtering in PostgreSQL (#491)
- ComputePlan query now uses correct SQL indexes (#500)
- Incorrect sort order when checking parent task compatibility (#507)

### Deprecated

- `RegisterModel` gRPC method (#530)

## [0.5.0] - 2022-01-16

### Added
- add a `logs_permission` field to the Dataset asset (#459)
- add a `GetDataSample` method to the DataSample service(#479)

## [0.4.0] - 2022-01-05

### Added
- add filter for compute plan query (#433)
- chaincode now properly propagate request ID in every logs (#443)
- log events as JSON (#452)
- add FailureReport asset to store compute task failure information (#456)

## [0.3.0] - 2021-11-30

### Added
- sort queried events (#417)
- expose basic metrics from server, chaincode and forwarder behind `METRICS_ENABLED` feature flag
- filter queried events on metadata (#422)

## [0.2.0] - 2021-11-02

### Changed
- (BREAKING) Replace objective by metric (#356)
- (BREAKING) Multiple metrics and performances per test task (#369)
- fail gRPC healthcheck and stop serving on message broker disconnection (#397)

### Added
- Get task counts grouped by status when querying compute plans (#400)

### Fixed
- Events queried from the gRPC API now have their channel properly set (#414)
- Leverage asset_key index when querying events

## [0.1.0] - 2021-10-04

### Fixed
- Stable sorting of tasks (#371)

### Added
- Expose the orchestrator version and chaincode version (#370)

## [0.0.2] - 2021-09-16

### Added
- Expose worker in task event metadata
- Assets expose a creation date (#328)
- Query algo by compute plan (#307)
- Handle event backlog (#288)
- Retry on fabric timeout
- Add request ID to log context

### Changed
- Do not retry on assets out of sync (#335)
- Do not compute plan status on model deletion (#329)
- Reuse gateway connection in distributed mode (#324)
- Replace readinessProbe by startupProbe (#314)
- Do not cascade canceled status (#313)

### Fixed
- Properly retry on postgres' serialization error
- Filtering events by asset in distributed mode (#321)
- Input models for composite child of aggregate (#280)

## [0.0.1] - 2021-06-29

- Automatic generation of graphviz documentation from *.proto file definition

### Added
- asset management
- asset event dispatch
- standalone database (postgresql) support
- distributed ledger (hyperledger-fabric) support
