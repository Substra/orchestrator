# Substra Orchestrator

<div align="left">
<a href="https://join.slack.com/t/substra-workspace/shared_invite/zt-1fqnk0nw6-xoPwuLJ8dAPXThfyldX8yA"><img src="https://img.shields.io/badge/chat-on%20slack-blue?logo=slack" /></a> <a href="https://docs.substra.org/"><img src="https://img.shields.io/badge/read-docs-purple?logo=mdbook" /></a>
<br /><br /></div>

<div align="center">
<picture>
  <object-position: center>
  <source media="(prefers-color-scheme: dark)" srcset="Substra-logo-white.svg">
  <source media="(prefers-color-scheme: light)" srcset="Substra-logo-colour.svg">
  <img alt="Substra" src="Substra-logo-colour.svg" width="500">
</picture>
</div>
<br>
<br>

Substra is an open source federated learning (FL) software. This specific repository contains the logic to orchestrate Substra assets.

## Mission statement

This component's purpose is to orchestrate task processing in multiple channels of _Substra_ partners:

- it is the single source of truth of _Substra_ organizations;
- it exposes necessary data to _Substra_ instances to process their tasks and register their assets;
- its API is aimed to serve backends, not end-users;
- it enforces that all registered data are valid;
- it ensures data consistency under multiple concurrent requests;

## Building the orchestrator

#### Dev tools versions

Make sure you have these requirements fulfilled before trying to build the orchestrator:

- [go](https://golang.org/): v1.21+
- [protoc](https://github.com/protocolbuffers/protobuf): v3.18.0
- [proto-gen-go](https://grpc.io/docs/languages/go/quickstart/#prerequisites): v1.28.0
- [golang-migrate](https://github.com/golang-migrate/migrate): optional, used to create migration files
- [skaffold](https://skaffold.dev/): used to run the orchestrator locally
- [mockery](https://github.com/vektra/mockery#installation): used to generate mocks
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports): used for formatting

#### Build

`make`

#### Run tests

`make test`

Before running e2e tests, you may need to generate and retrieve a client certificate.

```bash
./examples/tools/download_client_cert.sh
```

End-to-end testing requires a running orchestrator.
Assuming you have one up and ready on orchestrator.org-1.com port 443, here is how to launch the tests:

```bash
go test -tags=e2e ./e2e -short -tls \
    -cafile ../examples/tools/ca.crt \
    -keyfile ../examples/tools/client-org-1.key \
    -certfile ../examples/tools/client-org-1.crt \
    -server_addr orchestrator.org-1.com:443
```

Refer to `go test -tags=e2e ./e2e -args --help` for more options.

## Developing the orchestrator

An overview of the code structure is [available in the docs directory](./docs/architecture.md) and there is also a [documentation of the assets](./docs/assets/README.md).
If you are interested in adding a new asset there is a [step by step documentation](./docs/asset-dev.md) on this subject.

A good entry point to get an overview of the codebase is to launch `godoc -http=:6060` and [open module documentation](http://localhost:6060/pkg/github.com/substra/orchestrator/).


If you want to run the orchestrator with Skaffold you will need to add the jetstack and bitnami helm repo:

```sh
helm repo add jetstack https://charts.jetstack.io
helm repo add bitnami https://charts.bitnami.com/bitnami
```

### Standalone mode

When running in standalone mode, the orchestrator needs a [postgres](https://www.postgresql.org/)
database to persist its data.

To launch the orchestrator:

```bash
skaffold dev
```

or

```bash
skaffold run
```

Assuming `orchestrator.org-1.com` is pointing to your local k8s cluster IP (edit your `/etc/hosts` file for that), the following command should list available services:

```bash
grpcurl -insecure orchestrator.org-1.com:443 list
```

You can also deploy [substra-backend](https://github.com/substra/substra-backend) with a `skaffold dev` or `skaffold run`

### Testing

You can call the local orchestrator gRPC endpoint using [evans](https://github.com/ktr0731/evans)

Before launching Evans you may need to generate and retrieve a client certificate.

```bash
./examples/tools/download_client_cert.sh
```

```bash
evans --tls --cacert examples/tools/ca.crt --host orchestrator.org-1.com -p 443 -r repl --cert examples/tools/client-org-1.crt --certkey examples/tools/client-org-1.key
```

Then you can launch call like this:

```
package orchestrator
service OrganizationService
header mspid=MyOrg1MSP channel=mychannel
call GetAllOrganizations
```

or the one-liner

```sh
echo '{}' | evans \
    --host orchestrator.org-1.com -p 443  \
    --tls \
    --cacert examples/tools/ca.crt \
    --cert examples/tools/client-org-1.crt \
    --certkey examples/tools/client-org-1.key \
    -r cli \
    --header 'mspid=MyOrg1MSP' --header 'channel=mychannel' \
    call orchestrator.OrganizationService.RegisterOrganization
```

Note that you need your ingress manager to support SSL passthrough (`--enable-ssl-passthrough` with nginx-ingress).

For additional development tips, please refer to [the documentation](./docs/development.md).
