# Orchestrator

The orchestrator is the central component managing Substra [assets](./assets/index.md).

It exposes a [gRPC API](./api.md) for clients to interact with it.
A client may also be interested in listening to relevant [orchestration events](./events.md).

When contributing a new asset, refer to the [tutorial-like](./asset-dev.md) document.
Make sure to follow the [naming conventions](./naming.md).

## [Schemas](./schemas)

[The database diagram](./schemas/standalone-database.svg) is automatically generated using [tbls](https://github.com/k1LoW/tbls).
