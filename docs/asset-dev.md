# Developping assets

## General structure

As explained in the [overview](./architecture.md), asset handling is done through the following components:

- a protobuf definition in `lib/assets/<asset>/<asset>.proto`
- a service definition in `lib/assets/<asset>/service.go`
- a grpc server in `lib/assets/<asset>/grpc.go`
- a smart contract as a `chaincode` submodule

We will later reference `lib/assets/<asset>` (where `<assets>` is your new asset name) as the "asset directory".

## Step by step implementation

### 1. Define the proto

Create an `asset.proto` file in the asset directory and define your messages and methods.
This will be automatically picked up by the Makefile to generate the corresponding go code.

You can run `make proto-codegen` to generate go code from your protobuf.

Converting from the existing chaincode is mostly a 3 steps process:

- define RPC methods from the chaincode contracts (check [main.go](https://github.com/SubstraFoundation/substra-chaincode/blob/0.2.0/chaincode/main.go#L76) for the full list)
- define the parameter messages from the existing [inputs](https://github.com/SubstraFoundation/substra-chaincode/blob/0.2.0/chaincode/input.go)
- define the response messages from the existing [output](https://github.com/SubstraFoundation/substra-chaincode/blob/0.2.0/chaincode/output.go)

### 2. Write business logic

The business logic is abstracted behind an `asset.API` interface which exposes
appropriate methods to manipulate the asset.

Something along the line of:
```go
// API defines the methods to act on Nodes
type API interface {
    RegisterNode(*Node) error
}
```

Here `Node` comes from the protobuf description and go code was generated during the previous step (`make proto-codegen`).

This `API` interface is used by both the smartcontract and the grpc server.

An `asset.Service` structure is then defined, holding a reference to the storage.
This is where the asset logic is defined.
The service must implement the API.

There are helpers defined in `github.com/owkin/orchestrator/lib/persistence/testing`
to mock the persistence layer.
Example usage:

```go
import (
    "testing"

    persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
)

func TestSomeMethod(t *testing.T) {
    mockDB := new(persistenceHelper.MockDatabase)
    mockDB.On("PutState", "uuid1", mock.Anything).Return(nil).Once()

    service := NewService(mockDB)
    service.DoSomethingWhichPutState()
}
```

### 4. Create the gRPC server

Now that we have a service to manage our new asset, let's implement the gRPC server.

Create a `grpc.go` file in the asset directory, and implement the server:

```go
type Server {
    UnimplementedAssetServiceServer // This is required by gRPC codegen
    assetService *Service // This is your asset logic
}
```

What's left to implement is for each gRPC method, how to convert from the gRPC layer to business logic and back:
gRPC input -> assetService -> gRPC output.

Since the gRPC server only rely on the business logic abstracted in `AssetService`, unit testing is only a matter of mocking the service. [stretchr/testify/mock](https://pkg.go.dev/github.com/stretchr/testify/mock) is a convenient helper for this.

### 5. Declare the smart contract

Implementing the smart contract for an asset shouldn't be much more complex than the gRPC server.
You basically adapt input/outputs around calls to the asset service.

To ease testing, a mock implementation of the [ChaincodeStubInterface](../chaincode/testing/stub_mock.go)
and [Transactioncontextinterface](../chaincode/testing/context_mock.go) are available.
