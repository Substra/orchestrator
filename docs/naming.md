# Naming conventions

Asset retrival and querying methods should follow these patterns:

- GetXXX should take a key and return a single entity
- GetXXXs should take a slice of keys and return a slice of entities
- GetAllXXXs should take no arguments and return a slice of entities
- QueryXXX should take pagination param and optionally a filter, and return a response with entities (slice) + pagination

This should be the case in both proto and services.

## Protobuf inputs & outputs

Some RPC methods receive specific (ie: not an asset) input and output.
In those cases, the following convention should be used: the input should be the name of the function suffixed by `Param`, the output suffixed by `Response`.

Example: the function named `QueryComputeTask` takes a `QueryComputeTaskParam` and returns a `QueryComputeTaskResponse`.