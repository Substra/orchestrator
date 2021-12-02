# Permissions

Some assets have permissions: they state which actor of the network can act on those assets.

Permissions are either set during asset registration or inherited from dependent assets.

## Structure

Permissions are for several actions:

- Process: the ability to the use the asset in a node on the platform without downloading it (eg: processing of a datasample by a compute task)
- Download: possibility for a user of a node to download the asset

Each action has the same structure:

- public: a boolean flag, if true there is no restriction on the action.
- authorized_ids: a list of node IDs allowed to perform the action

## Asset permissions

The following assets have their permissions set by their owner at registration:

- algo
- metric
- [datamanager](./assets/datamanager.md)

If a permission is not `public` and `authorized_ids` does not contain the creator's node ID, it will be added.
This is to make sure the creator of an asset cannot be "locked out".

Compute tasks bears permissions related to their output models.
These are generally computed during registration, but follow specific rules depending on the task kind.

**Train Task**: model permissions are the [intersection](#intersection) of algo and datamanager permissions

**Composite Task**: this one is more complex since there are two output models.
The *Simple* model receives its permissions from the task input (i.e. set by the creator of the task), and is owned by the datamanager's owner.
The *Head* model is restricted to the datamanager's owner only.

**Aggregate Task**: permissions of the output model is the [union](#union) of the permissions of the parent models (only *Simple* model is considered for composite parents).


## Permissions operations

Permissions can be combined with the following operations.

### Intersection

This is the logical operation `A∧B`.
To match the resulting permissions, one has to satisfy both A **and** B permissions.

| A                                    | B                                          | A∧B                                  |
|--------------------------------------|--------------------------------------------|--------------------------------------|
| Public: true                         | Public: false, AuthorizedIds: [test]       | Public: false, Authorizedids: [test] |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org2]       | Public: false, AuthorizedIds: []     |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org1, org2] | Public: false, AuthorizedIds: [org1] |

### Union

This is the logical operation `A∨B`
To match the resulting permissions, one has to satisfy A **or** B permissions.

| A                                    | B                                          | A∨B                                        |
|--------------------------------------|--------------------------------------------|--------------------------------------------|
| Public: true                         | Public: false, AuthorizedIds: [test]       | Public: true                               |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org2]       | Public: false, AuthorizedIds: [org1, org2] |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org1, org2] | Public: false, AuthorizedIds: [org1, org2] |
