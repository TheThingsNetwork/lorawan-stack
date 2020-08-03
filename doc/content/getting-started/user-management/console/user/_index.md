---
title: "User Management"
description: ""
---

This section contains instructions for managing users with the CLI.

<!--more-->

User management can be found in the user dropdown in the top right corner of the Console.

{{< figure src="dropdown.png" alt="User Dropdown" >}}

## Listing Users

The list of users is shown immediately after going to **User Management** in the user dropdown.

{{< figure src="users-list.png" alt="User List" >}}

## Searching for Users

You can search for users by ID using the search field above the list of users. It is currently not possible to search for users by other fields than the user ID using the Console, but you can [do this with the CLI]({{< relref "../../cli/user#searching-for-users" >}}).

## Creating Users {{% new-in-version "3.9" %}}

To create a user, click the **Add user** button in the top right of the user management page.

{{< figure src="users-list.png" alt="User List" >}}

You'll be taken to a page where you can enter the new user's information.

{{< figure src="users-add.png" alt="User Add" >}}

After entering all of the user information, click **Add user** at the bottom to create the new user.

## Inviting Users

It is currently not possible to invite users from the Console, but you can [do this with the CLI]({{< relref "../../cli/user#inviting-users" >}}).

## Updating Users

In order to update a user, select that user from the list. You'll now see the edit view.

{{< figure src="users-edit.png" alt="Editing a User" >}}

After making the changes to the user, click **Save Changes** to update the user.

## Deleting Users

In the bottom of the edit view, you can click **Delete User** to delete the user.

{{< figure src="users-delete.png" alt="Deleting a User" >}}

> **NOTE:** When deleting users, their user IDs stay reserved in the system, it is not possible to create a new user with the same user ID. In most cases you'll probably want to update a user to set its state to "suspended" instead.
