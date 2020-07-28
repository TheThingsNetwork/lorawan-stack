---
title: "Organization Management"
description: ""
aliases: [/getting-started/user-management/cli/org, /getting-started/user-management/console/org]
---

This section contains instructions for managing organizations.

<!--more-->

{{< tabs/container "Console" "CLI" >}}

{{< tabs/tab "Console" >}}

## Managing Organizations using the Console

To manage organizations, click the **Organizations** tab in the top menu.

{{< figure src="orgs.png" alt="Organizations" >}}

To add an organization, click **Add organization**.

{{< figure src="add-org.png" alt="Add organization" >}}

Choose the rights you would like to grant the organization, and click the **Add organization** button to save your choices.

>Note: When a user is a member of an organization which is a collaborator for an entity, the user's rights are the intersection of the user's rights in the organization and the organization's rights on the entity.

{{< /tabs/tab >}}

{{< tabs/tab "CLI" >}}

## Managing Organizations using the CLI

## Creating Organizations

Administrators can create organizations as follows:

```bash
$ ttn-lw-cli organizations create org1 --user-id user1
```

This will create an organization `org1` with all the rights of `user1` and make `user1` a collaborator within the organization.

Output:

```json
{
  "ids": {
    "organization_id": "org1"
  },
  "created_at": "2020-07-14T09:37:01.938Z",
  "updated_at": "2020-07-14T09:37:01.938Z"
}
```

## Listing Organizations

To list organizations with the CLI, use the `organizations list` command.

```bash
$ ttn-lw-cli organizations list
```

```json
[{
  "ids": {
    "organization_id": "org1"
  },
  "created_at": "2020-07-09T12:39:35.129Z",
  "updated_at": "2020-07-09T12:39:35.129Z"
}
, {
  "ids": {
    "organization_id": "org2"
  },
  "created_at": "2020-07-14T09:37:01.938Z",
  "updated_at": "2020-07-14T09:37:01.938Z"
}]
```

## Searching for Organizations

To search for organizations with the CLI, use the `organizations search` command. Make sure to specify the fields you're interested in. This example will search for organizations with IDs that contain "org1":

```bash
$ ttn-lw-cli organizations search --id-contains org1
```

Output:

```json
[{
  "ids": {
    "organization_id": "org1"
  },
  "created_at": "2020-07-09T12:39:35.129Z",
  "updated_at": "2020-07-09T12:39:35.129Z"
}]
```

## Adding Users to Organizations

To add a user to an organization, use the  `organizations collaborators set` command. This will add user `user1` as a collaborator of organization `org1` with all organization rights:

```bash
$ ttn-lw-cli organizations collaborators set --organization-id org1 --user-id user1 --right-organization-all
```

>Note: You must specify rights when adding a collaborator. Use the `--help` flag to see the list of possible rights, e.g `$ ttn-lw-cli organizations collaborators set --help`.

## Removing Users from Organizations

To remove a user from an organization, use the  `organizations collaborators delete` command:

```bash
$ ttn-lw-cli organizations collaborators delete --organization-id org1 --user-id user1
```

This will remove user `user1` as a collaborator of organization `org1`

## Deleting Organizations

To delete an organization, use the `organizations delete` command.

```bash
$ ttn-lw-cli organizations delete --organization-id org1
```

> **NOTE:** When deleting organizations, their IDs stay reserved in the system. For security reasons, it is not possible to create a new organization with the same ID.

{{< /tabs/tab >}}

{{< /tabs/container >}}
