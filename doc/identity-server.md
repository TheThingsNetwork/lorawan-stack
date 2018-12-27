# Identity Server

The **Identity Server** component of The Things Network Stack implements functionality for authentication and access control.

## Users

## Applications

## Gateways

## OAuth clients

## Data Migrations

###  Updating the rights list

The Identity Server is not automatically aware of updates to the rights list. If the rights list is updated without reflecting in the Identity Server, inconsistencies such as newly introduced rights not being granted to any user or un-parseable stale rights being stored in the database can make applications unmanageable.

In such cases, data migrations should be written, and/or executed. A *data migration* is the process of applying a set of modifications to the existing dataset.

Currently the Identity Server has a sub-package called `migrations` where these data migrations must be placed.

Data migrations can currently:
 - Modify the rights (i.e. scope) of a third-party client
 - Modify the rights of Application and Gateway collaborations
 - Modify the rights of Organization Memberships
 - Modify the rights of API keys


Please refer to godoc for example migrations.
