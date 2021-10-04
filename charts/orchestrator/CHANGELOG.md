# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0] - 2021-10-04

### Added
- Add NodePort configuration

## [2.0.1] - 2021-10-04

### Removed
- `orchestrator.chaincode` value was not used

## [2.0.0] - 2021-09-16

### Added
- logSQL flag to debug SQL statements (default to false)
- support for Kubernetes 1.22

### Changed
- credentials are moved from a configmap to a secret
- default log level set to INFO
- replaced readiness probes by startup probes

### Removed
- support for Kubernetes versions prior to 1.19

## [0.1.2] - 2021-08-11

### Added
- expose fabricGatewayTimeout option (#310)

## [0.1.1] - 2021-08-06

### Fixed
- Generate a valid postgres and rabbitmq service name even when we use fullnameOverride (#255)

## [0.1.0] - 2021-06-14

### Added
- deploy the orchestrator in standalone or distributed mode
