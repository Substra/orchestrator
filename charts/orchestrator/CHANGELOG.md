# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [8.0.0] - 2023-05-04

### Added

- allow using an external database in standalone mode ([#210](https://github.com/Substra/orchestrator/pull/210))

### Changed

- BREAKING: `postgresql.enabled` is now called `postgresql.subchartEnabled`

## [7.4.13] - 2023-05-11

### Changed

- bump app version to `0.34.0`

## [7.4.12] - 2023-03-31

### Changed

- bump app version to `0.33.0`

## [7.4.11] - 2023-02-18

### Changed

- make image tag default to the chart `appVersion`, and set it to `null` in the default values

## [7.4.10] - 2023-01-31

### Changed

- bump app version to `0.32.0`

## [7.4.9] - 2023-01-30

### Changed

- change contact information

## [7.4.8] - 2023-01-09

### Changed

- bump app version to `0.31.1`

## [7.4.7] - 2022-12-19

### Changed

- bump app version to `0.31.0`

## [7.4.6] - 2022-10-03

### Changed

- bump app version to `0.30.0`

## [7.4.5] - 2022-10-03

### Changed

- bump app version to `0.29.0`

## [7.4.5] - 2022-09-26

### Changed

- health probe now chech liveness and readiness with `ingress.hostname` instead of an internal service hostname when `ingress.hostname` is defined.

## [7.4.4] - 2022-09-26

### Changed

- bump app version to `0.28.0`

## [7.4.3] - 2022-09-19

### Changed

- bump app version to `0.27.0`

## [7.4.2] - 2022-09-12

### Changed

- bump app version to `0.26.1`

## [7.4.1] - 2022-09-05

### Changed

- bump app version to `0.26.0`

## [7.4.0] - 2022-09-05

### Changed

- update image registry

## [7.3.0] - 2022-09-01

### Changed

- enable TCP keepalive on PostgreSQL

## [7.2.3] - 2022-08-29

### Changed

- bump app version to `0.25.0`

## [7.2.2] - 2022-08-26

### Fixed

- add image pull secrets to migrations job

## [7.2.1] - 2022-08-22

### Changed

- bump app version to `0.24.0`

## [7.2.0] - 2022-08-17

### Removed

- remove RabbitMQ

## [7.1.11] - 2022-08-17

### Changed

- bump app version to `0.23.0`

## [7.1.10] - 2022-08-09

### Changed

- bump app version to `0.22.0`

## [7.1.9] - 2022-08-01

### Changed

- bump app version to `0.21.0`

## [7.1.8] - 2022-07-28

### Added

- initContainer in the migration job to wait for postgres to be ready

## [7.1.7] - 2022-07-26

### Removed

- fabric configmap mount on /var/hyperledger/xxx

## [7.1.6] - 2022-07-25

### Changed

- bump app version to `0.20.0`

## [7.1.5] - 2022-07-11

### Changed

- bump app version to `0.19.1`

## [7.1.4] - 2022-07-11

### Changed

- bump app version to `0.19.0`

## [7.1.3] - 2022-07-05

### Changed

- bump app version to `0.18.0`

## [7.1.2] - 2022-06-20

### Changed

- bump app version to `0.17.0`

## [7.1.1] - 2022-06-15

### Fixed

- condition for migration job deployment

## [7.1.0] - 2022-06-13

### Changed

- rename node to organization

## [7.0.1] - 2022-06-07

### Removed

- Commented code in the fabric configmap.

## [7.0.0] - 2022-06-03

### Changed

- Dependencies versions for RabbitMQ and PostgreSQL including major upgrades.

## [6.2.2] - 2022-06-02

### Changed

- Dependencies versions for RabbitMQ and PostgreSQL

## [6.2.1] - 2022-05-25

### Fixed

- Issue with the namespace declaration in the ServiceMonitor resource

## [6.2.0] - 2022-05-23

### Added

- Support for ServiceMonitor resource creation directly from the chart

## [6.1.0] - 2022-05-20

### Added

- Possibility to use an _Issuer_ instead of a _ClusterIssuer_ for certificate generation.

## [6.0.0] - 2022-05-19

### Removed

- Ingress annotation `kubernetes.io/ingress.class: nginx`, you will now need to set this annotation manually in your own values.

## [5.2.0] - 2022-04-21

### Added

- Helm hook job to run DB migrations

## [5.1.0] - 2022-03-04

### Changed

- Renamed `logSQL` to `logSQLVerbose` (#587)

## [5.0.4] - 2022-03-02

### Fixed

- Rabbitmq settings when installing the chart without a password specified

### Changed

- Switch to explicit registry for images name
- Point to existing images
- bump appVersion to `0.6.1`
- default value for `verifyClientMSPID` is now `false`

## [5.0.3] - 2022-02-28

### Added

- auto generation of the values documentation

### Changed

- change to some values default value (from `nil` to `""`) with no impact on the generated output

## [5.0.2] - 2021-12-29

### Added

- configuration flag (`metrics.enabled`) to expose prometheus metrics

## [5.0.1] - 2021-12-22

### Changed

- Bump postgresql dependency from 10.13.8 to 10.13.14
- Bump rabbitmq dependency from 8.9.1 to 8.24.12

## [5.0.0] - 2021-11-30

### Changed

- Client CA certs volumes are kebab-case
- BREAKING: `orchestrator.tls.mtls.clientCACerts` now takes a list of secrets

## [4.0.1] - 2021-11-25

### Changed

- Bump postgresql dependency from 10.3.6 to 10.13.8

## [4.0.0] - 2021-11-04

### Added

- Create a _[Certificate](https://cert-manager.io/docs/concepts/certificate/)_ resource as part of the chart

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
