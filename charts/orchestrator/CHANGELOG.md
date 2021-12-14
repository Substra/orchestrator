# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [5.0.0] - 2021-11-30

### Changed
- Client CA certs volumes are kebabcase
- BREAKING: `orchestrator.tls.mtls.clientCACerts` now takes a list of secrets

## [4.0.1] - 2021-11-25

### Changed
- Bump postgresql dependency from 10.3.6 to 10.13.8

## [4.0.0] - 2021-11-04

### Added
- Create a _[Certificate](https://cert-manager.io/docs/concepts/certificate/)_ ressource as part of the chart

### Changed

- Moved `orchestrator.tls.mtls.secrets.clientCACerts` to `orchestrator.tls.mtls.clientCACerts`
- Moved `orchestrator.tls.secrets.cacert` to `orchestrator.tls.cacert`

## [3.0.3] - 2021-11-02
### Fixed
- Graceful shutdown of rabbitmq-operator

## [3.0.2] - 2021-10-25
### Changed
- Pass broker password to the operator through env var

## [3.0.1] - 2021-10-19
### Changed
- Use netcat from busybox image as init containers to wait for rabbitmq and postgresql

## [3.0.0] - 2021-10-07
### Changed

refacto of the Ingress

If you had a single host and a single path for your ingress:
- move `backend.ingress.hosts[0].host` to `backend.ingress.hostname`
- move `backend.ingress.hosts[0].paths[0]` to `backend.ingress.path`

If you had multiple hosts you can proceed as for a single host for your first host and then add your other hosts to `backend.ingress.extraHosts`.

The other significant change is a rename from `backend.ingress.tls` to `backend.ingress.extraTls`, the data structure inside is the same.

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
