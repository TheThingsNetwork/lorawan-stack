---
title: "User Management with the CLI"
description: ""
weight: 2
---

## Listing Users

To list users with the CLI, use the `users list` command. Make sure to specify the fields you're interested in.

```bash
$ ttn-lw-cli users list --name --state --admin
```

> **TIP:** Use the pagination flags `--limit` and `--page` when there are many users.

<details><summary>Show output</summary>
```
[{
  "ids": {
    "user_id": "new-user"
  },
  "created_at": "2019-12-19T09:10:31.426Z",
  "updated_at": "2019-12-19T09:10:40.527Z",
  "name": "New User"
}, {
  "ids": {
    "user_id": "admin"
  },
  "created_at": "2019-12-18T14:54:12.723Z",
  "updated_at": "2019-12-18T14:54:12.723Z",
  "state": "STATE_APPROVED",
  "admin": true
}]
```
</details>

## Searching for Users

To search for users with the CLI, use the `users search` command. Make sure to specify the fields you're interested in. We'll search for users with IDs that contain "new":

```bash
$ ttn-lw-cli users search --id-contains new --name
```

> **TIP:** Use the pagination flags `--limit` and `--page` when there are many users.

<details><summary>Show output</summary>
```
[{
  "ids": {
    "user_id": "new-user"
  },
  "created_at": "2019-12-19T09:10:31.426Z",
  "updated_at": "2019-12-19T09:10:40.527Z",
  "name": "New User"
}]
```
</details>

## Creating Users

Network Administrators can create user accounts as follows:

```bash
$ ttn-lw-cli users create colleague \
  --name "My Colleague" \
  --primary-email-address colleague@thethings.network
```

<details><summary>Show output</summary>
```
Please enter password:***************
Please confirm password:***************
{
  "ids": {
    "user_id": "colleague"
  },
  "created_at": "2019-12-19T10:54:53.677Z",
  "updated_at": "2019-12-19T10:54:53.677Z",
  "name": "My Colleague",
  "contact_info": [
    {
      "contact_method": "CONTACT_METHOD_EMAIL",
      "value": "colleague@thethings.network"
    }
  ],
  "primary_email_address": "colleague@thethings.network",
  "password_updated_at": "2019-12-19T10:54:53.674Z",
  "state": "STATE_APPROVED"
}
```
</details>

## Inviting Users

You can create invitations for users to join the network with the `users invitations create` command:

```bash
$ ttn-lw-cli users invitations create colleague@thethings.network
```

After you do this, you'll be able to list the invitations you've sent:

```bash
% ttn-lw-cli users invitations list
```

<details><summary>Show output</summary>
```
[{
  "email": "colleague@thethings.network",
  "token": "MW7INQWYOE46GLP3AEFQEHR5XIKRYPSRAXFF3CUCLIQPPQ3BNBLQ",
  "expires_at": "2019-12-26T11:41:29.485Z",
  "created_at": "2019-12-19T11:41:29.486Z",
  "updated_at": "2019-12-19T11:41:29.486Z"
}]
```
</details>

And delete an invitation if you want to revoke it:

```bash
$ ttn-lw-cli users invitations delete colleague@thethings.network
```

## Updating Users

To update users with the CLI, use the `users update` command. The following command updates the state of user `new-user` to "approved" and makes them admin of the network:

```bash
$ ttn-lw-cli users update new-user --state APPROVED --admin true
```

<details><summary>Show output</summary>
```
{
  "ids": {
    "user_id": "new-user"
  },
  "created_at": "2019-12-19T09:10:31.426Z",
  "updated_at": "2019-12-19T11:44:39.609Z",
  "state": "STATE_APPROVED",
  "admin": true
}
```
</details>

## Deleting Users

To delete a user, use the `users delete` command.

> **NOTE:** When deleting users, their user IDs stay reserved in the system, it is not possible to create a new user with the same user ID. In most cases you'll probably want to update a user to set its state to "suspended" instead.
