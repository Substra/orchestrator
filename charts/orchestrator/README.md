# Orchestrator

Orchestrator implements the orchestration components used by the [Substra](https://github.com/SubstraFoundation/substra) platform.

## Prerequisites

- Kubernetes 1.16+

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)

## Configuration

The following table lists the configurable parameters of the orchestrator chart and default values.

| Parameter                                     | Description                                                                                                                                                                   | Default                                                                                         |
| ----------------------------------            | ------------------------------------------------                                                                                                                              | ----------------------------------------------------------                                      |
| `imagePullSecrets`                            | Image pull secrets                                                                                                                                                            | `[]`                                                                                            |
| `nameOverride`                                | String to partially override `orchestrator.fullname`, `orchestrator.server.fullname`, `orchestrator.eventForwarder.fullname` and `orchestrator.rabbitmqOperator.fullname`     | nil                                                                                             |
| `fullnameOverride`                            | String to fully override `orchestrator.fullname`                                                                                                                              | nil                                                                                             |
| `nodeSelector`                                | Node labels for pod assignment                                                                                                                                                | `{}`                                                                                            |
| `affinity`                                    | Affinity settings for pod assignment                                                                                                                                          | `{}`                                                                                            |
| `resources`                                   | Resources configuration for the `orchestrator` container                                                                                                                      | `{}`                                                                                            |
| `tolerations`                                 | Toleration labels for pod assignment                                                                                                                                          | `[]`                                                                                            |
| `serviceAccount.create`                       | Enable creation of ServiceAccount for orchestrator pods                                                                                                                       | `true`                                                                                          |
| `serviceAccount.name`                         | Name of the created serviceAccount                                                                                                                                            | *Generated using the `fullname` template*                                                       |
| `serviceAccount.annotations`                  | Annotations to add to the service account                                                                                                                                     | `{}`                                                                                            |
| `podAnnotations`                              | `orchestrator` pod annotations                                                                                                                                                | `{}`                                                                                            |
| `podSecurityContext`                          | `orchestrator` pod security context                                                                                                                                           | `{}`                                                                                            |
| `securityContext`                             | `orchestrator` container security context                                                                                                                                     | `{}`                                                                                            |
| `service.type`                                | Orchestrator service type                                                                                                                                                     | `ClusterIP`                                                                                     |
| `service.port`                                | Orchestrator service port                                                                                                                                                     | `9000`                                                                                          |
| `ingress.enabled`                             | Enable ingress for orchestrator service                                                                                                                                       | `false`                                                                                         |
| `ingress.annotations`                         | Ingress annotations                                                                                                                                                           | `{"kubernetes.io/ingress.class":"nginx","nginx.ingress.kubernetes.io/backend-protocol":"GRPC"}` |
| `ingress.hosts`                               | Hosts for the ingress resource                                                                                                                                                | `[{"host":"orchestrator.node-1.com","paths":[]}]`                                               |
| `ingress.hosts[].host`                        | Hostname for the ingress resource                                                                                                                                             | nil                                                                                             |
| `ingress.hosts[].paths`                       | Paths for the host                                                                                                                                                            | nil                                                                                             |
| `ingress.tls`                                 | TLS configuration for the hosts defined in `ingress.hosts[]`                                                                                                                  | `[]`                                                                                            |
| `fabric.organization`                         | Hyperledger Fabric Peer organization name                                                                                                                                     | `MyOrg1`                                                                                        |
| `fabric.mspID`                                | Hyperledger Fabric peer MSP ID                                                                                                                                                | `MyOrg1MSP`                                                                                     |
| `fabric.channels`                             | A list of Hyperledger Fabric channels to connect to. See [hlf-k8s](https://github.com/SubstraFoundation/hlf-k8s).                                                             | `["mychannel"]`                                                                                 |
| `fabric.user.name`                            | Hyperledger Fabric Peer user name                                                                                                                                             | `User`                                                                                          |
| `fabric.peer.host`                            | Hyperledger Fabric peer hostname                                                                                                                                              | `network-org-1-peer-1-hlf-peer.org-1`                                                           |
| `fabric.peer.port`                            | Hyperledger Fabric peer port                                                                                                                                                  | `7051`                                                                                          |
| `fabric.waitForEventTimeoutSeconds`           | Time to wait for confirmation from the peers that the transaction has been committed successfully                                                                             | `45`                                                                                            |
| `fabric.strategy.invoke`                      | Chaincode invocation endorsement strategy. Can be `SELF` or `ALL` (request endorsement from all peers)                                                                        | `ALL`                                                                                           |
| `fabric.strategy.query`                       | Chaincode query endorsement strategy. Can be `SELF` or `ALL` (request endorsement from all peers)                                                                             | `SELF`                                                                                          |
| `fabric.secrets.caCert`                       | Hyperledger Fabric Peer CA Cert                                                                                                                                               | `hlf-cacert`                                                                                    |
| `fabric.secrets.user.cert`                    | Hyperledger Fabric Peer user certificate                                                                                                                                      | `hlf-msp-cert-user`                                                                             |
| `fabric.secrets.user.key`                     | Hyperledger Fabric Peer user key                                                                                                                                              | `hlf-msp-key-user`                                                                              |
| `fabric.secrets.peer.tls.client`              | Hyperledger Fabric Peer TLS client key/cert                                                                                                                                   | `hlf-tls-user`                                                                                  |
| `fabric.secrets.peer.tls.server`              | Hyperledger Fabric Peer TLS server key/cert                                                                                                                                   | `hlf-tls-admin`                                                                                 |
| `postgresql`                                  | PostgreSQL configuration. See [postgresql chart](https://github.com/bitnami/charts/tree/master/bitnami/postgresql)                                                            |                                                                                                 |
| `postgresql.enabled`                          | If true, deploy PostgreSQL                                                                                                                                                    | `true`                                                                                          |
| `postgresql.postgresqlDatabase`               | PostgreSQL database                                                                                                                                                           | `orchestrator`                                                                                  |
| `postgresql.postgresqlUsername`               | PostgreSQL user (creates a non-admin user when `postgresqlUsername` is not `postgres`)                                                                                        | `postgres`                                                                                      |
| `postgresql.postgresqlPassword`               | PostgreSQL user password                                                                                                                                                      | `postgres`                                                                                      |
| `rabbitmq`                                    | RabbitMQ configuration. See [rabbitmq chart](https://github.com/bitnami/charts/tree/master/bitnami/rabbitmq)                                                                  |                                                                                                 |
| `rabbitmq.auth.erlangCookie`                  | Erlang cookie                                                                                                                                                                 | `rabbitmqErlangCookie`                                                                          |
| `orchestrator.image.repository`               | `orchestrator` image repository                                                                                                                                               | `owkin/orchestrator`                                                                            |
| `orchestrator.image.tag`                      | `orchestrator` image tag                                                                                                                                                      | *Chart version*                                                                                 |
| `orchestrator.image.pullPolicy`               | Image pull policy                                                                                                                                                             | `IfNotPresent`                                                                                  |
| `orchestrator.fullnameOverride`               | String to fully override `orchestrator.server.fullname`                                                                                                                       | nil                                                                                             |
| `orchestrator.logLevel`                       | Orchestrator log level                                                                                                                                                        | `DEBUG`                                                                                         |
| `orchestrator.mode`                           | Orchestrator mode. Can be either "standalone" or "distributed".                                                                                                               | `standalone`                                                                                    |
| `orchestrator.chaincode`                      | Orchestrator chaincode (only used in distributed mode)                                                                                                                        | `mycc`                                                                                          |
| `orchestrator.verifyClientMSPID`              | If true, validate incoming gRPC requests by checking that the `mspid` header matches the subject organization of the client SSL certificate. See [MSPID check](#MSPID-check). | `true`                                                                                          |
| `orchestrator.tls`                            | Orchestrator TLS options. See [TLS](#TLS).                                                                                                                                    |                                                                                                 |
| `orchestrator.tls.enabled`                    | If true, enable TLS for the orchestrator gRPC endpoint                                                                                                                        | `false`                                                                                         |
| `orchestrator.tls.secrets.pair`               | A secret containing the server TLS cert/key pair `tls.crt` and `tls.key`                                                                                                      | `orchestrator-tls-server-pair`                                                                  |
| `orchestrator.tls.secrets.cacert`             | A secret containing the server TLS CA Cert `ca.crt`                                                                                                                           | `orchestrator-tls-cacert`                                                                       |
| `orchestrator.tls.mtls.enabled`               | If true, enable TLS client verification                                                                                                                                       | `false`                                                                                         |
| `orchestrator.tls.mtls.secrets.clientCACerts` | A map whose keys are names of CAs, and values are secrets containing CA certs `ca.crt`                                                                                        |                                                                                                 |
| `forwarder.image.repository`                  | Event forwarder image repository                                                                                                                                              | `owkin/forwarder`                                                                               |
| `forwarder.image.tag`                         | Event forwarder image tag                                                                                                                                                     | *Chart version*                                                                                 |
| `forwarder.image.pullPolicy`                  | Image pull policy                                                                                                                                                             | `IfNotPresent`                                                                                  |
| `forwarder.fullnameOverride`                  | String to fully override `orchestrator.eventForwarder.fullname`                                                                                                               | nil                                                                                             |
| `channels`                                    | List of channels and their members (MSPID)                                                                                                                                    | `{mychannel:[Org1MSP, Org2MSP], yourchannel: [Org1MSP, Org2MSP]}`                               |
| `rabbitmqOperator.credentials`                | Couples of username:password                                                                                                                                                  | `{Org1MSP: password1, Org2MSP: password2}`                                                      |
| `rabbitmqOperator.fullnameOverride`           | String to fully override `orchestrator.rabbitmqOperator.fullname`                                                                                                             | nil                                                                                             |

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
      secrets:
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
  - host: orchestrator.node-1.com
    paths:
      - "/"
```

Note: The `orchestrator.verifyClientMSPID` security option (See "MSPID check" below) requires TLS termination to be enabled at the orchestrator application level (see "Orchestrator endpoint" above). TLS termination at the LB / reverse proxy level can optionally come in addition to, and not instead of, SSL termination in the app. If `orchestrator.verifyClientMSPID` is set to true and TLS termination at the ingress level is enabled, then the ingress should pass the client certificate upstream so that it can be validated.

Note: If you use nginx-ingress, use the `--enable-ssl-passthrough`.

### Orchestrator RabbitMQ endpoint

Use this sample configuration to enable mutual TLS for the orchestrator RabbitMQ endpoint


```yaml
rabbitmq:
  auth:
    tls:
      enabled: true
      existingSecret: orchestrator-tls-server-pair  # Needs to have the keys ca.crt, tls.cert and tls.key
```

### MSPID check

In additional to the protections offered by mutual TLS, the identity of users can be validated with the setting `Values.verifyClientMSPID`. Without this extra check, it is possible for malicious users with a valid certificate to impersonate other valid users. The verification checks that the "Subject Organization" (`O=...`) of the SSL certificate provided by the client matches the  `mspid` gRPC header supplied by the client.

This options needs both `orchestrator.tls.enabled` and `orchestrator.tls.mtls.enabled` to be true.
