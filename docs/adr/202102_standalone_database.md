# Choice of persistence layer

* Status: accepted
* Deciders: Aurélien, Inal, Matthieu
* Date: 2021-02-25

## Context and Problem Statement

The orchestrator must be able to persist its state.
Depending on execution mode (standalone or chaincode), the persistence layer should be different.
While we don't have many alternatives in chaincode mode -- where the ledger is mandatory --
we have many possibilities for the standalone persistence.

The purpose of this ADR is to settle on a database.

Couchdb is the currently implemented backend, and we use it in a pretty basic key:value way.
This is a behavior inherited from the existing chaincode, where we store assets as serialized json (byte[]) identified by a key.

## Decision Drivers

* data integrity is paramount
* solution should not hinder evolution of data structures

## Considered Options

* couchdb
* postgresql

## Decision Outcome

Chosen option: postgres because it's the only way to have proper transactional safety.

### Positive Consequences

Switching to a relational database in standalone mode challenges the existing persistence layer.
This is a good thing because the target API is now much more specific to assets and allow for more future flexibility.
We will depart from our generic PutState strategy (inherited from the existing chaincode) to have per asset DBAL.
We can imagine things like `GetComputePlan(id string)`, `GetTrainTuplesByComputePlan(id string)`, etc.

### Negative Consequences

The main concern regarding the postgres solution was the maintenance of the database schema and migrations.
As first approach, we will implement a minimalist schema with one table shared by all serialized assets.

Even with that reduced schema, we have to manage a basic migration, but this is hopefully mitigated by leverage existing database libraries.

## Pros and Cons of the Options

### CouchDB

[couchdb](https://couchdb.apache.org/) is the same backend used by hyperledger fabric.

Couchdb is the underlying data store used by hyperledger fabric (chaincode mode),
that means we could *leak* some idiosyncrasies in the orchestration layer since they would be shared by both execution modes.
While convenient, leaking DB abstraction in the upper layer is certainly not a good practice, so that balances the potential shortcut.
So, in a way, **not** using couchdb would prevent such accidental leak.

The main drawback of couchdb is that it is not transactional,
thus we cannot ensure in standalone mode that a failing request (gRPC one) won't leave the state dirty.
That means if we ever go with couchdb, we need to add a transaction layer.
Not only this feels like reinventing the wheel, but we also should be extra careful with the implementation to never let a corruption go unnoticed.

* Good, because schemaless so it's lenient (store byte[])
* Bad, because it's not transactional (only in a document) so we can't make sure the stored state won't be corrupted

### Postgresql

[postgresql](https://www.postgresql.org/) is The World's Most Advanced Open Source Relational Database.

Postgres is transactional, so we can easily roll back changes at the end of a failed (gRPC) transaction.
It can handle unstructured data if need be: we can rely on the existing []byte state model.

An additional cost of postgres is that we have to deal with a schema and maintain it (migrations).

A nice feature though is that we have access to efficient indexing over multiple fields (asset type, owner, etc).

* Good, because it is a transactional database so we can have strong guarantees of non-corruption
* Bad, because it requires to handle schema migrations

## Links

* [initial transaction PR](https://github.com/owkin/orchestrator/pull/10)
* [mongodb/postgres comparison](https://www.aptuz.com/blog/is-postgres-nosql-database-better-than-mongodb/)
* [effectively store json data in postgres](https://scalegrid.io/blog/using-jsonb-in-postgresql-how-to-effectively-store-index-json-data-in-postgresql/)
