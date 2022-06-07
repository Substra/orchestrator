# Operating chaincode in production

This document provides configs that might be useful when running the orchestrator in distributed mode.

### `fabric-config.yaml`

This file is the configuration file of Hyperledger fabric.
When moving to the orchestrator, we omitted some configuration values that might be useful at some point.
We decided to document them here.


First, you can set additional orderer configs in the settings:

Optional orderers for the channel:
```yaml
channels
  - <channel_name>:
    orderers:
      - <orderer_name>
```

Orderer physical information by orderer name:
```yaml
orderers:
  <orderer_name>:
    url: <orderer_url>
```

You can also set gRPC options for the peer connection:
```yaml
peers:
  <peer_name>:
    grpcOptions:
      ssl-target-name-override: # should be the peer host
      keep-alive-timeout:
      grpc.keepalive_time_ms:
      grpc-max-send-message-length:
      grpc-max-receive-message-length:
      grpc.http2.max_pings_without_data:
      grpc.keepalive_permit_without_calls:
```

If you want more details about this configuration file, you can look at the [hyperledger documentation](https://hyperledger-fabric.readthedocs.io/en/release-2.2/developapps/connectionprofile.html#sample).

