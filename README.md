## Substra Orchestrator

This repository contains the logic to orchestrate Substra assets.

## Building the orchestrator

#### Dev tools versions

Make sure you have theses requirements fulfilled before trying to build the orchestrator:

- [go](https://golang.org/): v1.15.5
- [protoc](https://github.com/protocolbuffers/protobuf): v3.14.0
- [proto-gen-go](https://grpc.io/docs/languages/go/quickstart/#prerequisites): v1.25.0
- [go-bindata](https://github.com/go-bindata/go-bindata): v3.1.0
- [golang-migrate](https://github.com/golang-migrate/migrate): optional, used to create migration files
- [skaffold](https://skaffold.dev/): used to run the orchestrator locally

#### Build

`make`

#### Run tests

`make test`

## Developping the orchestrator

An overview of the code structure is [available in the docs directory](./docs/architecture.md)
There is also a step by step documentation on [how to implement an asset](./docs/asset-dev.md)

A good entry point to get an overview of the codebase is to launch `godoc -http=:6060` and [open module documentation](http://localhost:6060/pkg/github.com/owkin/orchestrator/).

### Standalone mode

When running in standalone mode, the orchestrator needs a [postgres](https://www.postgresql.org/)
database to persist its data and a [rabbitmq](https://www.rabbitmq.com/) broker to dispatch events.

To launch the orchestrator:
```
skaffold dev
```

Assuming `orchestrator.node-1.com` is pointing to your local k8s cluster, the following command should list available services:
```
grpcurl -insecure orchestrator.node-1.com:443 list
```

### Chaincode mode

In chaincode mode, the orchestrator only requires a matching chaincode:

```
docker build -f docker/chaincode/Dockerfile -t my-chaincode:1.0.0 .
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
