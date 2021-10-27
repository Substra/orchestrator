## Substra Orchestrator

This repository contains the logic to orchestrate Substra assets.

## Mission statement

This component's purpose is to orchestrate task processing in multiple channels of _connect_ partners:

- it is the single source of truth of _connect_ nodes;
- it exposes necessary data to _connect_ instances to process their tasks and register their assets;
- its API is aimed to serve backends, not end-users;
- it works the same way in both standalone and distributed mode;
- it enforces that all registered data are valid;
- it ensures data consistency under multiple concurrent requests;

## Building the orchestrator

#### Dev tools versions

Make sure you have theses requirements fulfilled before trying to build the orchestrator:

- [go](https://golang.org/): v1.16
- [protoc](https://github.com/protocolbuffers/protobuf): v3.14.0
- [proto-gen-go](https://grpc.io/docs/languages/go/quickstart/#prerequisites): v1.25.0
- [go-bindata](https://github.com/go-bindata/go-bindata): v3.1.0
- [golang-migrate](https://github.com/golang-migrate/migrate): optional, used to create migration files
- [skaffold](https://skaffold.dev/): used to run the orchestrator locally
- [mockery](https://github.com/vektra/mockery#installation): used to generate mocks

#### Build

`make`

#### Run tests

`make test`

End to end testing requires a running orchestrator.
Assuming you have one up and ready on orchestrator.node-1.com port 443, here is how to launch the tests:

```bash
make ./bin/e2e-tests
./bin/e2e-tests -tls \
    -cafile ./examples/tools/ca.crt \
    -keyfile ./examples/tools/client-org-1.key \
    -certfile ./examples/tools/client-org-1.crt \
    -server_addr orchestrator.node-1.com:443
```

Refer to `./bin/e2e-tests --help` for more options (you can run tests by name or tag).

## Developping the orchestrator

An overview of the code structure is [available in the docs directory](./docs/architecture.md) and there is also a [documentation of the assets](./docs/assets/README.md).
If you are interested in adding a new asset there is a [step by step documentation](./docs/asset-dev.md) on this subject.

A good entry point to get an overview of the codebase is to launch `godoc -http=:6060` and [open module documentation](http://localhost:6060/pkg/github.com/owkin/orchestrator/).

### Standalone mode

When running in standalone mode, the orchestrator needs a [postgres](https://www.postgresql.org/)
database to persist its data and a [rabbitmq](https://www.rabbitmq.com/) broker to dispatch events.

To launch the orchestrator:
```bash
skaffold dev
skaffold run
```

Assuming `orchestrator.node-1.com` is pointing to your local k8s cluster IP (edit your `/etc/hosts` file for that), the following command should list available services:
```bash
grpcurl -insecure orchestrator.node-1.com:443 list
```

You can also deploy [connect-backend](https://github.com/owkin/connect-backend/tree/orchestrator) (note that this is the `orchestrator` branch) with a `skaffold dev` or `skaffold run`

### Distributed mode

In distributed mode, the orchestrator only requires a matching chaincode.
So you need to build the chaincode image (from this repo) to be used in `hlf-k8s` in your k8s cluster.
Choose a tag (example uses `dev`) and update hlf-k8s' values accordingly.

```bash
# If you use minikube, run `eval $(minikube -p minikube docker-env)` before the `docker build` command
# If you use kind, run `kind load docker-image gcr.io/connect-314908/orchestrator-chaincode:dev` after the `docker build` command
docker build -f docker/orchestrator-chaincode/Dockerfile -t gcr.io/connect-314908/orchestrator-chaincode:dev .
```

Make sure you deploy [connect-hlf-k8s](https://github.com/owkin/connect-hlf-k8s/tree/orchestrator) (note that this is the `orchestrator` branch) with a `skaffold dev` or `skaffold run`

Then, in the orchestrator repo:

```bash
skaffold dev -p distributed
skaffold run -p distributed
```

Assuming `orchestrator.node-1.com` and `orchestrator.node-2.com` are pointing to your local k8s cluster IP (edit your `/etc/hosts` file for that), the following command should list available services:
```bash
grpcurl --cacert examples/tools/ca.crt --key examples/tools/client-org-1.key --cert examples/tools/client-org-1.crt --rpc-header 'mspid: MyOrg1MSP' --rpc-header 'channel: mychannel' --rpc-header 'chaincode: mycc' orchestrator.node-1.com:443 list
```

You can also deploy [connect-backend](https://github.com/owkin/connect-backend/tree/orchestrator) (note that this is the `orchestrator` branch) with a `skaffold dev -p distributed` or `skaffold run -p distributed`

### Testing

You can call the local orchestrator gRPC endpoint using [evans](https://github.com/ktr0731/evans)

```bash
evans --tls --cacert examples/tools/ca.crt --host orchestrator.node-1.com -p 443 -r repl --cert examples/tools/client-org-1.crt --certkey examples/tools/client-org-1.key
```

Then you can launch call like this :
```
package orchestrator
service NodeService
header mspid=MyOrg1MSP channel=mychannel chaincode=mycc
call QueryNodes
```

Note that you need your ingress manager to support SSL passthrough (`--enable-ssl-passthrough` with nginx-ingress).
Refer to [the wiki](https://github.com/owkin/orchestrator/wiki/Enabling-ssl-passthrough-for-ingress-in-minikube) for detailed instructions.
