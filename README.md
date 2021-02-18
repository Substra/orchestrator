## Substra Orchestrator

This repository contains the logic to orchestrate Substra assets.

## Building the orchestrator

#### Build

`make`

#### Run tests

`make test`

## Developping the orchestrator

An overview of the code structure is [available in the docs directory](./docs/architecture.md)

### Standalone mode

When running in standalone mode, the orchestrator needs a [couchdb](https://couchdb.apache.org/)
database to persist its data and a [rabbitmq](https://www.rabbitmq.com/) broker to dispatch events.

To launch the orchestrator:
```
skaffold dev
```

Fauxton (the couchdb frontend) is accesible on http://localhost:5984/_utils

Assuming `orchestrator.node-1.com` is pointing to your local k8s cluster, the following command should list available services:
```
grpcurl -insecure orchestrator.node-1.com:443 list
```

### Chaincode mode

In chaincode mode, the orchestrator only requires a matching chaincode:

```
docker build -t my-chaincode:1.0.0 .
```

Make sure you deploy [hlf-k8s](https://github.com/SubstraFoundation/hlf-k8s) on `orchestrator` branch.

Then:
```
skaffold dev -p chaincode
```

### CA certificate

In developpement environment, we rely on self signed certificates.
Some clients (such as evans) complain that the certificate is not valid.
You can explicitely provide the certificate itself as CA (since it's self-signed):

```
kubectl get secret orchestrator-tls -n org-1 -o 'go-template={{index .data "tls.crt"}}' | base64 -d > ca.crt
# Then pass it to your client:
evans --tls --cacert ca.crt --host orchestrator.node-1.com -p 443 -r repl
```

#### Dev tools versions

- [go](https://golang.org/): v1.15.5
- [protoc](https://github.com/protocolbuffers/protobuf): v3.14.0
- [proto-gen-go](https://grpc.io/docs/languages/go/quickstart/#prerequisites): v1.25.0
