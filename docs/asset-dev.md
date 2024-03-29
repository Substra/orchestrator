# Developing assets

## General structure

As explained in the [overview](./architecture.md), asset handling is done through the following components:

- a protobuf definition in `lib/asset/<asset>.proto`
- a service definition in `lib/service/<asset>.go`
- a standalone gRPC server in `server/standalone/handlers/<asset>.go`

## Step by step implementation

### 1. Define the proto

The protobuf definition define the main structures of the asset, it is used by both the communication layer (gRPC) and the persistence layer.

Create an `asset.proto` file in the asset directory and define your messages and methods.
This will be automatically picked up by the Makefile to generate the corresponding go code.

You can run `make proto-codegen` to generate go code from your protobuf or `make proto-docgen` to generate the [protobuf documentation](./assets/protos).

**Validation**: some assets are expected to have specific properties enforced: a valid URL, SHA256 hash, string length, etc
This validation should be implemented in `lib/asset/<asset>_validation.go`, there are several existing examples.
Validation is done with [ozzo-validation](https://github.com/go-ozzo/ozzo-validation) library.

**Naming**: to have a consistent API, make sure to follow the [naming conventions](./naming.md).

### 2. Database Abstraction Layer

Once the asset defined, you can define its <abbr title="database abstraction layer">DBAL</abbr> in `lib/persistence` module.
This should be the only interface used to manipulate the asset (ie. each asset's DBAL should be isolated).

The ability to scan a raw database value into an asset, this can be done by implementing the *sql.Scanner* interface.
Converting the other way around, from an asset into a database value can be done by implementing *driver.Value*.
Examples of such implementations are available in `lib/asset/sql.go` file, it boils down to serializing/deserializing the assets in JSON.

Now, implement the DBAL interface for postgres in `server/standalone` module.

You may have to create a new table for postgres, this can be done by adding a migration in the `server/standalone/migration` module.

Mocks are automatically generated when running `make test`.

At this point tests should pass, meaning other assets are not impacted by your changes.

### 3. Write business logic

You can proceed with writing the orchestration logic for the new asset, in a `lib/service/<asset>.go` file.
To match the existing patterns, you should define (*Asset* is the place-holder for the new asset):

- an **AssetAPI** which defines the public interface of the service, i.e. what you can do with the asset.
This is what will be used by the smartcontract and the standalone gRPC service.
- an **AssetServiceProvider** which should only expose a `GetAssetService() AssetAPI` method, this is used by dependency injection.
- an **AssetDependencyProvider** which should list necessary providers (like DatabaseProvider or other services).
- then **AssetService** is the structure implementing the *AssetAPI* defined above.

Example of a service:
```go
// API defines the methods to act on Organizations
type OrganizationAPI interface {
    RegisterOrganization(*assets.Organization) error
}
```

Here `Organization` comes from the protobuf description (in `lib/asset`) and go code was generated during the previous step (`make proto-codegen`).

This *AssetAPI* interface is used by both the smartcontract and the grpc server.

#### Unit testing

There is a *MockDependenciesProvider* defined in `lib/service` module.
There are also generated helpers (prefixed with `Mock`) in `github.com/substra/orchestrator/lib/persistence` to mock the persistence layer.

### 4. Create the standalone gRPC server

Define a gRPC server in `server/standalone/handlers/<asset>.go`, it should implement `assets.AssetServiceServer` (interface generated from the protobuf).

The first thing to do in a standalone handler is to wait for an execution token.
A scheduler makes sure no more than one handler at a time is processing a request.
This is critical to maintain consistency of the data.

This server will be able to get an *service.DependenciesProvider* from the context using `ExtractProvider` function.
This is the dependency injection store, from which you can retrieve your dependencies (Database, other service, event queue, etc).

You may need the MSPID in your logic, this can also be retrieved from context through the `common.ExtractMSPID` function.

The new gRPC service can be registered into the gRPC server: this happen in the `server/standalone/server.go` file somewhere in the *GetServer* function.

At this point, standalone orchestration should be functional, you can launch the orchestrator in standalone mode and manually test your gRPC service.
