# Testing

The orchestrator is tested in multiple ways: with unit tests & end-to-end (e2e) tests.

They are both executed by the CI on every pull-request.

## Unit tests

Unit tests are executed with `make test`.

They should be fairly fast and are self-contained: they do not rely on any external services.
However, they only cover small parts of the application, not the whole system.

## End-to-end tests

End-to-end tests are much more extensive since they test the application from the outside:
they consume the gRPC API without any knowledge of the internal structure of the app.

Launching e2e tests requires a running target server, the easiest way to get up and running
is to target the local dev environment (see the [readme](../README.md) on how to launch the app).
Then launching the tests would look like this:

```sh
make ./bin/e2e-tests
./bin/e2e-tests -tls -cafile ./examples/tools/ca.crt -keyfile ./examples/tools/client-org-1.key -certfile ./examples/tools/client-org-1.crt -server_addr orchestrator.org-1.com:443
```

This way, we can easily test the orchestrator in both standalone and distributed mode.

The `e2e-tests` binary offers also a `--debug` flag to detail every step, beware: this is **very** verbose.

There is also a way to filter tests by name (`--name MyTest`) or tag (`--tag task`).
You can always rely on `e2e-tests --help` to have an overview of available options.

### Adding an e2e test

You are more than welcome to write e2e tests!

The core component of e2e tests is the TestClient (see e2e/client module).
This is a grpc client consuming the public API of the orchestrator.
It deals with asset keys in a way that allows for convenient test writing:
instead of relying on hardcoded values, use _key references_ like `function` or `model`.
Then, on each execution your assets will have a new key to deal with repeatability of the scenario.

Some of the client's methods take a `XXXDefaultOptions` value, which follows the builder pattern and allows for easy customization.
Most of the time you'll use the default value, but sometimes you'll need specific inputs: linking a task to a specific function, etc.
This can be done by chaining the calls, eg:

```go
appClient.RegisterModel(client.DefaultModelOptions().WithKeyRef("model1").WithTaskRef("child1"))
```

You may hit a situation where the `WithXXX` method does not exist: feel free to add it.
If you need to introduce new calls to the orchestrator, add an option with appropriate default values (i.e. working with other default values).

E2e tests are organized in scenarios (see e2e/scenario.go), each scenario is made of client actions and optional assertions.
This is where you will organize the calls to the API.

### Current limitations

Due to the fact that TLS may validate the mspid, currently tests only support single msp scenarios.

However, it might be worthy to cover multi-org scenarios: this should be done behind a flag (`./bin/e2e-tests --enable-multi-org`).
