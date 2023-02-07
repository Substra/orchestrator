# Permissions

Some assets have permissions: they state which actor of the network can act on those assets.

Permissions are either set during asset registration or inherited from dependent assets.

## Structure

Permissions are for several actions:

- Process: the ability to the use the asset in an organization on the platform without downloading it (eg: processing of a datasample by a compute task)
- Download: possibility for a user of an organization to download the asset

Each action has the same structure:

- public: a boolean flag, if true there is no restriction on the action.
- authorized_ids: a list of organization IDs allowed to perform the action

## Asset permissions

The following assets have their permissions set by their owner at registration:

- function
- metric
- [datamanager](./assets/datamanager.md)

If a permission is not `public` and `authorized_ids` does not contain the creator's organization ID, it will be added.
This is to make sure the creator of an asset cannot be "locked out".

Compute tasks bears permissions related to their output models.
These are generally computed during registration, but follow specific rules depending on the task kind.

**Train Task**: model permissions are the [intersection](#intersection) of function and datamanager permissions

**Composite Task**: this one is more complex since there are two output models.
The _Simple_ model receives its permissions from the task input (i.e. set by the creator of the task), and is owned by the datamanager's owner.
The _Head_ model is restricted to the datamanager's owner only.

**Aggregate Task**: permissions of the output model is the [union](#union) of the permissions of the parent models (only _Simple_ model is considered for composite parents).

## Permissions operations

Permissions can be combined with the following operations.

### Intersection

This is the logical operation `A∧B`.
To match the resulting permissions, one has to satisfy both A **and** B permissions.

| A                                    | B                                          | A∧B                                  |
| ------------------------------------ | ------------------------------------------ | ------------------------------------ |
| Public: true                         | Public: false, AuthorizedIds: [test]       | Public: false, AuthorizedIds: [test] |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org2]       | Public: false, AuthorizedIds: []     |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org1, org2] | Public: false, AuthorizedIds: [org1] |

### Union

This is the logical operation `A∨B`
To match the resulting permissions, one has to satisfy A **or** B permissions.

| A                                    | B                                          | A∨B                                        |
| ------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| Public: true                         | Public: false, AuthorizedIds: [test]       | Public: true                               |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org2]       | Public: false, AuthorizedIds: [org1, org2] |
| Public: false, AuthorizedIds: [org1] | Public: false, AuthorizedIds: [org1, org2] | Public: false, AuthorizedIds: [org1, org2] |
