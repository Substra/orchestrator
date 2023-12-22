# Configuration

Orchestrator's binaries take most of their configuration from environment variables.

Unless specified, all settings are mandatory.

## Server settings

| Env Var                                       | type                                                               | usage                                                                                                                                                 |
| --------------------------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `TLS_ENABLED`                                 | bool: `true`/`false`                                               | whether to add TLS on transport                                                                                                                       |
| `TLS_CERT_PATH`                               | string (path)                                                      | path of the certificate to use                                                                                                                        |
| `TLS_KEY_PATH`                                | string (path)                                                      | path of the key to use                                                                                                                                |
| `MTLS_ENABLED`                                | bool: `true`/`false`                                               | whether to enable mutual TLS                                                                                                                          |
| `TLS_SERVER_CA_CERT`                          | string (path)                                                      | path of the CA certificate to use                                                                                                                     |
| `TLS_CLIENT_CA_CERT_DIR`                      | string (path)                                                      | directory containing CA certificates of the client                                                                                                    |
| `TX_RETRY_BUDGET`                             | duration ([go format](https://golang.org/pkg/time/#ParseDuration)) | duration during which the transaction can be retried in case of unserializable read/write dependencies                                                |
| `DATABASE_CONNECTION_STRING`                  | string                                                             | [postgresql connection string](http://www.postgresql.cn/docs/13/libpq-connect.html#LIBPQ-CONNSTRING); takes precedence over other PostgreSQL settings |
| `DATABASE_HOSTNAME`                           | string                                                             |                                                                                                                                                       |
| `DATABASE_PORT`                               | int                                                                |                                                                                                                                                       |
| `DATABASE_DATABASE`                           | string                                                             |                                                                                                                                                       |
| `DATABASE_USERNAME`                           | string                                                             |                                                                                                                                                       |
| `DATABASE_PASSWORD`                           | string                                                             |                                                                                                                                                       |
| `DATABASE_CONNECTION_PARAMETERS`              | string                                                             | connection parameters in space-separated `key=value` format                                                                                           |
| `VERIFY_CLIENT_MSP_ID`                        | bool: `true`/`false`                                               | whether to check that client certificate matches the MSPID header                                                                                     |
| `CHANNEL_CONFIG`                              | string (path)                                                      | where to find the [application configuration](#orchestration-configuration)                                                                           |
| `REPLAY_EVENTS_BATCH_SIZE`                    | integer                                                            | the size of the batch of events used by the `SubscribeToEvents` method to replay existing events (default to `100`)                                   |
| `GRPC_KEEPALIVE_POLICY_MIN_TIME`              | duration                                                           | the minimum amount of time a client should wait before sending a keepalive ping (default to `30s`).                                                   |
| `GRPC_KEEPALIVE_POLICY_PERMIT_WITHOUT_STREAM` | bool: `true`/`false`                                               | if true, server allows keepalive pings even when there are no active RPCs (default to `false`).                                                       |
| `LOG_LEVEL`                                   | string (INFO, WARN, NOTICE, ERROR, etc)                            | log verbosity (default to INFO)                                                                                                                       |
| `NO_COLOR`                                    | presence (regardless of its value)                                 | disable log color (see [no-color](https://no-color.org/))                                                                                             |
| `LOG_SQL_VERBOSE`                             | bool: `true`/`false`                                               | log SQL statements with debug verbosity.                                                                                                              |
| `METRICS_ENABLED`                             | bool: `true`/`false`                                               | whether to enable prometheus metrics.                                                                                                                 |

Here is a configuration example:
```yaml
listeners:
  mychannel:
    - mycc
  yourchannel:
    - yourcc
```

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
