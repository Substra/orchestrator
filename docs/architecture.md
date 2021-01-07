# General architecture

The orchestrator is the core piece handling Substra assets such as Nodes, ComputePlans, TrainTuples, etc.

This repository contains two binaries: `orchestrator` and `chaincode`.

```
,-------------------------.  ,---------.
|orchestrator (standalone)|  |chaincode|
|-------------------------|  |---------|
|-------------------------|  |---------|
`-------------------------'  `---------'
              |                  |
              |                  |
        ,-----------.     ,-------------.
        |gRPC server|     |smartContract|
        |-----------|     |-------------|
        |-----------|     |-------------|
        `-----------'     `-------------'
              |                 |
        ,-----------------------------.
        |AssetService (business logic)|
        |-----------------------------|
        |-----------------------------|
        `-----------------------------'
                        |
                        |
              ,-----------------.
              |Persistence layer|
              |-----------------|
              |-----------------|
              `-----------------'

```


## Orchestrator

`orchestrator` is a gRPC server which can run in two modes:
- standalone: no ledger is needed, the orchestrator talks directly to a database
- chaincode: the orchestrator is only a facade forwarding all calls to the fabric chaincode

## Chaincode

`chaincode` is the [hyperledger fabric chaincode](https://hyperledger-fabric.readthedocs.io/en/release-2.2/chaincode4ade.html#writing-your-first-chaincode) implementation and conforms to fabric API.

## Common lib

Since both the standalone orchestrator and the chaincode have to manipulate the assets,
it makes sense that they rely on the same common lib; which you can find in the `lib` directory.

It provides abstractions to manipulate the assets and implement your own persistence layer (`persistence.Database`).

All the assets are defined by their protobuf in `lib/assets`.
You'll also find in this directory the validation implementation for each asset.

The business logic to handle those assets is defined in `lib/orchestration`,
where each asset is managed by a dedicated service.
