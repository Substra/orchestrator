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

End to end testing requires some dependencies: a postgres database and a rabbitmq broker.
Assuming you use minikube, e2e tests can be run with the following:

```bash
docker run --name e2e-pg -e POSTGRES_PASSWORD=postgres -p5432:5432 -d postgres
docker run --name e2e-rabbit -p5672:5672 -d rabbitmq
export DATABASE_URL=postgresql://postgres:postgres@$(minikube ip):5432/postgres?sslmode=disable
export RABBITMQ_DSN=amqp://guest:guest@$(minikube ip):5672/
make e2e-tests
docker stop e2e-pg e2e-rabbit
docker rm e2e-pg e2e-rabbit
```

## Developping the orchestrator

An overview of the code structure is [available in the docs directory](./docs/architecture.md)
There is also a step by step documentation on [how to implement an asset](./docs/asset-dev.md)

A good entry point to get an overview of the codebase is to launch `godoc -http=:6060` and [open module documentation](http://localhost:6060/pkg/github.com/owkin/orchestrator/).

### Standalone mode

When running in standalone mode, the orchestrator needs a [postgres](https://www.postgresql.org/)
database to persist its data and a [rabbitmq](https://www.rabbitmq.com/) broker to dispatch events.

To launch the orchestrator:
```bash
skaffold dev -p solo
skaffold run -p solo
```

Or

```bash
ORCHESTRATOR_MODE="" skaffold dev
ORCHESTRATOR_MODE="" skaffold run

```

Assuming `orchestrator.node-1.com` is pointing to your local k8s cluster IP (edit your `/etc/hosts` file for that), the following command should list available services:
```bash
grpcurl -insecure orchestrator.node-1.com:443 list
```

You can also deploy [connect-backend](https://github.com/owkin/connect-backend/tree/orchestrator) (note that this is the `orchestrator` branch) with a `skaffold dev` or `skaffold run`

### Chaincode mode

In chaincode mode, the orchestrator only requires a matching chaincode:
So you need to build the chaincode image (from this repo) to be used in `hlf-k8s` in your k8s cluster

```bash
# In minikube context do not forget to set your docker-env like this: `eval $(minikube -p minikube docker-env)
docker build -f docker/chaincode/Dockerfile -t my-chaincode:1.0.0 .
```

Make sure you deploy [connect-hlf-k8s](https://github.com/owkin/connect-hlf-k8s/tree/orchestrator) (note that this is the `orchestrator` branch) with a `skaffold dev` or `skaffold run`

Then, in the orchestrator repo:

```bash
skaffold dev -p chaincode -p -solo
skaffold run -p chaincode -p -solo
```

Or

```bash
ORCHESTRATOR_MODE="chaincode" skaffold dev
ORCHESTRATOR_MODE="chaincode" skaffold run
```

Assuming `orchestrator.node-1.com` and `orchestrator.node-2.com` are pointing to your local k8s cluster IP (edit your `/etc/hosts` file for that), the following command should list available services:
```bash
grpcurl -insecure orchestrator.node-1.com:443 list
grpcurl -insecure orchestrator.node-2.com:443 list
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

Note that you need your ingress manager to support SSL passthrough (`--enable-ssl-passthrough` with nginx-ingress)

In minikube, you can patch it with something like this :

```yaml
---
spec:
  template:
    spec:
      containers:
      - name: controller
        args:
          - /nginx-ingress-controller
          - --configmap=$(POD_NAMESPACE)/nginx-load-balancer-conf
          - --report-node-internal-ip-address
          - --tcp-services-configmap=$(POD_NAMESPACE)/tcp-services
          - --udp-services-configmap=$(POD_NAMESPACE)/udp-services
          - --validating-webhook=:8443
          - --validating-webhook-certificate=/usr/local/certificates/cert
          - --validating-webhook-key=/usr/local/certificates/key
          - --enable-ssl-passthrough

```
