# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- towncrier release notes start -->
## [8.8.0] - 2024-12-16

Make init image (`postgres`) customizable

## [8.7.7] - 2024-10-14

Bump app version to 1.0.0

## [8.7.6] - 2024-09-12

Bump app version to 0.43.0

## [8.7.5] - 2024-06-13


### Changed

- Bump app version to 0.42.0

## [8.7.4] - 2024-06-10

### Fixed

- Remove Postgres default user (#430)

## [8.7.3] - 2024-06-06

### Changed

- Upgraded postgres bitnami chart to 15.4.1

## [8.7.2] - 2024-06-03


### Changed

- Bump app version to 0.41.0


## [8.7.1] - 2024-05-24

### Changed

- Allow more connection to the server to work with cloud provider

## [8.7.0] - 2024-05-21

### Added

- Network policies that:
  -  Limit connection from pods to the DB (except from server and migrations)
  -  Allow server pod to communicate with internet (outside of cluster) and pods that have the label `role-orchestrator-client: 'true'`

## [8.6.0] - 2024-04-15

### Changed

- `orchestrator-tls-cacert` is now stored as secret

## [8.5.0] - 2024-04-05

### Added

- Resources limits and requests (CPU and memory) for all containers.


## [8.4.0] - 2024-03-27

### Changed

- Bump app version to 0.40.0


## [8.3.0] - 2024-03-07

### Changed

- bump app version to `0.39.0`

## [8.2.1] - 2024-02-29

### Added

- Adds the `securityContext` required for PSA `restricted` and `baseline`.

## [8.1.1] - 2024-02-26

### Changed

- bump app version to `0.38.0`

## [8.0.1] - 2023-10-18

### Changed

- bump app version to `0.37.0`

## [8.0.0] - 2023-10-10

## Changed

- BREAKING: `postgresql` subchart version incremented to `13.1.`

## [7.5.6] - 2023-10-07

## Changed

- `wait-postgresql` initContainers refactored to Helm helper templates

## [7.5.5] - 2023-10-06

### Changed

- bump app version to `0.36.1`

## [7.5.4] - 2023-09-07

### Changed

- bump app version to `0.36.0`

## [7.5.3] - 2023-07-25

### Changed

- bump app version to `0.35.2`

## [7.5.2] - 2023-06-27

### Changed

- bump app version to `0.35.1`

## [7.5.1] - 2023-06-12

### Changed

- bump app version to `0.35.0`

## [7.5.0] - 2023-06-07

### Added

- allow using an external database in standalone mode through the `database` key in the values ([#210](https://github.com/Substra/orchestrator/pull/210))

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
