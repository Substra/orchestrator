# Developping assets

## General structure

As explained in the [overview](./architecture.md), asset handling is done through the following components:

- a protobuf definition in `lib/asset/<asset>.proto`
- a service definition in `lib/service/<asset>.go`
- a standalone grpc server in `server/standalone/<asset>.go`
- a distributed grpc server in `server/distributed/<asset>.go`
- a smart contract as a `chaincode` submodule

## Step by step implementation

### 1. Define the proto

The protobuf definition define the main structures of the asset, it is used by both the communication layer (gRPC/chaincode) and the persistence layer.

Create an `asset.proto` file in the asset directory and define your messages and methods.
This will be automatically picked up by the Makefile to generate the corresponding go code.

You can run `make proto-codegen` to generate go code from your protobuf.

Converting from the existing chaincode is mostly a 3 steps process:

- define RPC methods from the chaincode contracts (check [main.go](https://github.com/SubstraFoundation/substra-chaincode/blob/0.2.0/chaincode/main.go#L76) for the full list)
- define the parameter messages from the existing [inputs](https://github.com/SubstraFoundation/substra-chaincode/blob/0.2.0/chaincode/input.go)
- define the response messages from the existing [output](https://github.com/SubstraFoundation/substra-chaincode/blob/0.2.0/chaincode/output.go)

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

Now, implement the DBAL interface for both storage backends: postgres in `server/standalone` module and the ledger in `chaincode/ledger` module.
You may have to create a new table for postgres, this can be done by adding a migration in `server/standalone/migration` module.

You will also have to implement this new DBAL on the mocked persistence layer in `lib/persistence/testing` module.
This should be pretty straightforward: add missing methods and pass the arguments to the underlying mock.

At this point tests should pass, meaning other assets are not impacted by your changes.

### 3. Write business logic

You can proceed with writing the orchestration logic for the new asset, in a `lib/service/<asset>.go` file.
To match the existing patterns, you should define (*Asset* is the place-holder for the new asset):

- an **AssetAPI** which defines the public interface of the service, ie: what you can do with the asset.
This is what will be used by the smartcontract and the standalone gRPC service.
- an **AssetServiceProvider** which should only expose a `GetAssetService() AssetAPI` method, this is used by dependency injection.
- an **AssetDependencyProvider** which should list necessary providers (like DatabaseProvider or other services).
- then **AssetService** is the structure implementing the *AssetAPI* defined above.

Example of a service:
```go
// API defines the methods to act on Nodes
type NodeAPI interface {
    RegisterNode(*assets.Node) error
}
```

Here `Node` comes from the protobuf description (in `lib/asset`) and go code was generated during the previous step (`make proto-codegen`).

This *AssetAPI* interface is used by both the smartcontract and the grpc server.

#### Unit testing

There is a *MockServiceProvider* defined in `lib/service` module, it should be updated to implement *AssetServiceProvider*.
You can also define a **MockAssetService** for future use by other tests (distributed and standalone gRPC service).

There are also helpers defined in `github.com/owkin/orchestrator/lib/persistence/testing` to mock the persistence layer.

### 4. Create the standalone gRPC server

Define a gRPC server in `server/standalone/<asset>.go`, it should implement `assets.AssetServiceServer` (interface generated from the protobuf).

The first thing to do in a standalone handler is to wait for an execution token.
A scheduler makes sure no more than one handler at a time is processing a request.
This is critical to maintain consistency of the data.

This server will be able to get an *service.DependenciesProvider* from the context using `ExtractProvider` function.
This is the dependency injection store, from which you can retrieve your dependencies (Database, other service, event queue, etc).

You may need the MSPID in your logic, this can also be retrieved from context through the `common.ExtractMSPID` function.

The new gRPC service can be registered into the gRPC server: this happen in the `server/standalone/server.go` file somewhere in the *GetServer* function.

At this point, standalone orchestration should be functional, you can launch the orchestrator in standalone mode and manually test your gRPC service.

### 5. Declare the smart contract

Define a smartcontract according to [contractapi](https://github.com/hyperledger/fabric-contract-api-go) in `chaincode/<asset>/contract.go`.

Make sure that on creation the smartcontract gets its properties set correctly.
Ideally in the *NewSmartContract*  method:

```go
func NewSmartContract() *SmartContract {
    contract := &SmartContract{}
    contract.Name = "orchestrator.<asset>"
    contract.TransactionContextHandler = ledger.NewContext()
    contract.BeforeTransaction = ledger.GetBeforeTransactionHook(contract)
    contract.AfterTransaction = ledger.AfterTransactionHook

    return contract
}
```

It is **essential** that the contract has:
- `BeforeTransaction` set to `ledger.GetBeforeTransactionHook(contract)` to properly set the transaction context
- `AfterTransaction` set to `ledger.AfterTransactionHook` to properly dispatch events

The TransactionContextHandler is also necessary since that's how you can retrieve the *DependenciesProvider*.

Don't forget to flag the evaluate transaction (query only) by implementing `GetEvaluateTransactions() []string`
(refer to [contractapi documentation](https://pkg.go.dev/github.com/hyperledger/fabric-contract-api-go@v1.1.1/contractapi#EvaluationContractInterface) for more details).

Using the *DependenciesProvider* is as easy as writing: `provider := ctx.GetProvider()` in your contracts.

Finally you can add your smart contract to the contract provider in `chaincode/contracts/provider.go` to have it published.

**Note**: contracts should have the same inputs and outputs than the gRPC service.
That way, the *Invocator* (more below) can transparently handle the serialization/deserialization.

#### gRPC adapter

gRPC service relying on chaincode is defined in `server/distributed/<asset>.go`

This is done the same way than for the standalone mode, except that there is a chaincode invocation instead of orchestration logic.

A specific structure, the *Invocator* is provided to invoke the chaincode.
This *Invocator* is available from the context: `invocator, err := ExtractInvocator(ctx)`.

Most of the service methods should look like this:

```go
func (a *AssetAdapter) DoSomething(ctx context.Context, input *assets.AssetDoSomethingParam) (*assets.DoSomethingResponse, error) {
    invocator, err := ExtractInvocator(ctx)
    if err != nil {
        return nil, err
    }
    response := &assets.DoSomethingResponse

    err = invocator.Call("orchestrator.asset:DoSomething", input, response)

    return response, err
}
```

The new gRPC service can be registered into the gRPC server: this happen in the `server/distributed/server.go` file somewhere in the *GetServer* function.
