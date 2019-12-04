---
title: "Available Templates"
description: ""
weight: 3
---

The following email templates are currently used by {{% tts %}}:

Full Name | Identifier | Scope | Additional fields
----------|------------|-------|------------------
Invitation | `invitation` | Sent when inviting new users to the network. | `InvitationToken`
API Key changed | `api_key_changed` | Sent when the rights of an API Key have been changed. | `Identifiers` and `Rights`
API Key created | `api_key_created` | Send when an API Key has been created. | `Identifiers` and `Rights`
Collaborator changed | `collaborator_changed` | Sent when the rights of a collaborator have been changed. | `Collaborator`
Password changed | `password_changed` | Sent when the the password of a user has been changed.
Temporary password | `temporary_password` | Sent when a temporary password has been requested for an user. | `TemporaryPassword`
Email validation | `validate` | Sent when a user is added as a collaborator of an entity, in order to validate their email. | `ID` and `Token`
Entity State Changed | `entity_state_changed` | Sent when the approval state of an entity changed. | `State`

The following fields can be used inside all of the email templates:

Field | Description | Notes
------|-------------|------
`User.ID` | The identifier of the user. | If the recipient is the contact of the target entity, the value will be empty.
`User.Name` | The full name of the user. | If the recipient is the contact of the target entity, the value will be `user`.
`User.Email` | The email of the user. | None.
`Network.Name` | The full name of the network running on the stack. | None.
`Network.IdentityServerURL` | The base URL of the Identity Server. | None.
`Network.ConsoleURL` | The base URL of the Console. | None.
`Entity.Type` | The type of the target entity, such as `application`, `client`, `gateway`, `organization` or `user`. | None.
`Enitity.ID` | The identifier of the target entity. | None.
`Contact.Type` | The contact type of the recipient, such as `technical`, `billing` or `abuse`. | Only used if the recipient is the contact of an entity.
