---
title: "Services"
description: "List of services and methods available through the API"
weight: 1
tags: [ http, grpc ]
---



Name | Description
---|---
[ApplicationAccess](#ApplicationAccess) | 
[ApplicationRegistry](#ApplicationRegistry) | ApplicationRegistry is used to managed application
[AppAs](#AppAs) | The AppAs service connects an application or integration to an Application Server.
[As](#As) | The As service manages the Application Server.
[AsEndDeviceRegistry](#AsEndDeviceRegistry) | The AsEndDeviceRegistry service allows clients to manage their end devices on the Application Server.
[ApplicationPubSubRegistry](#ApplicationPubSubRegistry) | 
[ApplicationWebhookRegistry](#ApplicationWebhookRegistry) | 
[ClientAccess](#ClientAccess) | 
[ClientRegistry](#ClientRegistry) | 
[Configuration](#Configuration) | 
[ContactInfoRegistry](#ContactInfoRegistry) | 
[EndDeviceRegistry](#EndDeviceRegistry) | 
[Events](#Events) | The Events service serves events from the cluster.
[GatewayAccess](#GatewayAccess) | 
[GatewayConfigurator](#GatewayConfigurator) | 
[GatewayRegistry](#GatewayRegistry) | 
[Gs](#Gs) | 
[GtwGs](#GtwGs) | The GtwGs service connects a gateway to a Gateway Server.
[NsGs](#NsGs) | The NsGs service connects a Network Server to a Gateway Server.
[EntityAccess](#EntityAccess) | 
[ApplicationCryptoService](#ApplicationCryptoService) | Service for application layer cryptographic operations.
[AsJs](#AsJs) | The AsJs service connects an Application Server to a Join Server.
[Js](#Js) | 
[JsEndDeviceRegistry](#JsEndDeviceRegistry) | The JsEndDeviceRegistry service allows clients to manage their end devices on the Join Server.
[NetworkCryptoService](#NetworkCryptoService) | Service for network layer cryptographic operations.
[NsJs](#NsJs) | The NsJs service connects a Network Server to a Join Server.
[DownlinkMessageProcessor](#DownlinkMessageProcessor) | The DownlinkMessageProcessor service processes downlink messages.
[UplinkMessageProcessor](#UplinkMessageProcessor) | The UplinkMessageProcessor service processes uplink messages.
[AsNs](#AsNs) | The AsNs service connects an Application Server to a Network Server.
[GsNs](#GsNs) | The GsNs service connects a Gateway Server to a Network Server.
[Ns](#Ns) | 
[NsEndDeviceRegistry](#NsEndDeviceRegistry) | The NsEndDeviceRegistry service allows clients to manage their end devices on the Network Server.
[OAuthAuthorizationRegistry](#OAuthAuthorizationRegistry) | 
[OrganizationAccess](#OrganizationAccess) | 
[OrganizationRegistry](#OrganizationRegistry) | 
[EndDeviceRegistrySearch](#EndDeviceRegistrySearch) | The EndDeviceRegistrySearch service indexes devices in the EndDeviceRegistry and enables searching for them. This service is not implemented on all deployments.
[EntityRegistrySearch](#EntityRegistrySearch) | The EntityRegistrySearch service indexes entities in the various registries and enables searching for them. This service is not implemented on all deployments.
[UserAccess](#UserAccess) | 
[UserInvitationRegistry](#UserInvitationRegistry) | 
[UserRegistry](#UserRegistry) | 
[UserSessionRegistry](#UserSessionRegistry) | 
[Scalar Value Types](#scalar-value-types)
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="ApplicationAccess">ApplicationAccess</a>
  `lorawan-stack/api/application_services.proto`

  

  
### <a name="ListRights">ListRights</a>
  

  {{% reftab ListRights gRPCListRights HTTPListRights %}}

  **Request**: [ApplicationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ApplicationIdentifiers" >}})

  **Response**: [Rights]({{< ref "messages.md#ttn.lorawan.v3.Rights" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{application_id}/rights` |  |{{% /reftab %}}

  
### <a name="CreateAPIKey">CreateAPIKey</a>
  

  {{% reftab CreateAPIKey gRPCCreateAPIKey HTTPCreateAPIKey %}}

  **Request**: [CreateApplicationAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateApplicationAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/applications/{application_ids.application_id}/api-keys` | * |{{% /reftab %}}

  
### <a name="ListAPIKeys">ListAPIKeys</a>
  

  {{% reftab ListAPIKeys gRPCListAPIKeys HTTPListAPIKeys %}}

  **Request**: [ListApplicationAPIKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.ListApplicationAPIKeysRequest" >}})

  **Response**: [APIKeys]({{< ref "messages.md#ttn.lorawan.v3.APIKeys" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{application_ids.application_id}/api-keys` |  |{{% /reftab %}}

  
### <a name="GetAPIKey">GetAPIKey</a>
  

  {{% reftab GetAPIKey gRPCGetAPIKey HTTPGetAPIKey %}}

  **Request**: [GetApplicationAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.GetApplicationAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{application_ids.application_id}/api-keys/{key_id}` |  |{{% /reftab %}}

  
### <a name="UpdateAPIKey">UpdateAPIKey</a>
  Update the rights of an existing application API key. To generate an API key,
the CreateAPIKey should be used. To delete an API key, update it
with zero rights.

  {{% reftab UpdateAPIKey gRPCUpdateAPIKey HTTPUpdateAPIKey %}}

  **Request**: [UpdateApplicationAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/applications/{application_ids.application_id}/api-keys/{api_key.id}` | * |{{% /reftab %}}

  
### <a name="GetCollaborator">GetCollaborator</a>
  Get the rights of a collaborator (member) of the application.
Pseudo-rights in the response (such as the "_ALL" right) are not expanded.

  {{% reftab GetCollaborator gRPCGetCollaborator HTTPGetCollaborator %}}

  **Request**: [GetApplicationCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.GetApplicationCollaboratorRequest" >}})

  **Response**: [GetCollaboratorResponse]({{< ref "messages.md#ttn.lorawan.v3.GetCollaboratorResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{application_ids.application_id}/collaborator` |  |
 `GET` | `/api/v3/applications/{application_ids.application_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
 `GET` | `/api/v3/applications/{application_ids.application_id}/collaborator/organization/{collaborator.organization_ids.organization_id}` |  |{{% /reftab %}}

  
### <a name="SetCollaborator">SetCollaborator</a>
  Set the rights of a collaborator (member) on the application.
Setting a collaborator without rights, removes them.

  {{% reftab SetCollaborator gRPCSetCollaborator HTTPSetCollaborator %}}

  **Request**: [SetApplicationCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.SetApplicationCollaboratorRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/applications/{application_ids.application_id}/collaborators` | * |{{% /reftab %}}

  
### <a name="ListCollaborators">ListCollaborators</a>
  

  {{% reftab ListCollaborators gRPCListCollaborators HTTPListCollaborators %}}

  **Request**: [ListApplicationCollaboratorsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListApplicationCollaboratorsRequest" >}})

  **Response**: [Collaborators]({{< ref "messages.md#ttn.lorawan.v3.Collaborators" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{application_ids.application_id}/collaborators` |  |{{% /reftab %}}

  
  

## <a name="ApplicationRegistry">ApplicationRegistry</a>
  `lorawan-stack/api/application_services.proto`

  ApplicationRegistry is used to managed application

  
### <a name="Create">Create</a>
  Create a new application. This also sets the given organization or user as
first collaborator with all possible rights.

  {{% reftab Create gRPCCreate HTTPCreate %}}

  **Request**: [CreateApplicationRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateApplicationRequest" >}})

  **Response**: [Application]({{< ref "messages.md#ttn.lorawan.v3.Application" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/applications` | * |
 `POST` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/applications` | * |{{% /reftab %}}

  
### <a name="Get">Get</a>
  Get the application with the given identifiers, selecting the fields given
by the field mask. The method may return more or less fields, depending on
the rights of the caller.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetApplicationRequest]({{< ref "messages.md#ttn.lorawan.v3.GetApplicationRequest" >}})

  **Response**: [Application]({{< ref "messages.md#ttn.lorawan.v3.Application" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{application_ids.application_id}` |  |{{% /reftab %}}

  
### <a name="List">List</a>
  List applications. See request message for details.

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListApplicationsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListApplicationsRequest" >}})

  **Response**: [Applications]({{< ref "messages.md#ttn.lorawan.v3.Applications" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications` |  |
 `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/applications` |  |
 `GET` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/applications` |  |{{% /reftab %}}

  
### <a name="Update">Update</a>
  

  {{% reftab Update gRPCUpdate HTTPUpdate %}}

  **Request**: [UpdateApplicationRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateApplicationRequest" >}})

  **Response**: [Application]({{< ref "messages.md#ttn.lorawan.v3.Application" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/applications/{application.ids.application_id}` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [ApplicationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ApplicationIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/applications/{application_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  

## <a name="AppAs">AppAs</a>
  `lorawan-stack/api/applicationserver.proto`

  The AppAs service connects an application or integration to an Application Server.

  
### <a name="Subscribe">Subscribe</a>
  

  {{% reftab Subscribe gRPCSubscribe HTTPSubscribe %}}

  **Request**: [ApplicationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ApplicationIdentifiers" >}})

  **Response**: [ApplicationUp]({{< ref "messages.md#ttn.lorawan.v3.ApplicationUp" >}}) _stream_

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="DownlinkQueuePush">DownlinkQueuePush</a>
  

  {{% reftab DownlinkQueuePush gRPCDownlinkQueuePush HTTPDownlinkQueuePush %}}

  **Request**: [DownlinkQueueRequest]({{< ref "messages.md#ttn.lorawan.v3.DownlinkQueueRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/down/push` | * |{{% /reftab %}}

  
### <a name="DownlinkQueueReplace">DownlinkQueueReplace</a>
  

  {{% reftab DownlinkQueueReplace gRPCDownlinkQueueReplace HTTPDownlinkQueueReplace %}}

  **Request**: [DownlinkQueueRequest]({{< ref "messages.md#ttn.lorawan.v3.DownlinkQueueRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/down/replace` | * |{{% /reftab %}}

  
### <a name="DownlinkQueueList">DownlinkQueueList</a>
  

  {{% reftab DownlinkQueueList gRPCDownlinkQueueList HTTPDownlinkQueueList %}}

  **Request**: [EndDeviceIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.EndDeviceIdentifiers" >}})

  **Response**: [ApplicationDownlinks]({{< ref "messages.md#ttn.lorawan.v3.ApplicationDownlinks" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/applications/{application_ids.application_id}/devices/{device_id}/down` |  |{{% /reftab %}}

  
  

## <a name="As">As</a>
  `lorawan-stack/api/applicationserver.proto`

  The As service manages the Application Server.

  
### <a name="GetLink">GetLink</a>
  

  {{% reftab GetLink gRPCGetLink HTTPGetLink %}}

  **Request**: [GetApplicationLinkRequest]({{< ref "messages.md#ttn.lorawan.v3.GetApplicationLinkRequest" >}})

  **Response**: [ApplicationLink]({{< ref "messages.md#ttn.lorawan.v3.ApplicationLink" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/applications/{application_ids.application_id}/link` |  |{{% /reftab %}}

  
### <a name="SetLink">SetLink</a>
  Set a link configuration from the Application Server a Network Server.
This call returns immediately after setting the link configuration; it does not wait for a link to establish.
To get link statistics or errors, use the `GetLinkStats` call.

  {{% reftab SetLink gRPCSetLink HTTPSetLink %}}

  **Request**: [SetApplicationLinkRequest]({{< ref "messages.md#ttn.lorawan.v3.SetApplicationLinkRequest" >}})

  **Response**: [ApplicationLink]({{< ref "messages.md#ttn.lorawan.v3.ApplicationLink" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/as/applications/{application_ids.application_id}/link` | * |{{% /reftab %}}

  
### <a name="DeleteLink">DeleteLink</a>
  

  {{% reftab DeleteLink gRPCDeleteLink HTTPDeleteLink %}}

  **Request**: [ApplicationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ApplicationIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/as/applications/{application_id}/link` |  |{{% /reftab %}}

  
### <a name="GetLinkStats">GetLinkStats</a>
  GetLinkStats returns the link statistics.
This call returns a NotFound error code if there is no link for the given application identifiers.
This call returns the error code of the link error if linking to a Network Server failed.

  {{% reftab GetLinkStats gRPCGetLinkStats HTTPGetLinkStats %}}

  **Request**: [ApplicationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ApplicationIdentifiers" >}})

  **Response**: [ApplicationLinkStats]({{< ref "messages.md#ttn.lorawan.v3.ApplicationLinkStats" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/applications/{application_id}/link/stats` |  |{{% /reftab %}}

  
  

## <a name="AsEndDeviceRegistry">AsEndDeviceRegistry</a>
  `lorawan-stack/api/applicationserver.proto`

  The AsEndDeviceRegistry service allows clients to manage their end devices on the Application Server.

  
### <a name="Get">Get</a>
  Get returns the device that matches the given identifiers.
If there are multiple matches, an error will be returned.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.GetEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |{{% /reftab %}}

  
### <a name="Set">Set</a>
  Set creates or updates the device.

  {{% reftab Set gRPCSet HTTPSet %}}

  **Request**: [SetEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.SetEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/as/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | * |
 `POST` | `/api/v3/as/applications/{end_device.ids.application_ids.application_id}/devices` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  Delete deletes the device that matches the given identifiers.
If there are multiple matches, an error will be returned.

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [EndDeviceIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.EndDeviceIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/as/applications/{application_ids.application_id}/devices/{device_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  

## <a name="ApplicationPubSubRegistry">ApplicationPubSubRegistry</a>
  `lorawan-stack/api/applicationserver_pubsub.proto`

  

  
### <a name="GetFormats">GetFormats</a>
  

  {{% reftab GetFormats gRPCGetFormats HTTPGetFormats %}}

  **Request**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  **Response**: [ApplicationPubSubFormats]({{< ref "messages.md#ttn.lorawan.v3.ApplicationPubSubFormats" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/pubsub-formats` |  |{{% /reftab %}}

  
### <a name="Get">Get</a>
  

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetApplicationPubSubRequest]({{< ref "messages.md#ttn.lorawan.v3.GetApplicationPubSubRequest" >}})

  **Response**: [ApplicationPubSub]({{< ref "messages.md#ttn.lorawan.v3.ApplicationPubSub" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/pubsub/{ids.application_ids.application_id}/{ids.pub_sub_id}` |  |{{% /reftab %}}

  
### <a name="List">List</a>
  

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListApplicationPubSubsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListApplicationPubSubsRequest" >}})

  **Response**: [ApplicationPubSubs]({{< ref "messages.md#ttn.lorawan.v3.ApplicationPubSubs" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/pubsub/{application_ids.application_id}` |  |{{% /reftab %}}

  
### <a name="Set">Set</a>
  

  {{% reftab Set gRPCSet HTTPSet %}}

  **Request**: [SetApplicationPubSubRequest]({{< ref "messages.md#ttn.lorawan.v3.SetApplicationPubSubRequest" >}})

  **Response**: [ApplicationPubSub]({{< ref "messages.md#ttn.lorawan.v3.ApplicationPubSub" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/as/pubsub/{pubsub.ids.application_ids.application_id}/{pubsub.ids.pub_sub_id}` | * |
 `POST` | `/api/v3/as/pubsub/{pubsub.ids.application_ids.application_id}` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [ApplicationPubSubIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ApplicationPubSubIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/as/pubsub/{application_ids.application_id}/{pub_sub_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  

## <a name="ApplicationWebhookRegistry">ApplicationWebhookRegistry</a>
  `lorawan-stack/api/applicationserver_web.proto`

  

  
### <a name="GetFormats">GetFormats</a>
  

  {{% reftab GetFormats gRPCGetFormats HTTPGetFormats %}}

  **Request**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  **Response**: [ApplicationWebhookFormats]({{< ref "messages.md#ttn.lorawan.v3.ApplicationWebhookFormats" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/webhook-formats` |  |{{% /reftab %}}

  
### <a name="Get">Get</a>
  

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetApplicationWebhookRequest]({{< ref "messages.md#ttn.lorawan.v3.GetApplicationWebhookRequest" >}})

  **Response**: [ApplicationWebhook]({{< ref "messages.md#ttn.lorawan.v3.ApplicationWebhook" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/webhooks/{ids.application_ids.application_id}/{ids.webhook_id}` |  |{{% /reftab %}}

  
### <a name="List">List</a>
  

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListApplicationWebhooksRequest]({{< ref "messages.md#ttn.lorawan.v3.ListApplicationWebhooksRequest" >}})

  **Response**: [ApplicationWebhooks]({{< ref "messages.md#ttn.lorawan.v3.ApplicationWebhooks" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/as/webhooks/{application_ids.application_id}` |  |{{% /reftab %}}

  
### <a name="Set">Set</a>
  

  {{% reftab Set gRPCSet HTTPSet %}}

  **Request**: [SetApplicationWebhookRequest]({{< ref "messages.md#ttn.lorawan.v3.SetApplicationWebhookRequest" >}})

  **Response**: [ApplicationWebhook]({{< ref "messages.md#ttn.lorawan.v3.ApplicationWebhook" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/as/webhooks/{webhook.ids.application_ids.application_id}/{webhook.ids.webhook_id}` | * |
 `POST` | `/api/v3/as/webhooks/{webhook.ids.application_ids.application_id}` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [ApplicationWebhookIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ApplicationWebhookIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/as/webhooks/{application_ids.application_id}/{webhook_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="ClientAccess">ClientAccess</a>
  `lorawan-stack/api/client_services.proto`

  

  
### <a name="ListRights">ListRights</a>
  

  {{% reftab ListRights gRPCListRights HTTPListRights %}}

  **Request**: [ClientIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ClientIdentifiers" >}})

  **Response**: [Rights]({{< ref "messages.md#ttn.lorawan.v3.Rights" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/clients/{client_id}/rights` |  |{{% /reftab %}}

  
### <a name="GetCollaborator">GetCollaborator</a>
  Get the rights of a collaborator (member) of the client.
Pseudo-rights in the response (such as the "_ALL" right) are not expanded.

  {{% reftab GetCollaborator gRPCGetCollaborator HTTPGetCollaborator %}}

  **Request**: [GetClientCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.GetClientCollaboratorRequest" >}})

  **Response**: [GetCollaboratorResponse]({{< ref "messages.md#ttn.lorawan.v3.GetCollaboratorResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/clients/{client_ids.client_id}/collaborator` |  |
 `GET` | `/api/v3/clients/{client_ids.client_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
 `GET` | `/api/v3/clients/{client_ids.client_id}/collaborator/organization/{collaborator.organization_ids.organization_id}` |  |{{% /reftab %}}

  
### <a name="SetCollaborator">SetCollaborator</a>
  Set the rights of a collaborator (member) on the client.
Setting a collaborator without rights, removes them.

  {{% reftab SetCollaborator gRPCSetCollaborator HTTPSetCollaborator %}}

  **Request**: [SetClientCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.SetClientCollaboratorRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/clients/{client_ids.client_id}/collaborators` | * |{{% /reftab %}}

  
### <a name="ListCollaborators">ListCollaborators</a>
  

  {{% reftab ListCollaborators gRPCListCollaborators HTTPListCollaborators %}}

  **Request**: [ListClientCollaboratorsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListClientCollaboratorsRequest" >}})

  **Response**: [Collaborators]({{< ref "messages.md#ttn.lorawan.v3.Collaborators" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/clients/{client_ids.client_id}/collaborators` |  |{{% /reftab %}}

  
  

## <a name="ClientRegistry">ClientRegistry</a>
  `lorawan-stack/api/client_services.proto`

  

  
### <a name="Create">Create</a>
  Create a new OAuth client. This also sets the given organization or user as
first collaborator with all possible rights.

  {{% reftab Create gRPCCreate HTTPCreate %}}

  **Request**: [CreateClientRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateClientRequest" >}})

  **Response**: [Client]({{< ref "messages.md#ttn.lorawan.v3.Client" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/clients` | * |
 `POST` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/clients` | * |{{% /reftab %}}

  
### <a name="Get">Get</a>
  Get the OAuth client with the given identifiers, selecting the fields given
by the field mask. The method may return more or less fields, depending on
the rights of the caller.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetClientRequest]({{< ref "messages.md#ttn.lorawan.v3.GetClientRequest" >}})

  **Response**: [Client]({{< ref "messages.md#ttn.lorawan.v3.Client" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/clients/{client_ids.client_id}` |  |{{% /reftab %}}

  
### <a name="List">List</a>
  List OAuth clients. See request message for details.

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListClientsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListClientsRequest" >}})

  **Response**: [Clients]({{< ref "messages.md#ttn.lorawan.v3.Clients" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/clients` |  |
 `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/clients` |  |
 `GET` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/clients` |  |{{% /reftab %}}

  
### <a name="Update">Update</a>
  

  {{% reftab Update gRPCUpdate HTTPUpdate %}}

  **Request**: [UpdateClientRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateClientRequest" >}})

  **Response**: [Client]({{< ref "messages.md#ttn.lorawan.v3.Client" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/clients/{client.ids.client_id}` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [ClientIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.ClientIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/clients/{client_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="Configuration">Configuration</a>
  `lorawan-stack/api/configuration_services.proto`

  

  
### <a name="ListFrequencyPlans">ListFrequencyPlans</a>
  

  {{% reftab ListFrequencyPlans gRPCListFrequencyPlans HTTPListFrequencyPlans %}}

  **Request**: [ListFrequencyPlansRequest]({{< ref "messages.md#ttn.lorawan.v3.ListFrequencyPlansRequest" >}})

  **Response**: [ListFrequencyPlansResponse]({{< ref "messages.md#ttn.lorawan.v3.ListFrequencyPlansResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/configuration/frequency-plans` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  

## <a name="ContactInfoRegistry">ContactInfoRegistry</a>
  `lorawan-stack/api/contact_info.proto`

  

  
### <a name="RequestValidation">RequestValidation</a>
  Request validation for the non-validated contact info for the given entity.

  {{% reftab RequestValidation gRPCRequestValidation HTTPRequestValidation %}}

  **Request**: [EntityIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.EntityIdentifiers" >}})

  **Response**: [ContactInfoValidation]({{< ref "messages.md#ttn.lorawan.v3.ContactInfoValidation" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/contact_info/validation` |  |{{% /reftab %}}

  
### <a name="Validate">Validate</a>
  Validate confirms a contact info validation.

  {{% reftab Validate gRPCValidate HTTPValidate %}}

  **Request**: [ContactInfoValidation]({{< ref "messages.md#ttn.lorawan.v3.ContactInfoValidation" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PATCH` | `/api/v3/contact_info/validation` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="EndDeviceRegistry">EndDeviceRegistry</a>
  `lorawan-stack/api/end_device_services.proto`

  

  
### <a name="Create">Create</a>
  Create a new end device within an application.

  {{% reftab Create gRPCCreate HTTPCreate %}}

  **Request**: [CreateEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/applications/{end_device.ids.application_ids.application_id}/devices` | * |{{% /reftab %}}

  
### <a name="Get">Get</a>
  Get the end device with the given identifiers, selecting the fields given
by the field mask.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.GetEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |{{% /reftab %}}

  
### <a name="List">List</a>
  List applications. See request message for details.

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListEndDevicesRequest]({{< ref "messages.md#ttn.lorawan.v3.ListEndDevicesRequest" >}})

  **Response**: [EndDevices]({{< ref "messages.md#ttn.lorawan.v3.EndDevices" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/applications/{application_ids.application_id}/devices` |  |{{% /reftab %}}

  
### <a name="Update">Update</a>
  

  {{% reftab Update gRPCUpdate HTTPUpdate %}}

  **Request**: [UpdateEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [EndDeviceIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.EndDeviceIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/applications/{application_ids.application_id}/devices/{device_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="Events">Events</a>
  `lorawan-stack/api/events.proto`

  The Events service serves events from the cluster.

  
### <a name="Stream">Stream</a>
  Stream live events, optionally with a tail of historical events (depending on server support and retention policy).
Events may arrive out-of-order.

  {{% reftab Stream gRPCStream HTTPStream %}}

  **Request**: [StreamEventsRequest]({{< ref "messages.md#ttn.lorawan.v3.StreamEventsRequest" >}})

  **Response**: [Event]({{< ref "messages.md#ttn.lorawan.v3.Event" >}}) _stream_

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/events` | * |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="GatewayAccess">GatewayAccess</a>
  `lorawan-stack/api/gateway_services.proto`

  

  
### <a name="ListRights">ListRights</a>
  

  {{% reftab ListRights gRPCListRights HTTPListRights %}}

  **Request**: [GatewayIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.GatewayIdentifiers" >}})

  **Response**: [Rights]({{< ref "messages.md#ttn.lorawan.v3.Rights" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gateways/{gateway_id}/rights` |  |{{% /reftab %}}

  
### <a name="CreateAPIKey">CreateAPIKey</a>
  

  {{% reftab CreateAPIKey gRPCCreateAPIKey HTTPCreateAPIKey %}}

  **Request**: [CreateGatewayAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateGatewayAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys` | * |{{% /reftab %}}

  
### <a name="ListAPIKeys">ListAPIKeys</a>
  

  {{% reftab ListAPIKeys gRPCListAPIKeys HTTPListAPIKeys %}}

  **Request**: [ListGatewayAPIKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.ListGatewayAPIKeysRequest" >}})

  **Response**: [APIKeys]({{< ref "messages.md#ttn.lorawan.v3.APIKeys" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys` |  |{{% /reftab %}}

  
### <a name="GetAPIKey">GetAPIKey</a>
  

  {{% reftab GetAPIKey gRPCGetAPIKey HTTPGetAPIKey %}}

  **Request**: [GetGatewayAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.GetGatewayAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys/{key_id}` |  |{{% /reftab %}}

  
### <a name="UpdateAPIKey">UpdateAPIKey</a>
  Update the rights of an existing gateway API key. To generate an API key,
the CreateAPIKey should be used. To delete an API key, update it
with zero rights.

  {{% reftab UpdateAPIKey gRPCUpdateAPIKey HTTPUpdateAPIKey %}}

  **Request**: [UpdateGatewayAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateGatewayAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys/{api_key.id}` | * |{{% /reftab %}}

  
### <a name="GetCollaborator">GetCollaborator</a>
  Get the rights of a collaborator (member) of the gateway.
Pseudo-rights in the response (such as the "_ALL" right) are not expanded.

  {{% reftab GetCollaborator gRPCGetCollaborator HTTPGetCollaborator %}}

  **Request**: [GetGatewayCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.GetGatewayCollaboratorRequest" >}})

  **Response**: [GetCollaboratorResponse]({{< ref "messages.md#ttn.lorawan.v3.GetCollaboratorResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborator` |  |
 `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
 `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborator/organization/{collaborator.organization_ids.organization_id}` |  |{{% /reftab %}}

  
### <a name="SetCollaborator">SetCollaborator</a>
  Set the rights of a collaborator (member) on the gateway.
Setting a collaborator without rights, removes them.

  {{% reftab SetCollaborator gRPCSetCollaborator HTTPSetCollaborator %}}

  **Request**: [SetGatewayCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.SetGatewayCollaboratorRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborators` | * |{{% /reftab %}}

  
### <a name="ListCollaborators">ListCollaborators</a>
  

  {{% reftab ListCollaborators gRPCListCollaborators HTTPListCollaborators %}}

  **Request**: [ListGatewayCollaboratorsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListGatewayCollaboratorsRequest" >}})

  **Response**: [Collaborators]({{< ref "messages.md#ttn.lorawan.v3.Collaborators" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborators` |  |{{% /reftab %}}

  
  

## <a name="GatewayConfigurator">GatewayConfigurator</a>
  `lorawan-stack/api/gateway_services.proto`

  

  
### <a name="PullConfiguration">PullConfiguration</a>
  

  {{% reftab PullConfiguration gRPCPullConfiguration HTTPPullConfiguration %}}

  **Request**: [PullGatewayConfigurationRequest]({{< ref "messages.md#ttn.lorawan.v3.PullGatewayConfigurationRequest" >}})

  **Response**: [Gateway]({{< ref "messages.md#ttn.lorawan.v3.Gateway" >}}) _stream_

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="GatewayRegistry">GatewayRegistry</a>
  `lorawan-stack/api/gateway_services.proto`

  

  
### <a name="Create">Create</a>
  Create a new gateway. This also sets the given organization or user as
first collaborator with all possible rights.

  {{% reftab Create gRPCCreate HTTPCreate %}}

  **Request**: [CreateGatewayRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateGatewayRequest" >}})

  **Response**: [Gateway]({{< ref "messages.md#ttn.lorawan.v3.Gateway" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/gateways` | * |
 `POST` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/gateways` | * |{{% /reftab %}}

  
### <a name="Get">Get</a>
  Get the gateway with the given identifiers, selecting the fields given
by the field mask. The method may return more or less fields, depending on
the rights of the caller.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetGatewayRequest]({{< ref "messages.md#ttn.lorawan.v3.GetGatewayRequest" >}})

  **Response**: [Gateway]({{< ref "messages.md#ttn.lorawan.v3.Gateway" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}` |  |{{% /reftab %}}

  
### <a name="GetIdentifiersForEUI">GetIdentifiersForEUI</a>
  

  {{% reftab GetIdentifiersForEUI gRPCGetIdentifiersForEUI HTTPGetIdentifiersForEUI %}}

  **Request**: [GetGatewayIdentifiersForEUIRequest]({{< ref "messages.md#ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest" >}})

  **Response**: [GatewayIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.GatewayIdentifiers" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="List">List</a>
  List gateways. See request message for details.

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListGatewaysRequest]({{< ref "messages.md#ttn.lorawan.v3.ListGatewaysRequest" >}})

  **Response**: [Gateways]({{< ref "messages.md#ttn.lorawan.v3.Gateways" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gateways` |  |
 `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/gateways` |  |
 `GET` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/gateways` |  |{{% /reftab %}}

  
### <a name="Update">Update</a>
  

  {{% reftab Update gRPCUpdate HTTPUpdate %}}

  **Request**: [UpdateGatewayRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateGatewayRequest" >}})

  **Response**: [Gateway]({{< ref "messages.md#ttn.lorawan.v3.Gateway" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/gateways/{gateway.ids.gateway_id}` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [GatewayIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.GatewayIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/gateways/{gateway_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  

## <a name="Gs">Gs</a>
  `lorawan-stack/api/gatewayserver.proto`

  

  
### <a name="GetGatewayConnectionStats">GetGatewayConnectionStats</a>
  Get statistics about the current gateway connection to the Gateway Server.
This is not persisted between reconnects.

  {{% reftab GetGatewayConnectionStats gRPCGetGatewayConnectionStats HTTPGetGatewayConnectionStats %}}

  **Request**: [GatewayIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.GatewayIdentifiers" >}})

  **Response**: [GatewayConnectionStats]({{< ref "messages.md#ttn.lorawan.v3.GatewayConnectionStats" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/gs/gateways/{gateway_id}/connection/stats` |  |{{% /reftab %}}

  
  

## <a name="GtwGs">GtwGs</a>
  `lorawan-stack/api/gatewayserver.proto`

  The GtwGs service connects a gateway to a Gateway Server.

  
### <a name="LinkGateway">LinkGateway</a>
  Link the gateway to the Gateway Server.

  {{% reftab LinkGateway gRPCLinkGateway HTTPLinkGateway %}}

  **Request**: [GatewayUp]({{< ref "messages.md#ttn.lorawan.v3.GatewayUp" >}}) _stream_

  **Response**: [GatewayDown]({{< ref "messages.md#ttn.lorawan.v3.GatewayDown" >}}) _stream_

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="GetConcentratorConfig">GetConcentratorConfig</a>
  GetConcentratorConfig associated to the gateway.

  {{% reftab GetConcentratorConfig gRPCGetConcentratorConfig HTTPGetConcentratorConfig %}}

  **Request**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  **Response**: [ConcentratorConfig]({{< ref "messages.md#ttn.lorawan.v3.ConcentratorConfig" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="NsGs">NsGs</a>
  `lorawan-stack/api/gatewayserver.proto`

  The NsGs service connects a Network Server to a Gateway Server.

  
### <a name="ScheduleDownlink">ScheduleDownlink</a>
  ScheduleDownlink instructs the Gateway Server to schedule a downlink message.
The Gateway Server may refuse if there are any conflicts in the schedule or
if a duty cycle prevents the gateway from transmitting.

  {{% reftab ScheduleDownlink gRPCScheduleDownlink HTTPScheduleDownlink %}}

  **Request**: [DownlinkMessage]({{< ref "messages.md#ttn.lorawan.v3.DownlinkMessage" >}})

  **Response**: [ScheduleDownlinkResponse]({{< ref "messages.md#ttn.lorawan.v3.ScheduleDownlinkResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="EntityAccess">EntityAccess</a>
  `lorawan-stack/api/identityserver.proto`

  

  
### <a name="AuthInfo">AuthInfo</a>
  AuthInfo returns information about the authentication that is used on the request.

  {{% reftab AuthInfo gRPCAuthInfo HTTPAuthInfo %}}

  **Request**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  **Response**: [AuthInfoResponse]({{< ref "messages.md#ttn.lorawan.v3.AuthInfoResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/auth_info` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="ApplicationCryptoService">ApplicationCryptoService</a>
  `lorawan-stack/api/joinserver.proto`

  Service for application layer cryptographic operations.

  
### <a name="DeriveAppSKey">DeriveAppSKey</a>
  

  {{% reftab DeriveAppSKey gRPCDeriveAppSKey HTTPDeriveAppSKey %}}

  **Request**: [DeriveSessionKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.DeriveSessionKeysRequest" >}})

  **Response**: [AppSKeyResponse]({{< ref "messages.md#ttn.lorawan.v3.AppSKeyResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="GetAppKey">GetAppKey</a>
  Get the AppKey. Crypto Servers may return status code UNIMPLEMENTED when root keys are not exposed.

  {{% reftab GetAppKey gRPCGetAppKey HTTPGetAppKey %}}

  **Request**: [GetRootKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.GetRootKeysRequest" >}})

  **Response**: [KeyEnvelope]({{< ref "messages.md#ttn.lorawan.v3.KeyEnvelope" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="AsJs">AsJs</a>
  `lorawan-stack/api/joinserver.proto`

  The AsJs service connects an Application Server to a Join Server.

  
### <a name="GetAppSKey">GetAppSKey</a>
  

  {{% reftab GetAppSKey gRPCGetAppSKey HTTPGetAppSKey %}}

  **Request**: [SessionKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.SessionKeyRequest" >}})

  **Response**: [AppSKeyResponse]({{< ref "messages.md#ttn.lorawan.v3.AppSKeyResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="Js">Js</a>
  `lorawan-stack/api/joinserver.proto`

  

  
### <a name="GetJoinEUIPrefixes">GetJoinEUIPrefixes</a>
  

  {{% reftab GetJoinEUIPrefixes gRPCGetJoinEUIPrefixes HTTPGetJoinEUIPrefixes %}}

  **Request**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  **Response**: [JoinEUIPrefixes]({{< ref "messages.md#ttn.lorawan.v3.JoinEUIPrefixes" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/js/join_eui_prefixes` |  |{{% /reftab %}}

  
  

## <a name="JsEndDeviceRegistry">JsEndDeviceRegistry</a>
  `lorawan-stack/api/joinserver.proto`

  The JsEndDeviceRegistry service allows clients to manage their end devices on the Join Server.

  
### <a name="Get">Get</a>
  Get returns the device that matches the given identifiers.
If there are multiple matches, an error will be returned.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.GetEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/js/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |{{% /reftab %}}

  
### <a name="Set">Set</a>
  Set creates or updates the device.

  {{% reftab Set gRPCSet HTTPSet %}}

  **Request**: [SetEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.SetEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/js/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | * |
 `POST` | `/api/v3/js/applications/{end_device.ids.application_ids.application_id}/devices` | * |{{% /reftab %}}

  
### <a name="Provision">Provision</a>
  Provision returns end devices that are provisioned using the given vendor-specific data.
The devices are not set in the registry.

  {{% reftab Provision gRPCProvision HTTPProvision %}}

  **Request**: [ProvisionEndDevicesRequest]({{< ref "messages.md#ttn.lorawan.v3.ProvisionEndDevicesRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}}) _stream_

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/js/applications/{application_ids.application_id}/provision-devices` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  Delete deletes the device that matches the given identifiers.
If there are multiple matches, an error will be returned.

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [EndDeviceIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.EndDeviceIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/js/applications/{application_ids.application_id}/devices/{device_id}` |  |{{% /reftab %}}

  
  

## <a name="NetworkCryptoService">NetworkCryptoService</a>
  `lorawan-stack/api/joinserver.proto`

  Service for network layer cryptographic operations.

  
### <a name="JoinRequestMIC">JoinRequestMIC</a>
  

  {{% reftab JoinRequestMIC gRPCJoinRequestMIC HTTPJoinRequestMIC %}}

  **Request**: [CryptoServicePayloadRequest]({{< ref "messages.md#ttn.lorawan.v3.CryptoServicePayloadRequest" >}})

  **Response**: [CryptoServicePayloadResponse]({{< ref "messages.md#ttn.lorawan.v3.CryptoServicePayloadResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="JoinAcceptMIC">JoinAcceptMIC</a>
  

  {{% reftab JoinAcceptMIC gRPCJoinAcceptMIC HTTPJoinAcceptMIC %}}

  **Request**: [JoinAcceptMICRequest]({{< ref "messages.md#ttn.lorawan.v3.JoinAcceptMICRequest" >}})

  **Response**: [CryptoServicePayloadResponse]({{< ref "messages.md#ttn.lorawan.v3.CryptoServicePayloadResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="EncryptJoinAccept">EncryptJoinAccept</a>
  

  {{% reftab EncryptJoinAccept gRPCEncryptJoinAccept HTTPEncryptJoinAccept %}}

  **Request**: [CryptoServicePayloadRequest]({{< ref "messages.md#ttn.lorawan.v3.CryptoServicePayloadRequest" >}})

  **Response**: [CryptoServicePayloadResponse]({{< ref "messages.md#ttn.lorawan.v3.CryptoServicePayloadResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="EncryptRejoinAccept">EncryptRejoinAccept</a>
  

  {{% reftab EncryptRejoinAccept gRPCEncryptRejoinAccept HTTPEncryptRejoinAccept %}}

  **Request**: [CryptoServicePayloadRequest]({{< ref "messages.md#ttn.lorawan.v3.CryptoServicePayloadRequest" >}})

  **Response**: [CryptoServicePayloadResponse]({{< ref "messages.md#ttn.lorawan.v3.CryptoServicePayloadResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="DeriveNwkSKeys">DeriveNwkSKeys</a>
  

  {{% reftab DeriveNwkSKeys gRPCDeriveNwkSKeys HTTPDeriveNwkSKeys %}}

  **Request**: [DeriveSessionKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.DeriveSessionKeysRequest" >}})

  **Response**: [NwkSKeysResponse]({{< ref "messages.md#ttn.lorawan.v3.NwkSKeysResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="GetNwkKey">GetNwkKey</a>
  Get the NwkKey. Crypto Servers may return status code UNIMPLEMENTED when root keys are not exposed.

  {{% reftab GetNwkKey gRPCGetNwkKey HTTPGetNwkKey %}}

  **Request**: [GetRootKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.GetRootKeysRequest" >}})

  **Response**: [KeyEnvelope]({{< ref "messages.md#ttn.lorawan.v3.KeyEnvelope" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="NsJs">NsJs</a>
  `lorawan-stack/api/joinserver.proto`

  The NsJs service connects a Network Server to a Join Server.

  
### <a name="HandleJoin">HandleJoin</a>
  

  {{% reftab HandleJoin gRPCHandleJoin HTTPHandleJoin %}}

  **Request**: [JoinRequest]({{< ref "messages.md#ttn.lorawan.v3.JoinRequest" >}})

  **Response**: [JoinResponse]({{< ref "messages.md#ttn.lorawan.v3.JoinResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="GetNwkSKeys">GetNwkSKeys</a>
  

  {{% reftab GetNwkSKeys gRPCGetNwkSKeys HTTPGetNwkSKeys %}}

  **Request**: [SessionKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.SessionKeyRequest" >}})

  **Response**: [NwkSKeysResponse]({{< ref "messages.md#ttn.lorawan.v3.NwkSKeysResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="DownlinkMessageProcessor">DownlinkMessageProcessor</a>
  `lorawan-stack/api/message_services.proto`

  The DownlinkMessageProcessor service processes downlink messages.

  
### <a name="Process">Process</a>
  

  {{% reftab Process gRPCProcess HTTPProcess %}}

  **Request**: [ProcessDownlinkMessageRequest]({{< ref "messages.md#ttn.lorawan.v3.ProcessDownlinkMessageRequest" >}})

  **Response**: [ApplicationDownlink]({{< ref "messages.md#ttn.lorawan.v3.ApplicationDownlink" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="UplinkMessageProcessor">UplinkMessageProcessor</a>
  `lorawan-stack/api/message_services.proto`

  The UplinkMessageProcessor service processes uplink messages.

  
### <a name="Process">Process</a>
  

  {{% reftab Process gRPCProcess HTTPProcess %}}

  **Request**: [ProcessUplinkMessageRequest]({{< ref "messages.md#ttn.lorawan.v3.ProcessUplinkMessageRequest" >}})

  **Response**: [ApplicationUplink]({{< ref "messages.md#ttn.lorawan.v3.ApplicationUplink" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="AsNs">AsNs</a>
  `lorawan-stack/api/networkserver.proto`

  The AsNs service connects an Application Server to a Network Server.

  
### <a name="LinkApplication">LinkApplication</a>
  

  {{% reftab LinkApplication gRPCLinkApplication HTTPLinkApplication %}}

  **Request**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}}) _stream_

  **Response**: [ApplicationUp]({{< ref "messages.md#ttn.lorawan.v3.ApplicationUp" >}}) _stream_

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="DownlinkQueueReplace">DownlinkQueueReplace</a>
  

  {{% reftab DownlinkQueueReplace gRPCDownlinkQueueReplace HTTPDownlinkQueueReplace %}}

  **Request**: [DownlinkQueueRequest]({{< ref "messages.md#ttn.lorawan.v3.DownlinkQueueRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="DownlinkQueuePush">DownlinkQueuePush</a>
  

  {{% reftab DownlinkQueuePush gRPCDownlinkQueuePush HTTPDownlinkQueuePush %}}

  **Request**: [DownlinkQueueRequest]({{< ref "messages.md#ttn.lorawan.v3.DownlinkQueueRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
### <a name="DownlinkQueueList">DownlinkQueueList</a>
  

  {{% reftab DownlinkQueueList gRPCDownlinkQueueList HTTPDownlinkQueueList %}}

  **Request**: [EndDeviceIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.EndDeviceIdentifiers" >}})

  **Response**: [ApplicationDownlinks]({{< ref "messages.md#ttn.lorawan.v3.ApplicationDownlinks" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="GsNs">GsNs</a>
  `lorawan-stack/api/networkserver.proto`

  The GsNs service connects a Gateway Server to a Network Server.

  
### <a name="HandleUplink">HandleUplink</a>
  

  {{% reftab HandleUplink gRPCHandleUplink HTTPHandleUplink %}}

  **Request**: [UplinkMessage]({{< ref "messages.md#ttn.lorawan.v3.UplinkMessage" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----{{% /reftab %}}

  
  

## <a name="Ns">Ns</a>
  `lorawan-stack/api/networkserver.proto`

  

  
### <a name="GenerateDevAddr">GenerateDevAddr</a>
  GenerateDevAddr requests a device address assignment from the Network Server.

  {{% reftab GenerateDevAddr gRPCGenerateDevAddr HTTPGenerateDevAddr %}}

  **Request**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  **Response**: [GenerateDevAddrResponse]({{< ref "messages.md#ttn.lorawan.v3.GenerateDevAddrResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/ns/dev_addr` |  |{{% /reftab %}}

  
  

## <a name="NsEndDeviceRegistry">NsEndDeviceRegistry</a>
  `lorawan-stack/api/networkserver.proto`

  The NsEndDeviceRegistry service allows clients to manage their end devices on the Network Server.

  
### <a name="Get">Get</a>
  Get returns the device that matches the given identifiers.
If there are multiple matches, an error will be returned.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.GetEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/ns/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |{{% /reftab %}}

  
### <a name="Set">Set</a>
  Set creates or updates the device.

  {{% reftab Set gRPCSet HTTPSet %}}

  **Request**: [SetEndDeviceRequest]({{< ref "messages.md#ttn.lorawan.v3.SetEndDeviceRequest" >}})

  **Response**: [EndDevice]({{< ref "messages.md#ttn.lorawan.v3.EndDevice" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/ns/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | * |
 `POST` | `/api/v3/ns/applications/{end_device.ids.application_ids.application_id}/devices` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  Delete deletes the device that matches the given identifiers.
If there are multiple matches, an error will be returned.

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [EndDeviceIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.EndDeviceIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/ns/applications/{application_ids.application_id}/devices/{device_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="OAuthAuthorizationRegistry">OAuthAuthorizationRegistry</a>
  `lorawan-stack/api/oauth_services.proto`

  

  
### <a name="List">List</a>
  

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListOAuthClientAuthorizationsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest" >}})

  **Response**: [OAuthClientAuthorizations]({{< ref "messages.md#ttn.lorawan.v3.OAuthClientAuthorizations" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/users/{user_ids.user_id}/authorizations` |  |{{% /reftab %}}

  
### <a name="ListTokens">ListTokens</a>
  

  {{% reftab ListTokens gRPCListTokens HTTPListTokens %}}

  **Request**: [ListOAuthAccessTokensRequest]({{< ref "messages.md#ttn.lorawan.v3.ListOAuthAccessTokensRequest" >}})

  **Response**: [OAuthAccessTokens]({{< ref "messages.md#ttn.lorawan.v3.OAuthAccessTokens" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/users/{user_ids.user_id}/authorizations/{client_ids.client_id}/tokens` |  |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [OAuthClientAuthorizationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/users/{user_ids.user_id}/authorizations/{client_ids.client_id}` |  |{{% /reftab %}}

  
### <a name="DeleteToken">DeleteToken</a>
  

  {{% reftab DeleteToken gRPCDeleteToken HTTPDeleteToken %}}

  **Request**: [OAuthAccessTokenIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.OAuthAccessTokenIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/users/{user_ids.user_id}/authorizations/{client_ids.client_id}/tokens/{id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="OrganizationAccess">OrganizationAccess</a>
  `lorawan-stack/api/organization_services.proto`

  

  
### <a name="ListRights">ListRights</a>
  

  {{% reftab ListRights gRPCListRights HTTPListRights %}}

  **Request**: [OrganizationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.OrganizationIdentifiers" >}})

  **Response**: [Rights]({{< ref "messages.md#ttn.lorawan.v3.Rights" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/organizations/{organization_id}/rights` |  |{{% /reftab %}}

  
### <a name="CreateAPIKey">CreateAPIKey</a>
  

  {{% reftab CreateAPIKey gRPCCreateAPIKey HTTPCreateAPIKey %}}

  **Request**: [CreateOrganizationAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateOrganizationAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys` | * |{{% /reftab %}}

  
### <a name="ListAPIKeys">ListAPIKeys</a>
  

  {{% reftab ListAPIKeys gRPCListAPIKeys HTTPListAPIKeys %}}

  **Request**: [ListOrganizationAPIKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.ListOrganizationAPIKeysRequest" >}})

  **Response**: [APIKeys]({{< ref "messages.md#ttn.lorawan.v3.APIKeys" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys` |  |{{% /reftab %}}

  
### <a name="GetAPIKey">GetAPIKey</a>
  

  {{% reftab GetAPIKey gRPCGetAPIKey HTTPGetAPIKey %}}

  **Request**: [GetOrganizationAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.GetOrganizationAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys/{key_id}` |  |{{% /reftab %}}

  
### <a name="UpdateAPIKey">UpdateAPIKey</a>
  Update the rights of an existing organization API key. To generate an API key,
the CreateAPIKey should be used. To delete an API key, update it
with zero rights.

  {{% reftab UpdateAPIKey gRPCUpdateAPIKey HTTPUpdateAPIKey %}}

  **Request**: [UpdateOrganizationAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys/{api_key.id}` | * |{{% /reftab %}}

  
### <a name="GetCollaborator">GetCollaborator</a>
  Get the rights of a collaborator (member) of the organization.
Pseudo-rights in the response (such as the "_ALL" right) are not expanded.

  {{% reftab GetCollaborator gRPCGetCollaborator HTTPGetCollaborator %}}

  **Request**: [GetOrganizationCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.GetOrganizationCollaboratorRequest" >}})

  **Response**: [GetCollaboratorResponse]({{< ref "messages.md#ttn.lorawan.v3.GetCollaboratorResponse" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/organizations/{organization_ids.organization_id}/collaborator` |  |
 `GET` | `/api/v3/organizations/{organization_ids.organization_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |{{% /reftab %}}

  
### <a name="SetCollaborator">SetCollaborator</a>
  Set the rights of a collaborator (member) on the organization.
Setting a collaborator without rights, removes them.
Note that only users can collaborate (be member of) an organization.

  {{% reftab SetCollaborator gRPCSetCollaborator HTTPSetCollaborator %}}

  **Request**: [SetOrganizationCollaboratorRequest]({{< ref "messages.md#ttn.lorawan.v3.SetOrganizationCollaboratorRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/organizations/{organization_ids.organization_id}/collaborators` | * |{{% /reftab %}}

  
### <a name="ListCollaborators">ListCollaborators</a>
  

  {{% reftab ListCollaborators gRPCListCollaborators HTTPListCollaborators %}}

  **Request**: [ListOrganizationCollaboratorsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListOrganizationCollaboratorsRequest" >}})

  **Response**: [Collaborators]({{< ref "messages.md#ttn.lorawan.v3.Collaborators" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/organizations/{organization_ids.organization_id}/collaborators` |  |{{% /reftab %}}

  
  

## <a name="OrganizationRegistry">OrganizationRegistry</a>
  `lorawan-stack/api/organization_services.proto`

  

  
### <a name="Create">Create</a>
  Create a new organization. This also sets the given user as
first collaborator with all possible rights.

  {{% reftab Create gRPCCreate HTTPCreate %}}

  **Request**: [CreateOrganizationRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateOrganizationRequest" >}})

  **Response**: [Organization]({{< ref "messages.md#ttn.lorawan.v3.Organization" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/organizations` | * |{{% /reftab %}}

  
### <a name="Get">Get</a>
  Get the organization with the given identifiers, selecting the fields given
by the field mask. The method may return more or less fields, depending on
the rights of the caller.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetOrganizationRequest]({{< ref "messages.md#ttn.lorawan.v3.GetOrganizationRequest" >}})

  **Response**: [Organization]({{< ref "messages.md#ttn.lorawan.v3.Organization" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/organizations/{organization_ids.organization_id}` |  |{{% /reftab %}}

  
### <a name="List">List</a>
  List organizations. See request message for details.

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListOrganizationsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListOrganizationsRequest" >}})

  **Response**: [Organizations]({{< ref "messages.md#ttn.lorawan.v3.Organizations" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/organizations` |  |
 `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/organizations` |  |{{% /reftab %}}

  
### <a name="Update">Update</a>
  

  {{% reftab Update gRPCUpdate HTTPUpdate %}}

  **Request**: [UpdateOrganizationRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateOrganizationRequest" >}})

  **Response**: [Organization]({{< ref "messages.md#ttn.lorawan.v3.Organization" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/organizations/{organization.ids.organization_id}` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [OrganizationIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.OrganizationIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/organizations/{organization_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="EndDeviceRegistrySearch">EndDeviceRegistrySearch</a>
  `lorawan-stack/api/search_services.proto`

  The EndDeviceRegistrySearch service indexes devices in the EndDeviceRegistry
and enables searching for them.
This service is not implemented on all deployments.

  
### <a name="SearchEndDevices">SearchEndDevices</a>
  

  {{% reftab SearchEndDevices gRPCSearchEndDevices HTTPSearchEndDevices %}}

  **Request**: [SearchEndDevicesRequest]({{< ref "messages.md#ttn.lorawan.v3.SearchEndDevicesRequest" >}})

  **Response**: [EndDevices]({{< ref "messages.md#ttn.lorawan.v3.EndDevices" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/search/applications/{application_ids.application_id}/devices` |  |{{% /reftab %}}

  
  

## <a name="EntityRegistrySearch">EntityRegistrySearch</a>
  `lorawan-stack/api/search_services.proto`

  The EntityRegistrySearch service indexes entities in the various registries
and enables searching for them.
This service is not implemented on all deployments.

  
### <a name="SearchApplications">SearchApplications</a>
  

  {{% reftab SearchApplications gRPCSearchApplications HTTPSearchApplications %}}

  **Request**: [SearchEntitiesRequest]({{< ref "messages.md#ttn.lorawan.v3.SearchEntitiesRequest" >}})

  **Response**: [Applications]({{< ref "messages.md#ttn.lorawan.v3.Applications" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/search/applications` |  |{{% /reftab %}}

  
### <a name="SearchClients">SearchClients</a>
  

  {{% reftab SearchClients gRPCSearchClients HTTPSearchClients %}}

  **Request**: [SearchEntitiesRequest]({{< ref "messages.md#ttn.lorawan.v3.SearchEntitiesRequest" >}})

  **Response**: [Clients]({{< ref "messages.md#ttn.lorawan.v3.Clients" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/search/clients` |  |{{% /reftab %}}

  
### <a name="SearchGateways">SearchGateways</a>
  

  {{% reftab SearchGateways gRPCSearchGateways HTTPSearchGateways %}}

  **Request**: [SearchEntitiesRequest]({{< ref "messages.md#ttn.lorawan.v3.SearchEntitiesRequest" >}})

  **Response**: [Gateways]({{< ref "messages.md#ttn.lorawan.v3.Gateways" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/search/gateways` |  |{{% /reftab %}}

  
### <a name="SearchOrganizations">SearchOrganizations</a>
  

  {{% reftab SearchOrganizations gRPCSearchOrganizations HTTPSearchOrganizations %}}

  **Request**: [SearchEntitiesRequest]({{< ref "messages.md#ttn.lorawan.v3.SearchEntitiesRequest" >}})

  **Response**: [Organizations]({{< ref "messages.md#ttn.lorawan.v3.Organizations" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/search/organizations` |  |{{% /reftab %}}

  
### <a name="SearchUsers">SearchUsers</a>
  

  {{% reftab SearchUsers gRPCSearchUsers HTTPSearchUsers %}}

  **Request**: [SearchEntitiesRequest]({{< ref "messages.md#ttn.lorawan.v3.SearchEntitiesRequest" >}})

  **Response**: [Users]({{< ref "messages.md#ttn.lorawan.v3.Users" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/search/users` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

  
{{% refswitcher %}}

  

## <a name="UserAccess">UserAccess</a>
  `lorawan-stack/api/user_services.proto`

  

  
### <a name="ListRights">ListRights</a>
  

  {{% reftab ListRights gRPCListRights HTTPListRights %}}

  **Request**: [UserIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.UserIdentifiers" >}})

  **Response**: [Rights]({{< ref "messages.md#ttn.lorawan.v3.Rights" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/users/{user_id}/rights` |  |{{% /reftab %}}

  
### <a name="CreateAPIKey">CreateAPIKey</a>
  

  {{% reftab CreateAPIKey gRPCCreateAPIKey HTTPCreateAPIKey %}}

  **Request**: [CreateUserAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateUserAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/users/{user_ids.user_id}/api-keys` | * |{{% /reftab %}}

  
### <a name="ListAPIKeys">ListAPIKeys</a>
  

  {{% reftab ListAPIKeys gRPCListAPIKeys HTTPListAPIKeys %}}

  **Request**: [ListUserAPIKeysRequest]({{< ref "messages.md#ttn.lorawan.v3.ListUserAPIKeysRequest" >}})

  **Response**: [APIKeys]({{< ref "messages.md#ttn.lorawan.v3.APIKeys" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/users/{user_ids.user_id}/api-keys` |  |{{% /reftab %}}

  
### <a name="GetAPIKey">GetAPIKey</a>
  

  {{% reftab GetAPIKey gRPCGetAPIKey HTTPGetAPIKey %}}

  **Request**: [GetUserAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.GetUserAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/users/{user_ids.user_id}/api-keys/{key_id}` |  |{{% /reftab %}}

  
### <a name="UpdateAPIKey">UpdateAPIKey</a>
  Update the rights of an existing user API key. To generate an API key,
the CreateAPIKey should be used. To delete an API key, update it
with zero rights.

  {{% reftab UpdateAPIKey gRPCUpdateAPIKey HTTPUpdateAPIKey %}}

  **Request**: [UpdateUserAPIKeyRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateUserAPIKeyRequest" >}})

  **Response**: [APIKey]({{< ref "messages.md#ttn.lorawan.v3.APIKey" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/users/{user_ids.user_id}/api-keys/{api_key.id}` | * |{{% /reftab %}}

  
  

## <a name="UserInvitationRegistry">UserInvitationRegistry</a>
  `lorawan-stack/api/user_services.proto`

  

  
### <a name="Send">Send</a>
  

  {{% reftab Send gRPCSend HTTPSend %}}

  **Request**: [SendInvitationRequest]({{< ref "messages.md#ttn.lorawan.v3.SendInvitationRequest" >}})

  **Response**: [Invitation]({{< ref "messages.md#ttn.lorawan.v3.Invitation" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/invitations` | * |{{% /reftab %}}

  
### <a name="List">List</a>
  

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListInvitationsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListInvitationsRequest" >}})

  **Response**: [Invitations]({{< ref "messages.md#ttn.lorawan.v3.Invitations" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/invitations` |  |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [DeleteInvitationRequest]({{< ref "messages.md#ttn.lorawan.v3.DeleteInvitationRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/invitations` |  |{{% /reftab %}}

  
  

## <a name="UserRegistry">UserRegistry</a>
  `lorawan-stack/api/user_services.proto`

  

  
### <a name="Create">Create</a>
  Register a new user. This method may be restricted by network settings.

  {{% reftab Create gRPCCreate HTTPCreate %}}

  **Request**: [CreateUserRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateUserRequest" >}})

  **Response**: [User]({{< ref "messages.md#ttn.lorawan.v3.User" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/users` | * |{{% /reftab %}}

  
### <a name="Get">Get</a>
  Get the user with the given identifiers, selecting the fields given by the
field mask. The method may return more or less fields, depending on the rights
of the caller.

  {{% reftab Get gRPCGet HTTPGet %}}

  **Request**: [GetUserRequest]({{< ref "messages.md#ttn.lorawan.v3.GetUserRequest" >}})

  **Response**: [User]({{< ref "messages.md#ttn.lorawan.v3.User" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/users/{user_ids.user_id}` |  |{{% /reftab %}}

  
### <a name="Update">Update</a>
  

  {{% reftab Update gRPCUpdate HTTPUpdate %}}

  **Request**: [UpdateUserRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateUserRequest" >}})

  **Response**: [User]({{< ref "messages.md#ttn.lorawan.v3.User" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/users/{user.ids.user_id}` | * |{{% /reftab %}}

  
### <a name="CreateTemporaryPassword">CreateTemporaryPassword</a>
  Create a temporary password that can be used for updating a forgotten password.
The generated password is sent to the user's email address.

  {{% reftab CreateTemporaryPassword gRPCCreateTemporaryPassword HTTPCreateTemporaryPassword %}}

  **Request**: [CreateTemporaryPasswordRequest]({{< ref "messages.md#ttn.lorawan.v3.CreateTemporaryPasswordRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `POST` | `/api/v3/users/{user_ids.user_id}/temporary_password` |  |{{% /reftab %}}

  
### <a name="UpdatePassword">UpdatePassword</a>
  

  {{% reftab UpdatePassword gRPCUpdatePassword HTTPUpdatePassword %}}

  **Request**: [UpdateUserPasswordRequest]({{< ref "messages.md#ttn.lorawan.v3.UpdateUserPasswordRequest" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `PUT` | `/api/v3/users/{user_ids.user_id}/password` | * |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [UserIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.UserIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/users/{user_id}` |  |{{% /reftab %}}

  
  

## <a name="UserSessionRegistry">UserSessionRegistry</a>
  `lorawan-stack/api/user_services.proto`

  

  
### <a name="List">List</a>
  

  {{% reftab List gRPCList HTTPList %}}

  **Request**: [ListUserSessionsRequest]({{< ref "messages.md#ttn.lorawan.v3.ListUserSessionsRequest" >}})

  **Response**: [UserSessions]({{< ref "messages.md#ttn.lorawan.v3.UserSessions" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `GET` | `/api/v3/users/{user_ids.user_id}/sessions` |  |{{% /reftab %}}

  
### <a name="Delete">Delete</a>
  

  {{% reftab Delete gRPCDelete HTTPDelete %}}

  **Request**: [UserSessionIdentifiers]({{< ref "messages.md#ttn.lorawan.v3.UserSessionIdentifiers" >}})

  **Response**: [.google.protobuf.Empty]({{< ref "messages.md#google.protobuf.Empty" >}})

  $$$$$$

Method | Pattern | Body
------|-------|----
 `DELETE` | `/api/v3/users/{user_ids.user_id}/sessions/{session_id}` |  |{{% /reftab %}}

  
  
{{% refswitcher %}}

