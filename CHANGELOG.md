# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Stable sorting of tasks (#371)

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
