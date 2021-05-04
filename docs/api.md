# Orchestrator's API

## Common interface

The orchestrator should behave the same, whether it runs as in standalone or distributed mode.

The interface is the same, however additional headers might be required in distributed mode.

## Consuming the gRPC API

When running in development mode, the orchestrator exposes a gRPC endpoint on port 9000.
gRPC reflection is enabled, and protobuf definitions are in [lib/assets](../lib/assets) directory.

**distributed mode**: requests **MUST** have the following 3 headers set:

- mspid, example `MyOrg1MSP`
- channel, example `mychannel`
- chaincode, example `mycc`

**standalone mode**: requests **MUST** have the following header set:

- mspid, example `MyOrg1MSP`
- channel, example `mychannel`

## Consuming chaincode API

From a peer node:
```bash
peer chaincode invoke \
        -C mychannel \
        -n mycc \
        --tls \
        --clientauth \
        --cafile /var/hyperledger/tls/ord/cert/cacert.pem \
        --certfile /var/hyperledger/tls/server/pair/tls.crt \
        --keyfile /var/hyperledger/tls/server/pair/tls.key \
        -o network-orderer-hlf-ord.orderer:7050 \
        -c '{"Args":["orchestrator.objective:RegisterObjective", "test", "Test", "{\"checksum\":\"669831a3180f1e77e9e3c904b76d625403924303118ff97acff2d8599b9dc91b\",\"storage_address\":\"Qsdf\"}", "TestMetrics", "{\"checksum\":\"669831a3180f1e77e9e3c904b76d625403924303118ff97acff2d8599b9dc91b\",\"storage_address\":\"Test\"}", "{\"key\":\"Test\",\"sample_keys\":[\"1\",\"2\"]}", "{\"test\":\"True\"}", "{\"public\":true,\"authorized_ids\":[\"1\"]}"]}'
```
