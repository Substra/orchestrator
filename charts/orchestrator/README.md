# Orchestrator

Orchestrator implements the orchestration components used by the [Substra](https://github.com/Substra/substra) platform.

## Prerequisites

- Kubernetes 1.16+

## Changelog

See [CHANGELOG.md](https://github.com/Substra/orchestrator/blob/main/charts/orchestrator/CHANGELOG.md)

## Installing the chart

### Standalone

to install the chart with the release name `my-release` with one organization named `MyOrg1MSP`:

```bash
helm install my-release charts/orchestrator --set 'channels[0].name=mychannel' --set 'channels[0].organizations={MyOrg1MSP}'
```

## Parameters

### Global orchestrator settings

| Name                                       | Description                                                                   | Value                    |
| ------------------------------------------ | ----------------------------------------------------------------------------- | ------------------------ |
| `imagePullSecrets`                         | Image pull secrets                                                            | `[]`                     |
| `nameOverride`                             | String to partially override the `orchestrator.fullname`                      | `""`                     |
| `fullnameOverride`                         | String to fully override the `orchestrator.fullname`                          | `""`                     |
| `serviceAccount.create`                    | Enable creation of a ServiceAccount for the orchestrator pods                 | `true`                   |
| `serviceAccount.annotations`               | Annotations to add to the ServiceAccount                                      | `{}`                     |
| `serviceAccount.name`                      | Name of the created ServiceAccount                                            | `""`                     |
| `podAnnotations`                           | Orchestrator pod annotations                                                  | `{}`                     |
| `podSecurityContext`                       | Orchestrator pod security context                                             | `{}`                     |
| `securityContext`                          | Orchestrator container security context                                       | `{}`                     |
| `service.type`                             | Orchestrator service type                                                     | `ClusterIP`              |
| `service.port`                             | Orchestrator service port                                                     | `9000`                   |
| `service.nodePort`                         | Orchestrator service port on the node if service type is `NodePort`           | `""`                     |
| `metrics.enabled`                          | Expose Prometheus metrics                                                     | `false`                  |
| `metrics.serviceMonitor.enabled`           | Create ServiceMonitor resource for scraping metrics using Prometheus Operator | `false`                  |
| `metrics.serviceMonitor.namespace`         | Namespace for the ServiceMonitor resource (defaults to the Release Namespace) | `""`                     |
| `metrics.serviceMonitor.interval`          | Interval at which metrics should be scraped                                   | `""`                     |
| `metrics.serviceMonitor.scrapeTimeout`     | Timeout after which the scrape is ended                                       | `""`                     |
| `metrics.serviceMonitor.relabelings`       | RelabelConfigs to apply to samples before scraping                            | `[]`                     |
| `metrics.serviceMonitor.metricRelabelings` | MetricRelabelConfigs to apply to samples before insertion                     | `[]`                     |
| `metrics.serviceMonitor.honorLabels`       | Specify honorLabels parameter of the scrape endpoint                          | `false`                  |
| `ingress.enabled`                          | Enable ingress for Orchestrator service                                       | `false`                  |
| `ingress.ingressClassName`                 | Ingress class name                                                            | `nil`                    |
| `ingress.path`                             | path of the deault host                                                       | `/`                      |
| `ingress.hostname`                         | hostname of the default host                                                  | `""`                     |
| `ingress.extraPaths`                       | The list of extra paths to be created for the default host                    | `[]`                     |
| `ingress.pathType`                         | Ingress path type                                                             | `ImplementationSpecific` |
| `ingress.extraHosts`                       | The list of additional hostnames to be covered with this ingress record       | `[]`                     |
| `ingress.extraTls`                         | The tls configuration for hostnames to be coverred by the ingress             | `[]`                     |
| `resources.requests.cpu`                   | CPU request for the `orchestrator` container                                  | `500m`                   |
| `resources.requests.memory`                | memory request for the `orchestrator` container                               | `200Mi`                  |
| `resources.limits.cpu`                     | CPU limits for the `orchestrator` container                                   | `500m`                   |
| `resources.limits.memory`                  | memory limit for the `orchestrator` container                                 | `800Mi`                  |
| `nodeSelector`                             | Node labels used for pod assignment                                           | `{}`                     |
| `tolerations`                              | Tolerations labels for pod assignment                                         | `[]`                     |
| `affinity`                                 | Affinity settings for pod assignment                                          | `{}`                     |

### Database connection settings

| Name                                  | Description                                                                                                 | Value          |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------- | -------------- |
| `database.auth.database`              | what DB to connect to                                                                                       | `orchestrator` |
| `database.auth.username`              | what user to connect as                                                                                     | `postgres`     |
| `database.auth.password`              | what password to use for connecting                                                                         | `postgres`     |
| `database.auth.credentialsSecretName` | An alternative to giving username and password; must have `DATABASE_USERNAME` and `DATABASE_PASSWORD` keys. | `nil`          |
| `database.host`                       | Hostname of the database to connect to (defaults to local)                                                  | `nil`          |
| `database.port`                       | Port of an external database to connect to                                                                  | `5432`         |
| `database.connectionParameters`       | database URI parameters (`key=value&key=value`)                                                             | `""`           |

### PostgreSQL settings

Database included as a subchart used by default.

See Bitnami documentation: https://bitnami.com/stack/postgresql/helm

| Name                 | Description                                                     | Value  |
| -------------------- | --------------------------------------------------------------- | ------ |
| `postgresql.enabled` | Deploy a PostgreSQL instance along the orchestrator for its use | `true` |

### Orchestrator application specific parameters

| Name                                             | Description                                                                                                                                                              | Value                          |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| `orchestrator.image.registry`                    | `orchestrator` image repository                                                                                                                                          | `ghcr.io`                      |
| `orchestrator.image.repository`                  | `orchestrator` image repository                                                                                                                                          | `substra/orchestrator-server`  |
| `orchestrator.image.pullPolicy`                  | `orchestrator` image pull policy                                                                                                                                         | `IfNotPresent`                 |
| `orchestrator.image.tag`                         | `orchestrator` image tag (defaults to AppVersion)                                                                                                                        | `nil`                          |
| `orchestrator.fullnameOverride`                  | String to fully override the `orchestrator.server.fullname`                                                                                                              | `""`                           |
| `orchestrator.logLevel`                          | Orchestrator log level                                                                                                                                                   | `INFO`                         |
| `orchestrator.logSQLVerbose`                     | Log SQL statements with debug verbosity                                                                                                                                  | `false`                        |
| `orchestrator.verifyClientMSPID`                 | If true, validates incoming gRPC requests by checking the `mspid` header matches the subject organization of the client SSL certificate. See [MSPID check](#MSPID-check) | `false`                        |
| `orchestrator.txRetryBudget`                     | Duration ([go format](https://golang.org/pkg/time/#ParseDuration)) during which the transaction can be retried in case of conflicting writes                             | `500ms`                        |
| `orchestrator.tls.createCertificates.enabled`    | If true creates a cert-manager _Certificate_ resource for the Orchestrator                                                                                               | `false`                        |
| `orchestrator.tls.createCertificates.domains`    | A list of domains to be covered by the generated certificate                                                                                                             | `[]`                           |
| `orchestrator.tls.createCertificates.duration`   | TTL of the Orchestrator certificate                                                                                                                                      | `2160h`                        |
| `orchestrator.tls.createCertificates.issuer`     | _Issuer_ or _ClusterIssuer_ responsible for the creation of this _Certificate_                                                                                           | `""`                           |
| `orchestrator.tls.createCertificates.issuerKind` | Certificate issuer kind (`Issuer` or `ClusterIssuer`)                                                                                                                    | `ClusterIssuer`                |
| `orchestrator.tls.enabled`                       | If true, enable TLS for the orchestrator gRPC endpoint                                                                                                                   | `false`                        |
| `orchestrator.tls.secrets.pair`                  | A secret containing the server TLS cert/key pair `tls.crt` and `tls.key`                                                                                                 | `orchestrator-tls-server-pair` |
| `orchestrator.tls.cacert`                        | A ConfigMap containing the server TLS CA cert `cat.crt`                                                                                                                  | `orchestrator-tls-cacert`      |
| `orchestrator.tls.mtls.enabled`                  | If true, enable TLS client verification                                                                                                                                  | `false`                        |
| `orchestrator.tls.mtls.clientCACerts`            | A map whose keys are names of the CAs, and values are a list of configmaps containing CA certificates                                                                    | `{}`                           |

### Channels settings

| Name       | Description                                | Value |
| ---------- | ------------------------------------------ | ----- |
| `channels` | List of channels and their members (MSPID) | `[]`  |

### migration job settings

| Name                          | Description                                               | Value |
| ----------------------------- | --------------------------------------------------------- | ----- |
| `migrations.fullnameOverride` | String to fully override the `migrations.server.fullname` | `""`  |

### Images

| Name                                   | Description                               | Value       |
| -------------------------------------- | ----------------------------------------- | ----------- |
| `initImages.initPostgresql.repository` | PostgreSQL image                          | `postgres`  |
| `initImages.initPostgresql.tag`        | PostgreSQL tag                            | `17`        |
| `initImages.initPostgresql.registry`   | The registry to pull the PostgreSQL image | `docker.io` |


## Usage

### TLS

TLS can be configured for both the gRPC endpoint and the RabbitMQ endpoint.

### Orchestrator endpoint

Use this sample configuration to enable mutual TLS for the orchestrator endpoint


```yaml
orchestrator:
  tls:
    enabled: true
    secrets:
      pair: orchestrator-tls-server-pair
    cacert: orchestrator-tls-cacert
    mtls:
      enabled: true # enable mutual TLS
      clientCACerts: # list of client CA certs
        orchestrator-ca: orchestrator-tls-cacert
```

#### Ingress

Here's a sample ingress configuration:

```yaml
ingress:
  enabled: true
  ingressClassName: nginx
  annotations:
    nginx.ingress.kubernetes.io/ssl-passthrough: "true"
  hosts:
  - host: orchestrator.org-1.com
    paths:
      - "/"
```

Note: The `orchestrator.verifyClientMSPID` security option (See "MSPID check" below) requires TLS termination to be enabled at the orchestrator application level (see "Orchestrator endpoint" above). TLS termination at the LB / reverse proxy level can optionally come in addition to, and not instead of, SSL termination in the app. If `orchestrator.verifyClientMSPID` is set to true and TLS termination at the ingress level is enabled, then the ingress should pass the client certificate upstream so that it can be validated.

Note: If you use nginx-ingress, use the `--enable-ssl-passthrough`.


### MSPID check

In additional to the protections offered by mutual TLS, the identity of users can be validated with the setting `Values.verifyClientMSPID`. Without this extra check, it is possible for malicious users with a valid certificate to impersonate other valid users. The verification checks that the "Subject Organization" (`O=...`) of the SSL certificate provided by the client matches the  `mspid` gRPC header supplied by the client.

This options needs both `orchestrator.tls.enabled` and `orchestrator.tls.mtls.enabled` to be true.

### Database

#### Internal

If you change connection settings for the internal database such as credentials, don't forget to also update the ones used for connecting:

```yaml
database:
  auth:
    password: abcd1234 # the password the backend will use

postgresql:
  auth:
    password: abcd1234 # the password the database expects
```

(you could use YAML anchors for this)

#### External

In standalone mode (`orchestrator.mode=standalone`), the orchestrator uses a PostgreSQL database. By default it will deploy one as a subchart. To avoid this behavior, set the appropriate values:

```yaml
database:
  host: my.database.host

  auth:
    username: my-user
    password: aStrongPassword
    database: orchestrator

postgresql:
  enabled: false
```