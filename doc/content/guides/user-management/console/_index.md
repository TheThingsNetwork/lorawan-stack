---
title: "User Management with the Console"
description: ""
weight: 1
---

User management can be found in the user dropdown in the top right corner of the Console.

{{< figure src="dropdown.png" alt="User Dropdown" >}}

## Listing Users

The list of users is shown immediately after going to **User Management** in the user dropdown.

{{< figure src="users-list.png" alt="User List" >}}

## Searching for Users

You can search for users by ID using the search field above the list of users. It is currently not possible to search for users by other fields than the user ID using the Console, but you can [do this with the CLI]({{< relref "../cli#searching-for-users" >}}).

## Creating Users

It is currently not possible to create users in the Console, but users can register themselves, or you can [create them with the CLI]({{< relref "../cli#creating-users" >}}).

## Inviting Users

It is currently not possible to invite users from the Console, but you can [do this with the CLI]({{< relref "../cli#inviting-users" >}}).

## Updating Users

In order to update a user, select that user from the list. You'll now see the edit view.

{{< figure src="users-edit.png" alt="Editing a User" >}}

After making the changes to the user, click **Save Changes** to update the user.

## Deleting Users

In the bottom of the edit view, you can click **Delete User** to delete the user.

{{< figure src="users-delete.png" alt="Deleting a User" >}}

> **NOTE:** When deleting users, their user IDs stay reserved in the system, it is not possible to create a new user with the same user ID. In most cases you'll probably want to update a user to set its state to "suspended" instead.
