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
| `resources`                                | Resource configuration for the `orchestrator` container                       | `{}`                     |
| `nodeSelector`                             | Node labels used for pod assignment                                           | `{}`                     |
| `tolerations`                              | Tolerations labels for pod assignment                                         | `[]`                     |
| `affinity`                                 | Affinity settings for pod assignment                                          | `{}`                     |


### PostgreSQL settings

| Name                                       | Description                                                                   | Value                     |
| ------------------------------------------ | ----------------------------------------------------------------------------- | ------------------------- |
| `postgresql.subchartEnabled`               | If true, deploy PostgreSQL as a subchart                                      | `true`                    |
| `postgresql.host`                          | Hostname of the database to connect to (defaults to local)                    | `nil`                     |
| `postgresql.port`                          | Port of an external database to connect to                                    | `nil`                     |
| `postgresql.uriParams`                     | database URI parameters                                                       | `""`                      |
| `postgresql.auth.enablePostgresUser`       | creates a PostgreSQL user                                                     | `true`                    |
| `postgresql.auth.postgresPassword`         | password for the postgres admin user                                          | `postgres`                |
| `postgresql.auth.username`                 | PostgreSQL user (creates a non-admin user when username is not `postgres`)    | `postgres`                |
| `postgresql.auth.password`                 | PostgreSQL user password                                                      | `postgres`                |
| `postgresql.auth.database`                 | PostgreSQL database the orchestrator should use                               | `orchestrator`            |
| `postgresql.primary.extendedConfiguration` | Extended PostgreSQL configuration (appended to main or default configuration) | `tcp_keepalives_idle = 5` |


### Hyperledger Fabric settings

| Name                                | Description                                                                                            | Value                                                   |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------- |
| `fabric.organization`               | Hyperledger Fabric Peer organization name                                                              | `MyOrg1`                                                |
| `fabric.mspID`                      | Hyperledger Fabric Peer MSP ID                                                                         | `MyOrg1MSP`                                             |
| `fabric.channels`                   | A list of Hyperledger Fabric channels to connect to. See [hlf-k8s](https://github.com/substra/hlf-k8s) | `["mychannel","yourchannel"]`                           |
| `fabric.user.name`                  | Hyperledger Fabric Peer user name                                                                      | `User`                                                  |
| `fabric.peer.host`                  | Hyperledger Fabric Peer hostname                                                                       | `network-org-1-peer-1-hlf-peer.org-1.svc.cluster.local` |
| `fabric.peer.port`                  | Hyperledger Fabric Peer port                                                                           | `7051`                                                  |
| `fabric.waitForEventTimeoutSeconds` | Time to wait for confirmation from the Peers that the transaction has been committed                   | `45`                                                    |
| `fabric.logLevel`                   | Log level for `fabric-sdk-go`                                                                          | `INFO`                                                  |
| `fabric.strategy.invoke`            | Chaincode invocation endorsement strategy. Can be `SELF` or `ALL` (request endorsement from all Peers) | `ALL`                                                   |
| `fabric.strategy.query`             | Chaincode query endorsement strategy. Can be `SELF` or `ALL` (request endorsement from all Peers)      | `SELF`                                                  |
| `fabric.secrets.caCert`             | Hyperledger Fabric CA Cert                                                                             | `hlf-cacert`                                            |
| `fabric.secrets.user.cert`          | Hyperledger Fabric Peer user certificate                                                               | `hlf-msp-cert-user`                                     |
| `fabric.secrets.user.key`           | Hyperledger Fabric Peer user certificate key                                                           | `hlf-msp-key-user`                                      |
| `fabric.secrets.peer.tls.client`    | Hyperledger Fabric Peer TLS client key/cert                                                            | `hlf-tls-user`                                          |
| `fabric.secrets.peer.tls.server`    | Hyperledger Fabric Peer TLS server key/cert                                                            | `hlf-tls-admin`                                         |


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
| `orchestrator.mode`                              | Orchestrator mode, either "standalone" or "distributed"                                                                                                                  | `standalone`                   |
| `orchestrator.verifyClientMSPID`                 | If true, validates incoming gRPC requests by checking the `mspid` header matches the subject organization of the client SSL certificate. See [MSPID check](#MSPID-check) | `false`                        |
| `orchestrator.txRetryBudget`                     | Duration ([go format](https://golang.org/pkg/time/#ParseDuration)) during which the transaction can be retried in case of conflicting writes                             | `500ms`                        |
| `orchestrator.fabricGatewayTimeout`              | Commit timeout ([go format](https://golang.org/pkg/time/#ParseDuration)) for all transaction submissions for the gateway (only used in distributed mode)                 | `20s`                          |
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
  annotations:
    kubernetes.io/ingress.class: nginx
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

### External database

In standalone mode, the orchestrator uses a PostgreSQL database. By default it will deploy one as a subchart. To avoid this behavior, set the appropriate values:

```yaml
postgresql:
  subchartEnabled: false
  host: my.database.host
  
  auth:
    username: my-user
    password: aStrongPassword
    database: orchestrator
```