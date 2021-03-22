# Configuration

Orchestrator's binaries take most of their configuration from environment variables.

Unless specified, all settings are mandatory.

## Server settings

| Env Var                               | mode                    | type                           | usage                                                                                                |
|---------------------------------------|-------------------------|--------------------------------|------------------------------------------------------------------------------------------------------|
| `ORCHESTRATOR_MODE`                   | standalone, distributed | enum: `standalone`/`chaincode` | specify in which mode to run the orchestrator (defaults to `standalone`)                             |
| `ORCHESTRATOR_TLS_ENABLED`            | standalone, distributed | bool: `true`/`false`           | whether to add TLS on transport                                                                      |
| `ORCHESTRATOR_TLS_CERT_PATH`          | standalone, distributed | string                         | path of the certificate to use                                                                       |
| `ORCHESTRATOR_TLS_KEY_PATH`           | standalone, distributed | string                         | path of the key to use                                                                               |
| `ORCHESTRATOR_MTLS_ENABLED`           | standalone, distributed | bool: `true`/`false`           | whether to enable mutual TLS                                                                         |
| `ORCHESTRATOR_TLS_SERVER_CA_CERT`     | standalone, distributed | string                         | path of the CA certificate to use                                                                    |
| `ORCHESTRATOR_TLS_CLIENT_CA_CERT_DIR` | standalone, distributed | string                         | directory containing CA certificates of the client                                                   |
| `ORCHESTRATOR_NETWORK_CONFIG`         | distributed             | string                         | path of the hyperledger fabric's network configuration                                               |
| `ORCHESTRATOR_FABRIC_CERT`            | distributed             | string                         | path of the certificate to present to fabric's peer                                                  |
| `ORCHESTRATOR_FABRIC_KEY`             | distributed             | string                         | path of the key corresponding to fabric's certificate                                                |
| `ORCHESTRATOR_DATABASE_URL`           | standalone              | string                         | [postgresql connection string](http://www.postgresql.cn/docs/13/libpq-connect.html#LIBPQ-CONNSTRING) |
| `ORCHESTRATOR_AMQP_DSN`               | standalone              | string                         | [rabbitmq connection string](https://www.rabbitmq.com/uri-spec.html)                                 |
| `ORCHESTRATOR_VERIFY_CLIENT_MSP_ID`   | standalone, distributed | bool: `true`/`false`           | whether to check that client certificate matches the MSPID header                                    |

## Forwarder settings

**Note**: forwarder is only meaningful in distributed mode

| Env Var                    | type   | usage                                                                 |
|----------------------------|--------|-----------------------------------------------------------------------|
| `FORWARDER_NETWORK_CONFIG` | string | path of the hyperledger fabric's network configuration                |
| `FORWARDER_FABRIC_CERT`    | string | path of the certificate to present to fabric's peer                   |
| `FORWARDER_FABRIC_KEY`     | string | path of the key corresponding to fabric's certificate                 |
| `FORWARDER_AMQP_DSN`       | string | [rabbitmq connection string](https://www.rabbitmq.com/uri-spec.html)  |
| `FORWARDER_CONFIG_PATH`    | string | which channel/chaincode combination to forward events for (see below) |
| `FORWARDER_MSPID`          | string | MSP ID to use to connect to the channels                              |

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
