# Configuration

Orchestrator's binaries take most of their configuration from environment variables.

Unless specified, all settings are mandatory.

## Server settings

| Env Var                                                    | mode                    | type                                                               | usage                                                                                                               |
|------------------------------------------------------------|-------------------------|--------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `ORCHESTRATOR_MODE`                           | standalone | enum: `standalone`/`chaincode`                                     | specify in which mode to run the orchestrator (defaults to `standalone`)                                            |
| `TLS_ENABLED`                                 | standalone | bool: `true`/`false`                                               | whether to add TLS on transport                                                                                     |
| `TLS_CERT_PATH`                               | standalone | string (path)                                                      | path of the certificate to use                                                                                      |
| `TLS_KEY_PATH`                                | standalone | string (path)                                                      | path of the key to use                                                                                              |
| `MTLS_ENABLED`                                | standalone | bool: `true`/`false`                                               | whether to enable mutual TLS                                                                                        |
| `TLS_SERVER_CA_CERT`                          | standalone | string (path)                                                      | path of the CA certificate to use                                                                                   |
| `TLS_CLIENT_CA_CERT_DIR`                      | standalone | string (path)                                                      | directory containing CA certificates of the client                                                                  |
| `TX_RETRY_BUDGET`                             | standalone | duration ([go format](https://golang.org/pkg/time/#ParseDuration)) | duration during which the transaction can be retried in case of unserializable read/write dependencies              |
| `DATABASE_CONNECTION_STRING`                | standalone              | string                                                             | [postgresql connection string](http://www.postgresql.cn/docs/13/libpq-connect.html#LIBPQ-CONNSTRING); takes precedence over other PostgreSQL settings               |
| `DATABASE_HOSTNAME`                         | standalone              | string | |
| `DATABASE_PORT`                             | standalone              | int    | |
| `DATABASE_DATABASE`                         | standalone              | string | |
| `DATABASE_USERNAME`                         | standalone              | string | |
| `DATABASE_PASSWORD`                         | standalone              | string | |
| `DATABASE_CONNECTION_PARAMETERS`            | standalone              | string | connection parameters in space-separated `key=value` format |
| `VERIFY_CLIENT_MSP_ID`                        | standalone | bool: `true`/`false`                                               | whether to check that client certificate matches the MSPID header                                                   |
| `CHANNEL_CONFIG`                              | standalone | string (path)                                                      | where to find the [application configuration](#orchestration-configuration)                                         |
| `REPLAY_EVENTS_BATCH_SIZE`                    | standalone              | integer                                                            | the size of the batch of events used by the `SubscribeToEvents` method to replay existing events (default to `100`) |
| `GRPC_KEEPALIVE_POLICY_MIN_TIME`              | standalone | duration                                                           | the minimum amount of time a client should wait before sending a keepalive ping (default to `30s`).                 |
| `GRPC_KEEPALIVE_POLICY_PERMIT_WITHOUT_STREAM` | standalone | bool: `true`/`false`                                               | if true, server allows keepalive pings even when there are no active RPCs (default to `false`).                     |
| `LOG_LEVEL`                                   | standalone | string (INFO, WARN, NOTICE, ERROR, etc)                            | log verbosity (default to INFO)                                                                                     |
| `NO_COLOR`                                    | standalone | presence (regardless of its value)                                 | disable log color (see [no-color](https://no-color.org/))                                                           |
| `LOG_SQL_VERBOSE`                             | standalone              | bool: `true`/`false`                                               | log SQL statements with debug verbosity.                                                                            |
| `METRICS_ENABLED`                             | standalone | bool: `true`/`false`                                               | whether to enable prometheus metrics.                                                                               |

Here is a configuration example:
```yaml
listeners:
  mychannel:
    - mycc
  yourchannel:
    - yourcc
```

`listeners` in a map of *channel*: []*chaincode*.

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
