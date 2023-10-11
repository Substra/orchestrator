# Orchestrator's API

## Common interface

The orchestrator should behave the same, whether it runs as in standalone or distributed mode.

The interface is the same, however additional headers might be required in distributed mode.

## Consuming the gRPC API

When running in development mode, the orchestrator exposes a gRPC endpoint on port 9000.
gRPC reflection is enabled, and protobuf definitions are in the [lib/asset](../lib/asset) directory.

**distributed mode**: requests **MUST** have the following 3 headers set:

- mspid, example `MyOrg1MSP`
- channel, example `mychannel`
- chaincode, example `mycc`

**standalone mode**: requests **MUST** have the following header set:

- mspid, example `MyOrg1MSP`
- channel, example `mychannel`

## Consuming chaincode API

From a peer organization ("toolbox" pod):
```bash
peer chaincode invoke \
        -C mychannel \
        -n mycc \
        --tls \
        --clientauth \
        --cafile /var/hyperledger/tls/ord/cert/cacert.pem \
        --certfile /var/hyperledger/tls/server/pair/tls.crt \
        --keyfile /var/hyperledger/tls/server/pair/tls.key \
        -o network-orderer-hlf-ord.orderer.svc.cluster.local:7050 \
        -c '{"Args":["orchestrator.organization:RegisterOrganization", "{\"msg\":\"\",\"request_id\":\"\"}"]}' \
        --tlsRootCertFiles /var/hyperledger/tls/ord/cert/cacert.pem
```
