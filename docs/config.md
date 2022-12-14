# Configuration

Orchestrator's binaries take most of their configuration from environment variables.

Unless specified, all settings are mandatory.

## Server settings

| Env Var                                                    | mode                    | type                                                               | usage                                                                                                               |
|------------------------------------------------------------|-------------------------|--------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `ORCHESTRATOR_MODE`                                        | standalone, distributed | enum: `standalone`/`chaincode`                                     | specify in which mode to run the orchestrator (defaults to `standalone`)                                            |
| `ORCHESTRATOR_TLS_ENABLED`                                 | standalone, distributed | bool: `true`/`false`                                               | whether to add TLS on transport                                                                                     |
| `ORCHESTRATOR_TLS_CERT_PATH`                               | standalone, distributed | string (path)                                                      | path of the certificate to use                                                                                      |
| `ORCHESTRATOR_TLS_KEY_PATH`                                | standalone, distributed | string (path)                                                      | path of the key to use                                                                                              |
| `ORCHESTRATOR_MTLS_ENABLED`                                | standalone, distributed | bool: `true`/`false`                                               | whether to enable mutual TLS                                                                                        |
| `ORCHESTRATOR_TLS_SERVER_CA_CERT`                          | standalone, distributed | string (path)                                                      | path of the CA certificate to use                                                                                   |
| `ORCHESTRATOR_TLS_CLIENT_CA_CERT_DIR`                      | standalone, distributed | string (path)                                                      | directory containing CA certificates of the client                                                                  |
| `ORCHESTRATOR_TX_RETRY_BUDGET`                             | standalone, distributed | duration ([go format](https://golang.org/pkg/time/#ParseDuration)) | duration during which the transaction can be retried in case of unserializable read/write dependencies              |
| `ORCHESTRATOR_NETWORK_CONFIG`                              | distributed             | string (path)                                                      | path of the hyperledger fabric's network configuration                                                              |
| `ORCHESTRATOR_FABRIC_CERT`                                 | distributed             | string (path)                                                      | path of the certificate to present to fabric's peer                                                                 |
| `ORCHESTRATOR_FABRIC_KEY`                                  | distributed             | string (path)                                                      | path of the key corresponding to fabric's certificate                                                               |
| `ORCHESTRATOR_FABRIC_GATEWAY_TIMEOUT`                      | distributed             | duration ([go format](https://golang.org/pkg/time/#ParseDuration)) | Commit timeout for all transaction submissions for the gateway                                                      |
| `ORCHESTRATOR_DATABASE_URL`                                | standalone              | string                                                             | [postgresql connection string](http://www.postgresql.cn/docs/13/libpq-connect.html#LIBPQ-CONNSTRING)                |
| `ORCHESTRATOR_VERIFY_CLIENT_MSP_ID`                        | standalone, distributed | bool: `true`/`false`                                               | whether to check that client certificate matches the MSPID header                                                   |
| `ORCHESTRATOR_CHANNEL_CONFIG`                              | standalone, distributed | string (path)                                                      | where to find the [application configuration](#orchestration-configuration)                                         |
| `ORCHESTRATOR_REPLAY_EVENTS_BATCH_SIZE`                    | standalone              | integer                                                            | the size of the batch of events used by the `SubscribeToEvents` method to replay existing events (default to `100`) |
| `ORCHESTRATOR_GRPC_KEEPALIVE_POLICY_MIN_TIME`              | standalone, distributed | duration                                                           | the minimum amount of time a client should wait before sending a keepalive ping (default to `30s`).                 |
| `ORCHESTRATOR_GRPC_KEEPALIVE_POLICY_PERMIT_WITHOUT_STREAM` | standalone, distributed | bool: `true`/`false`                                               | if true, server allows keepalive pings even when there are no active RPCs (default to `false`).                     |
| `LOG_LEVEL`                                                | standalone, distributed | string (INFO, WARN, NOTICE, ERROR, etc)                            | log verbosity (default to INFO)                                                                                     |
| `NO_COLOR`                                                 | standalone, distributed | presence (regardless of its value)                                 | disable log color (see [no-color](https://no-color.org/))                                                           |
| `LOG_SQL_VERBOSE`                                          | standalone              | bool: `true`/`false`                                               | log SQL statements with debug verbosity.                                                                            |
| `METRICS_ENABLED`                                          | standalone, distributed | bool: `true`/`false`                                               | whether to enable prometheus metrics.                                                                               |

Here is a configuration example:
```yaml
listeners:
  mychannel:
    - mycc
  yourchannel:
    - yourcc
```

`listeners` in a map of *channel*: []*chaincode*.

## Chaincode settings

**Note**: chaincode is only meaningful in distributed mode

| Env Var             | type   | usage                                                        |
|---------------------|--------|--------------------------------------------------------------|
| `CHAINCODE_CCID`    | string | chaincode id                                                 |
| `TLS_KEY_FILE`      | string | path of the TLS key                                          |
| `TLS_CERT_FILE`     | string | path of the TLS certificate                                  |
| `TLS_ROOTCERT_FILE` | string | path of the CA certificate                                   |
| `CHAINCODE_ADDRESS` | string | on which address should the chaincode listen for connections |

## Orchestration configuration

The orchestrator controls the access to channels for each call.
To that end, it needs to be aware of the channels and their allowed organizations.
This is the `ORCHESTRATOR_CHANNEL_CONFIG` file, which content could look like this for two shared channels:

```yml
---
channels:
  mychannel:
    - MyOrg1MSP
    - MyOrg2MSP
  yourchannel:
    - MyOrg1MSP
    - MyOrg2MSP
```
