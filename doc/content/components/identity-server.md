---
title: "Identity Server"
description: ""
weight: 1
---

The Identity Server provides the registries that store entities such as applications with their end devices, gateways, users, organizations and OAuth clients. It also manages access control through memberships and API keys.

<!--more-->

## Entity Registries

The entity registries store common information about all major entities in {{% tts %}}. This includes a name, description, and attributes (user-defined key-value pairs).

### Users

The first entity registered in the Identity Server is usually the admin user. Users can register by providing an email address and choosing a user ID and password. This information can later be used to login.

The user ID is the unique identifier of a user. User IDs are in the same namespace as organization IDs. This means that it is not possible to create a user with the same ID as an existing organization. For security reasons it is also not possible to re-use the ID of a deleted user or organization.

Users can be "admin", which gives them elevated privileges. The normal registration process does not allow users to register as admin, which is why the [Getting Started guide]({{< ref "/guides/getting-started" >}}) creates the admin user with a different command.

Users can be in one of multiple states: requested, approved, rejected, suspended, etc. The state of a user determines if the user is able to perform actions, and which actions. Normally, users are in the "approved" state. If {{% tts %}} is configured to require admin approval for new users, users are initially in the "requested" state, and an admin user can update them to the "approved" or "rejected" state. Users can also be suspended by admins if they misbehave.

### Gateways

Gateways can be registered by choosing an ID and optionally registering the EUI of the gateway. After registration, an API key can be created; the gateway can use its ID (or EUI) together with that API key to authenticate with the {{% tts %}}.

For correct operation of the gateway, it is important to set the frequency plan ID and set whether the gateway needs to be compliant with a duty cycle.

If the gateway is capable of communicating with a [Gateway Configuration Server]({{< relref "gateway-configuration-server.md" >}}), the Gateway Server address, and firmware update settings can be set in the gateway registry.

The gateway registry also stores information about the antenna(s) of the gateway, such as location and gain.

For more details, see the the [gateway API reference]({{< ref "/reference/api/gateway.md" >}}).

### Applications

Applications are used to organize registrations and traffic of multiple end devices in once place. An application typically corresponds to a collection of end devices that are in the same deployment, or of the same type.

For more details, see the the [application API reference]({{< ref "/reference/api/application.md" >}}).

### End Devices

The end device registry in the Identity Server stores only metadata of end devices, allowing clients such as the Console and CLI to list end devices in an application. It typically stores metadata about the brand, model, hardware, firmware and location of the end device. It also stores addresses of the Network Server, Application Server and Join Server, so that clients such as the Console and CLI know where other properties of the end device are stored.

For more details, see the the [end device API reference]({{< ref "/reference/api/end_device.md" >}}).

### Organizations

Organizations are used to create groups of multiple users, and easily assign rights to the entire group of users.

The organization ID is the unique identifier of a organization. Organization IDs are in the same namespace as user IDs. This means that it is not possible to create an organization with the same ID as an existing user. For security reasons it is also not possible to re-use the ID of a deleted user or organization.

### OAuth Clients

It is possible to register external OAuth clients in the OAuth registries. OAuth clients are registered with a client ID and secret.

As with users, OAuth clients can be in one of multiple states. OAuth clients created by non-admin users need to be approved by an admin user.

Official OAuth clients can be marked as "endorsed", or can be pre-authorized for all users.

## Entity Access

The Identity Server is responsible for access control to entities.

### Memberships

Memberships define the rights that a user or organization has on another entity. The simplest membership is one where a user is a direct member (or collaborator) of an application or gateway. The rights of a membership indicate what the user is allowed to do with the application or gateway.

An indirect membership means that a user is a member of an organization, and the organization is a member of an application or gateway. In this case, the rights of the user-organization membership and the organization-application or organization-gateway membership are intersected in order to compute the effective rights of the user.

### API Keys

With most entities it is possible to create API keys. These API keys allow you to call APIs on behalf of the entity. API keys do not expire, but it is possible to revoke an API key.

API keys have associated rights. This means that it is possible to create for instance an _Application API key_ that can be used to read information about end devices, but not see the root keys, and not make changes.

It is possible to combine memberships and API keys, so you can for instance create an _Organization API key_ with the right to list the applications where the organization is a member, and the right to read information about end devices in these applications.

## OAuth

The Identity Server is an OAuth 2.0 server. Users can authorize OAuth clients to access their data, manage authorizations, and even manage individual access tokens.
