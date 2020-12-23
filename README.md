## Substra Orchestrator

This repository contains the logic to orchestrate Substra assets.

### Developping the orchestrator

An overview of the code structure is [available in the docs directory](./docs/architecture.md)

#### Build

`make`

#### Run tests

`make test`

#### Local chaincode

To alleviate the pain of running a local hyperledger fabric network, hyperledger has a [devmode](https://hyperledger-fabric.readthedocs.io/en/latest/peer-chaincode-devmode.html).
You can run the described chaincode in devmode with a single command:

`./devchain.sh`

This does not require any configuration and will build the chaincode if needed.

#### Dev tools versions

- [go](https://golang.org/): v1.15.5
- [protoc](https://github.com/protocolbuffers/protobuf): v3.14.0
- [proto-gen-go](https://grpc.io/docs/languages/go/quickstart/#prerequisites): v1.25.0
