# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Organization hostname in the organization object (#805)

## [0.18.0] - 2022-07-05

### Removed
- Metadata set in events (#787)

### Changed
- (BREAKING) Removed the `MetricKeys` property of test tasks in favor of the generic `Algo` field (#776)

## [0.17.0] - 2022-06-20

### Added
- Enable transition to DONE through ApplyTaskAction (#785)

## [0.16.0] - 2022-06-14

### Changed
- BREAKING: rename node to organization (#730)

### Fixed
- allow a worker to cancel a task it does not own (#780)

## [0.15.0] - 2022-06-07

### Added
- Introduce Predict task type (#707)
- Introduce compute task outputs (#747)

### Changed
- use go test to run e2e tests (#754)

## [0.14.0] - 2022-05-31

### Added
- Introduce empty compute plan status (#726)

### Changed
- base docker image from alpine 3.15 to alpine 3.16 (#751)

### Fixed
- event asset migration (#750)

### Changed
- only update status on task update (#753).

## [0.13.2] - 2022-05-24

### Fixed
- `conn busy` error when querying Tasks (#749)

## [0.13.1] - 2022-05-24

### Fixed
- `conn busy` error when querying Algos (#748)

## [0.13.0] - 2022-05-23

### Fixed
- In standalone mode, truncate TimeService time to microsecond resolution to match
  PostgreSQL timestamp resolution (#718).

### Changed
- Disable CGO (#724).
- More validation of Algo inputs (data managers / data samples) (#736)

### Added
- Introduce compute task inputs (#691) **existing tasks won't have any inputs**
- Embed historical assets in the event messages (#715).

## [0.12.0] - 2022-05-16

### Added
- New mandatory name field to compute plan (#696)

### Changed
- Remove event column (#695)

## [0.11.0] - 2022-05-09

### Added
- Add a new `ALGO_PREDICT` algo category (#693)

### Changed
- Validate algo inputs and outputs (#699)

## [0.10.0] - 2022-05-03

### Changed
- Remove model asset column (#636)
- Remove performance asset column (#640)
- Remove datamanager asset column (#652)
- Remove datasample asset column (#666)
- Algos now have Inputs and Outputs (#641)
- The orchestrator-server doesn't run DB migrations on startup anymore (#670)

### Removed
- `ASSET_METRIC` kind (#672)

## [0.9.2] - 2022-04-15

### Changed
- Build with go 1.18 (#639)

### Fixed
- Update failure report asset column migration to prevent null value error when migrating a populated database (#658)
- Parent tasks keys format validation (#662)

## [0.9.1] - 2022-04-13

### Fixed
- Order parent tasks keys by task position (#649)

## [0.9.0] - 2022-04-13

### Added
- Added ALGO_METRICS Algo category (#628)

### Changed
- Remove compute task asset column (#619)
- QueryAlgos filter "Category" is now "Categories" (#628)
- Remove failure report asset column (#631)

### Removed
- Metrics gRPC routes. Use Algo gRPC routes and ALGO_METRICS category instead. (#628)

## [0.8.0] - 2022-04-11

### Added
- Allow querying datasamples by keys (#627)

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
