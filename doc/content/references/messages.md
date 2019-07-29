---
title: "Messages"
description: "All messages type available use by the API."
weight: 2
---

Name | Description
---|---
[Application](#ttn.lorawan.v3.Application) | Application is the message that defines an Application in the network.
  [AttributesEntry](#ttn.lorawan.v3.Application.AttributesEntry) | 
  [Applications](#ttn.lorawan.v3.Applications) | 
  [CreateApplicationAPIKeyRequest](#ttn.lorawan.v3.CreateApplicationAPIKeyRequest) | 
  [CreateApplicationRequest](#ttn.lorawan.v3.CreateApplicationRequest) | 
  [GetApplicationAPIKeyRequest](#ttn.lorawan.v3.GetApplicationAPIKeyRequest) | 
  [GetApplicationCollaboratorRequest](#ttn.lorawan.v3.GetApplicationCollaboratorRequest) | 
  [GetApplicationRequest](#ttn.lorawan.v3.GetApplicationRequest) | 
  [ListApplicationAPIKeysRequest](#ttn.lorawan.v3.ListApplicationAPIKeysRequest) | 
  [ListApplicationCollaboratorsRequest](#ttn.lorawan.v3.ListApplicationCollaboratorsRequest) | 
  [ListApplicationsRequest](#ttn.lorawan.v3.ListApplicationsRequest) | By default we list all applications the caller has rights on. Set the user or the organization (not both) to instead list the applications where the user or organization is collaborator on.
  [SetApplicationCollaboratorRequest](#ttn.lorawan.v3.SetApplicationCollaboratorRequest) | 
  [UpdateApplicationAPIKeyRequest](#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest) | 
  [UpdateApplicationRequest](#ttn.lorawan.v3.UpdateApplicationRequest) | 
  [ApplicationLink](#ttn.lorawan.v3.ApplicationLink) | 
  [ApplicationLinkStats](#ttn.lorawan.v3.ApplicationLinkStats) | Link stats as monitored by the Application Server.
  [GetApplicationLinkRequest](#ttn.lorawan.v3.GetApplicationLinkRequest) | 
  [SetApplicationLinkRequest](#ttn.lorawan.v3.SetApplicationLinkRequest) | 
  [ApplicationPubSub](#ttn.lorawan.v3.ApplicationPubSub) | 
  [Message](#ttn.lorawan.v3.ApplicationPubSub.Message) | 
  [NATSProvider](#ttn.lorawan.v3.ApplicationPubSub.NATSProvider) | The NATS provider settings.
  [ApplicationPubSubFormats](#ttn.lorawan.v3.ApplicationPubSubFormats) | 
  [FormatsEntry](#ttn.lorawan.v3.ApplicationPubSubFormats.FormatsEntry) | 
  [ApplicationPubSubIdentifiers](#ttn.lorawan.v3.ApplicationPubSubIdentifiers) | 
  [ApplicationPubSubs](#ttn.lorawan.v3.ApplicationPubSubs) | 
  [GetApplicationPubSubRequest](#ttn.lorawan.v3.GetApplicationPubSubRequest) | 
  [ListApplicationPubSubsRequest](#ttn.lorawan.v3.ListApplicationPubSubsRequest) | 
  [SetApplicationPubSubRequest](#ttn.lorawan.v3.SetApplicationPubSubRequest) | 
  [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook) | 
  [HeadersEntry](#ttn.lorawan.v3.ApplicationWebhook.HeadersEntry) | 
  [Message](#ttn.lorawan.v3.ApplicationWebhook.Message) | 
  [ApplicationWebhookFormats](#ttn.lorawan.v3.ApplicationWebhookFormats) | 
  [FormatsEntry](#ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry) | 
  [ApplicationWebhookIdentifiers](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) | 
  [ApplicationWebhooks](#ttn.lorawan.v3.ApplicationWebhooks) | 
  [GetApplicationWebhookRequest](#ttn.lorawan.v3.GetApplicationWebhookRequest) | 
  [ListApplicationWebhooksRequest](#ttn.lorawan.v3.ListApplicationWebhooksRequest) | 
  [SetApplicationWebhookRequest](#ttn.lorawan.v3.SetApplicationWebhookRequest) | 
  [Client](#ttn.lorawan.v3.Client) | An OAuth client on the network.
  [AttributesEntry](#ttn.lorawan.v3.Client.AttributesEntry) | 
  [Clients](#ttn.lorawan.v3.Clients) | 
  [CreateClientRequest](#ttn.lorawan.v3.CreateClientRequest) | 
  [GetClientCollaboratorRequest](#ttn.lorawan.v3.GetClientCollaboratorRequest) | 
  [GetClientRequest](#ttn.lorawan.v3.GetClientRequest) | 
  [ListClientCollaboratorsRequest](#ttn.lorawan.v3.ListClientCollaboratorsRequest) | 
  [ListClientsRequest](#ttn.lorawan.v3.ListClientsRequest) | By default we list all OAuth clients the caller has rights on. Set the user or the organization (not both) to instead list the OAuth clients where the user or organization is collaborator on.
  [SetClientCollaboratorRequest](#ttn.lorawan.v3.SetClientCollaboratorRequest) | 
  [UpdateClientRequest](#ttn.lorawan.v3.UpdateClientRequest) | 
  [PeerInfo](#ttn.lorawan.v3.PeerInfo) | PeerInfo
  [TagsEntry](#ttn.lorawan.v3.PeerInfo.TagsEntry) | 
  [FrequencyPlanDescription](#ttn.lorawan.v3.FrequencyPlanDescription) | 
  [ListFrequencyPlansRequest](#ttn.lorawan.v3.ListFrequencyPlansRequest) | 
  [ListFrequencyPlansResponse](#ttn.lorawan.v3.ListFrequencyPlansResponse) | 
  [ContactInfo](#ttn.lorawan.v3.ContactInfo) | 
  [ContactInfoValidation](#ttn.lorawan.v3.ContactInfoValidation) | 
  [CreateEndDeviceRequest](#ttn.lorawan.v3.CreateEndDeviceRequest) | 
  [EndDevice](#ttn.lorawan.v3.EndDevice) | Defines an End Device registration and its state on the network. The persistence of the EndDevice is divided between the Network Server, Application Server and Join Server. SDKs are responsible for combining (if desired) the three.
  [AttributesEntry](#ttn.lorawan.v3.EndDevice.AttributesEntry) | 
  [LocationsEntry](#ttn.lorawan.v3.EndDevice.LocationsEntry) | 
  [EndDeviceBrand](#ttn.lorawan.v3.EndDeviceBrand) | 
  [EndDeviceModel](#ttn.lorawan.v3.EndDeviceModel) | 
  [EndDeviceVersion](#ttn.lorawan.v3.EndDeviceVersion) | Template for creating end devices.
  [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) | Identifies an end device model with version information.
  [EndDevices](#ttn.lorawan.v3.EndDevices) | 
  [GetEndDeviceRequest](#ttn.lorawan.v3.GetEndDeviceRequest) | 
  [ListEndDevicesRequest](#ttn.lorawan.v3.ListEndDevicesRequest) | 
  [MACParameters](#ttn.lorawan.v3.MACParameters) | MACParameters represent the parameters of the device's MAC layer (active or desired). This is used internally by the Network Server and is read only.
  [Channel](#ttn.lorawan.v3.MACParameters.Channel) | 
  [MACSettings](#ttn.lorawan.v3.MACSettings) | 
  [AggregatedDutyCycleValue](#ttn.lorawan.v3.MACSettings.AggregatedDutyCycleValue) | 
  [DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue) | 
  [PingSlotPeriodValue](#ttn.lorawan.v3.MACSettings.PingSlotPeriodValue) | 
  [RxDelayValue](#ttn.lorawan.v3.MACSettings.RxDelayValue) | 
  [MACState](#ttn.lorawan.v3.MACState) | MACState represents the state of MAC layer of the device. MACState is reset on each join for OTAA or ResetInd for ABP devices. This is used internally by the Network Server and is read only.
  [JoinAccept](#ttn.lorawan.v3.MACState.JoinAccept) | 
  [Session](#ttn.lorawan.v3.Session) | 
  [SetEndDeviceRequest](#ttn.lorawan.v3.SetEndDeviceRequest) | 
  [UpdateEndDeviceRequest](#ttn.lorawan.v3.UpdateEndDeviceRequest) | 
  [ErrorDetails](#ttn.lorawan.v3.ErrorDetails) | Error details that are communicated over gRPC (and HTTP) APIs. The messages (for translation) are stored as "error:<namespace>:<name>".
  [Event](#ttn.lorawan.v3.Event) | 
  [ContextEntry](#ttn.lorawan.v3.Event.ContextEntry) | 
  [StreamEventsRequest](#ttn.lorawan.v3.StreamEventsRequest) | 
  [CreateGatewayAPIKeyRequest](#ttn.lorawan.v3.CreateGatewayAPIKeyRequest) | 
  [CreateGatewayRequest](#ttn.lorawan.v3.CreateGatewayRequest) | 
  [Gateway](#ttn.lorawan.v3.Gateway) | Gateway is the message that defines a gateway on the network.
  [AttributesEntry](#ttn.lorawan.v3.Gateway.AttributesEntry) | 
  [GatewayAntenna](#ttn.lorawan.v3.GatewayAntenna) | GatewayAntenna is the message that defines a gateway antenna.
  [AttributesEntry](#ttn.lorawan.v3.GatewayAntenna.AttributesEntry) | 
  [GatewayBrand](#ttn.lorawan.v3.GatewayBrand) | 
  [GatewayConnectionStats](#ttn.lorawan.v3.GatewayConnectionStats) | Connection stats as monitored by the Gateway Server.
  [RoundTripTimes](#ttn.lorawan.v3.GatewayConnectionStats.RoundTripTimes) | 
  [GatewayModel](#ttn.lorawan.v3.GatewayModel) | 
  [GatewayRadio](#ttn.lorawan.v3.GatewayRadio) | 
  [TxConfiguration](#ttn.lorawan.v3.GatewayRadio.TxConfiguration) | 
  [GatewayStatus](#ttn.lorawan.v3.GatewayStatus) | 
  [MetricsEntry](#ttn.lorawan.v3.GatewayStatus.MetricsEntry) | 
  [VersionsEntry](#ttn.lorawan.v3.GatewayStatus.VersionsEntry) | 
  [GatewayVersion](#ttn.lorawan.v3.GatewayVersion) | Template for creating gateways.
  [GatewayVersionIdentifiers](#ttn.lorawan.v3.GatewayVersionIdentifiers) | Identifies an end device model with version information.
  [Gateways](#ttn.lorawan.v3.Gateways) | 
  [GetGatewayAPIKeyRequest](#ttn.lorawan.v3.GetGatewayAPIKeyRequest) | 
  [GetGatewayCollaboratorRequest](#ttn.lorawan.v3.GetGatewayCollaboratorRequest) | 
  [GetGatewayIdentifiersForEUIRequest](#ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest) | 
  [GetGatewayRequest](#ttn.lorawan.v3.GetGatewayRequest) | 
  [ListGatewayAPIKeysRequest](#ttn.lorawan.v3.ListGatewayAPIKeysRequest) | 
  [ListGatewayCollaboratorsRequest](#ttn.lorawan.v3.ListGatewayCollaboratorsRequest) | 
  [ListGatewaysRequest](#ttn.lorawan.v3.ListGatewaysRequest) | By default we list all gateways the caller has rights on. Set the user or the organization (not both) to instead list the gateways where the user or organization is collaborator on.
  [SetGatewayCollaboratorRequest](#ttn.lorawan.v3.SetGatewayCollaboratorRequest) | 
  [UpdateGatewayAPIKeyRequest](#ttn.lorawan.v3.UpdateGatewayAPIKeyRequest) | 
  [UpdateGatewayRequest](#ttn.lorawan.v3.UpdateGatewayRequest) | 
  [PullGatewayConfigurationRequest](#ttn.lorawan.v3.PullGatewayConfigurationRequest) | 
  [GatewayDown](#ttn.lorawan.v3.GatewayDown) | GatewayDown contains downlink messages for the gateway.
  [GatewayUp](#ttn.lorawan.v3.GatewayUp) | GatewayUp may contain zero or more uplink messages and/or a status message for the gateway.
  [ScheduleDownlinkErrorDetails](#ttn.lorawan.v3.ScheduleDownlinkErrorDetails) | 
  [ScheduleDownlinkResponse](#ttn.lorawan.v3.ScheduleDownlinkResponse) | 
  [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) | 
  [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) | 
  [CombinedIdentifiers](#ttn.lorawan.v3.CombinedIdentifiers) | Combine the identifiers of multiple entities. The main purpose of this message is its use in events.
  [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | 
  [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | EntityIdentifiers contains one of the possible entity identifiers.
  [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) | 
  [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) | 
  [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) | OrganizationOrUserIdentifiers contains either organization or user identifiers.
  [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) | 
  [AuthInfoResponse](#ttn.lorawan.v3.AuthInfoResponse) | 
  [APIKeyAccess](#ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess) | 
  [JoinRequest](#ttn.lorawan.v3.JoinRequest) | 
  [JoinResponse](#ttn.lorawan.v3.JoinResponse) | 
  [AppSKeyResponse](#ttn.lorawan.v3.AppSKeyResponse) | 
  [CryptoServicePayloadRequest](#ttn.lorawan.v3.CryptoServicePayloadRequest) | 
  [CryptoServicePayloadResponse](#ttn.lorawan.v3.CryptoServicePayloadResponse) | 
  [DeriveSessionKeysRequest](#ttn.lorawan.v3.DeriveSessionKeysRequest) | 
  [GetRootKeysRequest](#ttn.lorawan.v3.GetRootKeysRequest) | 
  [JoinAcceptMICRequest](#ttn.lorawan.v3.JoinAcceptMICRequest) | 
  [JoinEUIPrefix](#ttn.lorawan.v3.JoinEUIPrefix) | 
  [JoinEUIPrefixes](#ttn.lorawan.v3.JoinEUIPrefixes) | 
  [NwkSKeysResponse](#ttn.lorawan.v3.NwkSKeysResponse) | 
  [ProvisionEndDevicesRequest](#ttn.lorawan.v3.ProvisionEndDevicesRequest) | 
  [IdentifiersFromData](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData) | 
  [IdentifiersList](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList) | 
  [IdentifiersRange](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange) | 
  [SessionKeyRequest](#ttn.lorawan.v3.SessionKeyRequest) | 
  [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) | 
  [RootKeys](#ttn.lorawan.v3.RootKeys) | Root keys for a LoRaWAN device. These are stored on the Join Server.
  [SessionKeys](#ttn.lorawan.v3.SessionKeys) | Session keys for a LoRaWAN session. Only the components for which the keys were meant, will have the key-encryption-key (KEK) to decrypt the individual keys.
  [CFList](#ttn.lorawan.v3.CFList) | 
  [DLSettings](#ttn.lorawan.v3.DLSettings) | 
  [DataRate](#ttn.lorawan.v3.DataRate) | 
  [DownlinkPath](#ttn.lorawan.v3.DownlinkPath) | 
  [FCtrl](#ttn.lorawan.v3.FCtrl) | 
  [FHDR](#ttn.lorawan.v3.FHDR) | 
  [FSKDataRate](#ttn.lorawan.v3.FSKDataRate) | 
  [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers) | 
  [JoinAcceptPayload](#ttn.lorawan.v3.JoinAcceptPayload) | 
  [JoinRequestPayload](#ttn.lorawan.v3.JoinRequestPayload) | 
  [LoRaDataRate](#ttn.lorawan.v3.LoRaDataRate) | 
  [MACCommand](#ttn.lorawan.v3.MACCommand) | 
  [ADRParamSetupReq](#ttn.lorawan.v3.MACCommand.ADRParamSetupReq) | 
  [BeaconFreqAns](#ttn.lorawan.v3.MACCommand.BeaconFreqAns) | 
  [BeaconFreqReq](#ttn.lorawan.v3.MACCommand.BeaconFreqReq) | 
  [BeaconTimingAns](#ttn.lorawan.v3.MACCommand.BeaconTimingAns) | 
  [DLChannelAns](#ttn.lorawan.v3.MACCommand.DLChannelAns) | 
  [DLChannelReq](#ttn.lorawan.v3.MACCommand.DLChannelReq) | 
  [DevStatusAns](#ttn.lorawan.v3.MACCommand.DevStatusAns) | 
  [DeviceModeConf](#ttn.lorawan.v3.MACCommand.DeviceModeConf) | 
  [DeviceModeInd](#ttn.lorawan.v3.MACCommand.DeviceModeInd) | 
  [DeviceTimeAns](#ttn.lorawan.v3.MACCommand.DeviceTimeAns) | 
  [DutyCycleReq](#ttn.lorawan.v3.MACCommand.DutyCycleReq) | 
  [ForceRejoinReq](#ttn.lorawan.v3.MACCommand.ForceRejoinReq) | 
  [LinkADRAns](#ttn.lorawan.v3.MACCommand.LinkADRAns) | 
  [LinkADRReq](#ttn.lorawan.v3.MACCommand.LinkADRReq) | 
  [LinkCheckAns](#ttn.lorawan.v3.MACCommand.LinkCheckAns) | 
  [NewChannelAns](#ttn.lorawan.v3.MACCommand.NewChannelAns) | 
  [NewChannelReq](#ttn.lorawan.v3.MACCommand.NewChannelReq) | 
  [PingSlotChannelAns](#ttn.lorawan.v3.MACCommand.PingSlotChannelAns) | 
  [PingSlotChannelReq](#ttn.lorawan.v3.MACCommand.PingSlotChannelReq) | 
  [PingSlotInfoReq](#ttn.lorawan.v3.MACCommand.PingSlotInfoReq) | 
  [RejoinParamSetupAns](#ttn.lorawan.v3.MACCommand.RejoinParamSetupAns) | 
  [RejoinParamSetupReq](#ttn.lorawan.v3.MACCommand.RejoinParamSetupReq) | 
  [RekeyConf](#ttn.lorawan.v3.MACCommand.RekeyConf) | 
  [RekeyInd](#ttn.lorawan.v3.MACCommand.RekeyInd) | 
  [ResetConf](#ttn.lorawan.v3.MACCommand.ResetConf) | 
  [ResetInd](#ttn.lorawan.v3.MACCommand.ResetInd) | 
  [RxParamSetupAns](#ttn.lorawan.v3.MACCommand.RxParamSetupAns) | 
  [RxParamSetupReq](#ttn.lorawan.v3.MACCommand.RxParamSetupReq) | 
  [RxTimingSetupReq](#ttn.lorawan.v3.MACCommand.RxTimingSetupReq) | 
  [TxParamSetupReq](#ttn.lorawan.v3.MACCommand.TxParamSetupReq) | 
  [MACPayload](#ttn.lorawan.v3.MACPayload) | 
  [MHDR](#ttn.lorawan.v3.MHDR) | 
  [Message](#ttn.lorawan.v3.Message) | 
  [RejoinRequestPayload](#ttn.lorawan.v3.RejoinRequestPayload) | 
  [TxRequest](#ttn.lorawan.v3.TxRequest) | TxRequest is a request for transmission. If sent to a roaming partner, this request is used to generate the DLMetadata Object (see Backend Interfaces 1.0, Table 22). If the gateway has a scheduler, this request is sent to the gateway, in the order of gateway_ids. Otherwise, the Gateway Server attempts to schedule the request and creates the TxSettings.
  [TxSettings](#ttn.lorawan.v3.TxSettings) | TxSettings contains the settings for a transmission. This message is used on both uplink and downlink. On downlink, this is a scheduled transmission.
  [Downlink](#ttn.lorawan.v3.TxSettings.Downlink) | Transmission settings for downlink.
  [UplinkToken](#ttn.lorawan.v3.UplinkToken) | 
  [ProcessDownlinkMessageRequest](#ttn.lorawan.v3.ProcessDownlinkMessageRequest) | 
  [ProcessUplinkMessageRequest](#ttn.lorawan.v3.ProcessUplinkMessageRequest) | 
  [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | 
  [ClassBC](#ttn.lorawan.v3.ApplicationDownlink.ClassBC) | 
  [ApplicationDownlinkFailed](#ttn.lorawan.v3.ApplicationDownlinkFailed) | 
  [ApplicationDownlinks](#ttn.lorawan.v3.ApplicationDownlinks) | 
  [ApplicationInvalidatedDownlinks](#ttn.lorawan.v3.ApplicationInvalidatedDownlinks) | 
  [ApplicationJoinAccept](#ttn.lorawan.v3.ApplicationJoinAccept) | 
  [ApplicationLocation](#ttn.lorawan.v3.ApplicationLocation) | 
  [AttributesEntry](#ttn.lorawan.v3.ApplicationLocation.AttributesEntry) | 
  [ApplicationUp](#ttn.lorawan.v3.ApplicationUp) | 
  [ApplicationUplink](#ttn.lorawan.v3.ApplicationUplink) | 
  [DownlinkMessage](#ttn.lorawan.v3.DownlinkMessage) | Downlink message from the network to the end device
  [DownlinkQueueRequest](#ttn.lorawan.v3.DownlinkQueueRequest) | 
  [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters) | 
  [TxAcknowledgment](#ttn.lorawan.v3.TxAcknowledgment) | 
  [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | Uplink message from the end device to the network
  [Location](#ttn.lorawan.v3.Location) | 
  [RxMetadata](#ttn.lorawan.v3.RxMetadata) | Contains metadata for a received message. Each antenna that receives a message corresponds to one RxMetadata.
  [GenerateDevAddrResponse](#ttn.lorawan.v3.GenerateDevAddrResponse) | 
  [ListOAuthAccessTokensRequest](#ttn.lorawan.v3.ListOAuthAccessTokensRequest) | 
  [ListOAuthClientAuthorizationsRequest](#ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest) | 
  [OAuthAccessToken](#ttn.lorawan.v3.OAuthAccessToken) | 
  [OAuthAccessTokenIdentifiers](#ttn.lorawan.v3.OAuthAccessTokenIdentifiers) | 
  [OAuthAccessTokens](#ttn.lorawan.v3.OAuthAccessTokens) | 
  [OAuthAuthorizationCode](#ttn.lorawan.v3.OAuthAuthorizationCode) | 
  [OAuthClientAuthorization](#ttn.lorawan.v3.OAuthClientAuthorization) | 
  [OAuthClientAuthorizationIdentifiers](#ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers) | 
  [OAuthClientAuthorizations](#ttn.lorawan.v3.OAuthClientAuthorizations) | 
  [CreateOrganizationAPIKeyRequest](#ttn.lorawan.v3.CreateOrganizationAPIKeyRequest) | 
  [CreateOrganizationRequest](#ttn.lorawan.v3.CreateOrganizationRequest) | 
  [GetOrganizationAPIKeyRequest](#ttn.lorawan.v3.GetOrganizationAPIKeyRequest) | 
  [GetOrganizationCollaboratorRequest](#ttn.lorawan.v3.GetOrganizationCollaboratorRequest) | 
  [GetOrganizationRequest](#ttn.lorawan.v3.GetOrganizationRequest) | 
  [ListOrganizationAPIKeysRequest](#ttn.lorawan.v3.ListOrganizationAPIKeysRequest) | 
  [ListOrganizationCollaboratorsRequest](#ttn.lorawan.v3.ListOrganizationCollaboratorsRequest) | 
  [ListOrganizationsRequest](#ttn.lorawan.v3.ListOrganizationsRequest) | By default we list all organizations the caller has rights on. Set the user to instead list the organizations where the user or organization is collaborator on.
  [Organization](#ttn.lorawan.v3.Organization) | 
  [AttributesEntry](#ttn.lorawan.v3.Organization.AttributesEntry) | 
  [Organizations](#ttn.lorawan.v3.Organizations) | 
  [SetOrganizationCollaboratorRequest](#ttn.lorawan.v3.SetOrganizationCollaboratorRequest) | 
  [UpdateOrganizationAPIKeyRequest](#ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest) | 
  [UpdateOrganizationRequest](#ttn.lorawan.v3.UpdateOrganizationRequest) | 
  [ConcentratorConfig](#ttn.lorawan.v3.ConcentratorConfig) | 
  [Channel](#ttn.lorawan.v3.ConcentratorConfig.Channel) | 
  [FSKChannel](#ttn.lorawan.v3.ConcentratorConfig.FSKChannel) | 
  [LBTConfiguration](#ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration) | 
  [LoRaStandardChannel](#ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel) | 
  [APIKey](#ttn.lorawan.v3.APIKey) | 
  [APIKeys](#ttn.lorawan.v3.APIKeys) | 
  [Collaborator](#ttn.lorawan.v3.Collaborator) | 
  [Collaborators](#ttn.lorawan.v3.Collaborators) | 
  [GetCollaboratorResponse](#ttn.lorawan.v3.GetCollaboratorResponse) | 
  [Rights](#ttn.lorawan.v3.Rights) | 
  [SearchEndDevicesRequest](#ttn.lorawan.v3.SearchEndDevicesRequest) | 
  [AttributesContainEntry](#ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry) | 
  [SearchEntitiesRequest](#ttn.lorawan.v3.SearchEntitiesRequest) | This message is used for finding entities in the EntityRegistrySearch service.
  [AttributesContainEntry](#ttn.lorawan.v3.SearchEntitiesRequest.AttributesContainEntry) | 
  [CreateTemporaryPasswordRequest](#ttn.lorawan.v3.CreateTemporaryPasswordRequest) | 
  [CreateUserAPIKeyRequest](#ttn.lorawan.v3.CreateUserAPIKeyRequest) | 
  [CreateUserRequest](#ttn.lorawan.v3.CreateUserRequest) | 
  [DeleteInvitationRequest](#ttn.lorawan.v3.DeleteInvitationRequest) | 
  [GetUserAPIKeyRequest](#ttn.lorawan.v3.GetUserAPIKeyRequest) | 
  [GetUserRequest](#ttn.lorawan.v3.GetUserRequest) | 
  [Invitation](#ttn.lorawan.v3.Invitation) | 
  [Invitations](#ttn.lorawan.v3.Invitations) | 
  [ListInvitationsRequest](#ttn.lorawan.v3.ListInvitationsRequest) | 
  [ListUserAPIKeysRequest](#ttn.lorawan.v3.ListUserAPIKeysRequest) | 
  [ListUserSessionsRequest](#ttn.lorawan.v3.ListUserSessionsRequest) | 
  [Picture](#ttn.lorawan.v3.Picture) | 
  [Embedded](#ttn.lorawan.v3.Picture.Embedded) | 
  [SizesEntry](#ttn.lorawan.v3.Picture.SizesEntry) | 
  [SendInvitationRequest](#ttn.lorawan.v3.SendInvitationRequest) | 
  [UpdateUserAPIKeyRequest](#ttn.lorawan.v3.UpdateUserAPIKeyRequest) | 
  [UpdateUserPasswordRequest](#ttn.lorawan.v3.UpdateUserPasswordRequest) | 
  [UpdateUserRequest](#ttn.lorawan.v3.UpdateUserRequest) | 
  [User](#ttn.lorawan.v3.User) | User is the message that defines an user on the network.
  [AttributesEntry](#ttn.lorawan.v3.User.AttributesEntry) | 
  [UserSession](#ttn.lorawan.v3.UserSession) | 
  [UserSessionIdentifiers](#ttn.lorawan.v3.UserSessionIdentifiers) | 
  [UserSessions](#ttn.lorawan.v3.UserSessions) | 
  [Users](#ttn.lorawan.v3.Users) | 
  



 
 

## <a name="ttn.lorawan.v3.Application">Application</a>

  Application is the message that defines an Application in the network.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
description | [string](#string) |  |  | <p>`string.max_len`: `2000`</p>
attributes | [Application.AttributesEntry](#ttn.lorawan.v3.Application.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  | 

## <a name="ttn.lorawan.v3.Application.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.Applications">Applications</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
applications | [Application](#ttn.lorawan.v3.Application) | repeated |  | 

## <a name="ttn.lorawan.v3.CreateApplicationAPIKeyRequest">CreateApplicationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | <p>`repeated.items.enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.CreateApplicationRequest">CreateApplicationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application | [Application](#ttn.lorawan.v3.Application) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created application. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.GetApplicationAPIKeyRequest">GetApplicationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
key_id | [string](#string) |  | Unique public identifier for the API key. | 

## <a name="ttn.lorawan.v3.GetApplicationCollaboratorRequest">GetApplicationCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.GetApplicationRequest">GetApplicationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.ListApplicationAPIKeysRequest">ListApplicationAPIKeysRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListApplicationCollaboratorsRequest">ListApplicationCollaboratorsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListApplicationsRequest">ListApplicationsRequest</a>

  By default we list all applications the caller has rights on.
Set the user or the organization (not both) to instead list the applications
where the user or organization is collaborator on.

Field | Type | Label | Description | Validation
---|---|---|---|---
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | 
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.SetApplicationCollaboratorRequest">SetApplicationCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateApplicationAPIKeyRequest">UpdateApplicationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateApplicationRequest">UpdateApplicationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application | [Application](#ttn.lorawan.v3.Application) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 
 

## <a name="ttn.lorawan.v3.ApplicationLink">ApplicationLink</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
network_server_address | [string](#string) |  | The address of the external Network Server where to link to. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. Leave empty when linking to a cluster Network Server. | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p>
api_key | [string](#string) |  |  | <p>`string.min_len`: `1`</p>
default_formatters | [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationLinkStats">ApplicationLinkStats</a>

  Link stats as monitored by the Application Server.

Field | Type | Label | Description | Validation
---|---|---|---|---
linked_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
network_server_address | [string](#string) |  |  | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p>
last_up_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the last upstream message has been received from a Network Server. This can be a join-accept, uplink message or downlink message event. | 
up_count | [uint64](#uint64) |  | Number of upstream messages received. | 
last_downlink_forwarded_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the last downlink message has been forwarded to a Network Server. | 
downlink_count | [uint64](#uint64) |  | Number of downlink messages forwarded. | 

## <a name="ttn.lorawan.v3.GetApplicationLinkRequest">GetApplicationLinkRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.SetApplicationLinkRequest">SetApplicationLinkRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
link | [ApplicationLink](#ttn.lorawan.v3.ApplicationLink) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 

## <a name="ttn.lorawan.v3.ApplicationPubSub">ApplicationPubSub</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [ApplicationPubSubIdentifiers](#ttn.lorawan.v3.ApplicationPubSubIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
format | [string](#string) |  | The format to use for the body. Supported values depend on the Application Server configuration. | <p>`string.max_len`: `10`</p>
nats | [ApplicationPubSub.NATSProvider](#ttn.lorawan.v3.ApplicationPubSub.NATSProvider) |  |  | 
base_topic | [string](#string) |  | Base topic name to which the messages topic is appended. | <p>`string.max_len`: `100`</p>
downlink_push | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  | The topic to which the Application Server subscribes for downlink queue push operations. | 
downlink_replace | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  | The topic to which the Application Server subscribes for downlink queue replace operations. | 
uplink_message | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 
join_accept | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 
downlink_ack | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 
downlink_nack | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 
downlink_sent | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 
downlink_failed | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 
downlink_queued | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 
location_solved | [ApplicationPubSub.Message](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationPubSub.Message">Message</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
topic | [string](#string) |  | The topic on which the Application Server publishes or receives the messages. | <p>`string.max_len`: `100`</p>

## <a name="ttn.lorawan.v3.ApplicationPubSub.NATSProvider">NATSProvider</a>

  The NATS provider settings.

Field | Type | Label | Description | Validation
---|---|---|---|---
server_url | [string](#string) |  | The server connection URL. | <p>`string.uri`: `true`</p>

## <a name="ttn.lorawan.v3.ApplicationPubSubFormats">ApplicationPubSubFormats</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
formats | [ApplicationPubSubFormats.FormatsEntry](#ttn.lorawan.v3.ApplicationPubSubFormats.FormatsEntry) | repeated | Format and description. | 

## <a name="ttn.lorawan.v3.ApplicationPubSubFormats.FormatsEntry">FormatsEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationPubSubIdentifiers">ApplicationPubSubIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
pub_sub_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>

## <a name="ttn.lorawan.v3.ApplicationPubSubs">ApplicationPubSubs</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
pubsubs | [ApplicationPubSub](#ttn.lorawan.v3.ApplicationPubSub) | repeated |  | 

## <a name="ttn.lorawan.v3.GetApplicationPubSubRequest">GetApplicationPubSubRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [ApplicationPubSubIdentifiers](#ttn.lorawan.v3.ApplicationPubSubIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.ListApplicationPubSubsRequest">ListApplicationPubSubsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.SetApplicationPubSubRequest">SetApplicationPubSubRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
pubsub | [ApplicationPubSub](#ttn.lorawan.v3.ApplicationPubSub) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 

## <a name="ttn.lorawan.v3.ApplicationWebhook">ApplicationWebhook</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [ApplicationWebhookIdentifiers](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
base_url | [string](#string) |  | Base URL to which the message's path is appended. | <p>`string.uri`: `true`</p>
headers | [ApplicationWebhook.HeadersEntry](#ttn.lorawan.v3.ApplicationWebhook.HeadersEntry) | repeated | HTTP headers to use. | 
format | [string](#string) |  | The format to use for the body. Supported values depend on the Application Server configuration. | 
uplink_message | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 
join_accept | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 
downlink_ack | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 
downlink_nack | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 
downlink_sent | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 
downlink_failed | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 
downlink_queued | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 
location_solved | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationWebhook.HeadersEntry">HeadersEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationWebhook.Message">Message</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
path | [string](#string) |  | Path to append to the base URL. | 

## <a name="ttn.lorawan.v3.ApplicationWebhookFormats">ApplicationWebhookFormats</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
formats | [ApplicationWebhookFormats.FormatsEntry](#ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry) | repeated | Format and description. | 

## <a name="ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry">FormatsEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationWebhookIdentifiers">ApplicationWebhookIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
webhook_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>

## <a name="ttn.lorawan.v3.ApplicationWebhooks">ApplicationWebhooks</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
webhooks | [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook) | repeated |  | 

## <a name="ttn.lorawan.v3.GetApplicationWebhookRequest">GetApplicationWebhookRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [ApplicationWebhookIdentifiers](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.ListApplicationWebhooksRequest">ListApplicationWebhooksRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.SetApplicationWebhookRequest">SetApplicationWebhookRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
webhook | [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 

## <a name="ttn.lorawan.v3.Client">Client</a>

  An OAuth client on the network.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
description | [string](#string) |  |  | <p>`string.max_len`: `2000`</p>
attributes | [Client.AttributesEntry](#ttn.lorawan.v3.Client.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  | 
secret | [string](#string) |  | The client secret is only visible to collaborators of the client. | 
redirect_uris | [string](#string) | repeated | The allowed redirect URIs against which authorization requests are checked. If the authorization request does not pass a redirect URI, the first one from this list is taken. | 
state | [State](#ttn.lorawan.v3.State) |  | The reviewing state of the client. This field can only be modified by admins. | <p>`enum.defined_only`: `true`</p>
skip_authorization | [bool](#bool) |  | If set, the authorization page will be skipped. This field can only be modified by admins. | 
endorsed | [bool](#bool) |  | If set, the authorization page will show endorsement. This field can only be modified by admins. | 
grants | [GrantType](#ttn.lorawan.v3.GrantType) | repeated | OAuth flows that can be used for the client to get a token. After a client is created, this field can only be modified by admins. | <p>`repeated.items.enum.defined_only`: `true`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated | Rights denotes what rights the client will have access to. Users that previously authorized this client will have to re-authorize the client after rights are added to this list. | <p>`repeated.items.enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.Client.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.Clients">Clients</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
clients | [Client](#ttn.lorawan.v3.Client) | repeated |  | 

## <a name="ttn.lorawan.v3.CreateClientRequest">CreateClientRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
client | [Client](#ttn.lorawan.v3.Client) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created client. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.GetClientCollaboratorRequest">GetClientCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.GetClientRequest">GetClientRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.ListClientCollaboratorsRequest">ListClientCollaboratorsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListClientsRequest">ListClientsRequest</a>

  By default we list all OAuth clients the caller has rights on.
Set the user or the organization (not both) to instead list the OAuth clients
where the user or organization is collaborator on.

Field | Type | Label | Description | Validation
---|---|---|---|---
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | 
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.SetClientCollaboratorRequest">SetClientCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateClientRequest">UpdateClientRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
client | [Client](#ttn.lorawan.v3.Client) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 
 

## <a name="ttn.lorawan.v3.PeerInfo">PeerInfo</a>

  PeerInfo

Field | Type | Label | Description | Validation
---|---|---|---|---
grpc_port | [uint32](#uint32) |  | Port on which the gRPC server is exposed. | 
tls | [bool](#bool) |  | Indicates whether the gRPC server uses TLS. | 
roles | [PeerInfo.Role](#ttn.lorawan.v3.PeerInfo.Role) | repeated | Roles of the peer. | 
tags | [PeerInfo.TagsEntry](#ttn.lorawan.v3.PeerInfo.TagsEntry) | repeated | Tags of the peer | 

## <a name="ttn.lorawan.v3.PeerInfo.TagsEntry">TagsEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 
 

## <a name="ttn.lorawan.v3.FrequencyPlanDescription">FrequencyPlanDescription</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
id | [string](#string) |  |  | 
base_id | [string](#string) |  | The ID of the frequency that the current frequency plan is based on. | 
name | [string](#string) |  |  | 
base_frequency | [uint32](#uint32) |  | Base frequency in MHz for hardware support (433, 470, 868 or 915) | 

## <a name="ttn.lorawan.v3.ListFrequencyPlansRequest">ListFrequencyPlansRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
base_frequency | [uint32](#uint32) |  | Optional base frequency in MHz for hardware support (433, 470, 868 or 915) | 

## <a name="ttn.lorawan.v3.ListFrequencyPlansResponse">ListFrequencyPlansResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency_plans | [FrequencyPlanDescription](#ttn.lorawan.v3.FrequencyPlanDescription) | repeated |  | 
 

## <a name="ttn.lorawan.v3.ContactInfo">ContactInfo</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
contact_type | [ContactType](#ttn.lorawan.v3.ContactType) |  |  | 
contact_method | [ContactMethod](#ttn.lorawan.v3.ContactMethod) |  |  | 
value | [string](#string) |  |  | 
public | [bool](#bool) |  |  | 
validated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 

## <a name="ttn.lorawan.v3.ContactInfoValidation">ContactInfoValidation</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
id | [string](#string) |  |  | 
token | [string](#string) |  |  | 
entity | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) |  |  | 
contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  | 
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
 

## <a name="ttn.lorawan.v3.CreateEndDeviceRequest">CreateEndDeviceRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
end_device | [EndDevice](#ttn.lorawan.v3.EndDevice) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.EndDevice">EndDevice</a>

  Defines an End Device registration and its state on the network.
The persistence of the EndDevice is divided between the Network Server, Application Server and Join Server.
SDKs are responsible for combining (if desired) the three.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
name | [string](#string) |  | Friendly name of the device. Stored in Entity Registry. | <p>`string.max_len`: `50`</p>
description | [string](#string) |  | Description of the device. Stored in Entity Registry. | <p>`string.max_len`: `2000`</p>
attributes | [EndDevice.AttributesEntry](#ttn.lorawan.v3.EndDevice.AttributesEntry) | repeated | Attributes of the device. Stored in Entity Registry. | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
version_ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | Version Identifiers. Stored in Entity Registry, Network Server and Application Server. | 
service_profile_id | [string](#string) |  | Default service profile. Stored in Entity Registry. | <p>`string.max_len`: `64`</p>
network_server_address | [string](#string) |  | The address of the Network Server where this device is supposed to be registered. Stored in Entity Registry and Join Server. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p>
application_server_address | [string](#string) |  | The address of the Application Server where this device is supposed to be registered. Stored in Entity Registry and Join Server. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p>
join_server_address | [string](#string) |  | The address of the Join Server where this device is supposed to be registered. Stored in Entity Registry. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p>
locations | [EndDevice.LocationsEntry](#ttn.lorawan.v3.EndDevice.LocationsEntry) | repeated | Location of the device. Stored in Entity Registry. | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
supports_class_b | [bool](#bool) |  | Whether the device supports class B. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | 
supports_class_c | [bool](#bool) |  | Whether the device supports class C. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | 
lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | <p>`enum.defined_only`: `true`</p>
lorawan_phy_version | [PHYVersion](#ttn.lorawan.v3.PHYVersion) |  | LoRaWAN PHY version. Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | <p>`enum.defined_only`: `true`</p>
frequency_plan_id | [string](#string) |  | ID of the frequency plan used by this device. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | <p>`string.max_len`: `64`</p>
min_frequency | [uint64](#uint64) |  | Minimum frequency the device is capable of using (Hz). Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | 
max_frequency | [uint64](#uint64) |  | Maximum frequency the device is capable of using (Hz). Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | 
supports_join | [bool](#bool) |  | The device supports join (it's OTAA). Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | 
resets_join_nonces | [bool](#bool) |  | Whether the device resets the join and dev nonces (not LoRaWAN 1.1 compliant). Stored in Join Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. | 
root_keys | [RootKeys](#ttn.lorawan.v3.RootKeys) |  | Device root keys. Stored in Join Server. | 
net_id | [bytes](#bytes) |  | Home NetID. Stored in Join Server. | 
mac_settings | [MACSettings](#ttn.lorawan.v3.MACSettings) |  | Settings for how the Network Server handles MAC layer for this device. Stored in Network Server. | 
mac_state | [MACState](#ttn.lorawan.v3.MACState) |  | MAC state of the device. Stored in Network Server. | 
pending_mac_state | [MACState](#ttn.lorawan.v3.MACState) |  | Pending MAC state of the device. Stored in Network Server. | 
session | [Session](#ttn.lorawan.v3.Session) |  | Current session of the device. Stored in Network Server and Application Server. | 
pending_session | [Session](#ttn.lorawan.v3.Session) |  | Pending session. Stored in Network Server and Application Server until RekeyInd is received. | 
last_dev_nonce | [uint32](#uint32) |  | Last DevNonce used. This field is only used for devices using LoRaWAN version 1.1 and later. Stored in Join Server. | 
used_dev_nonces | [uint32](#uint32) | repeated | Used DevNonces sorted in ascending order. This field is only used for devices using LoRaWAN versions preceding 1.1. Stored in Join Server. | 
last_join_nonce | [uint32](#uint32) |  | Last JoinNonce/AppNonce(for devices using LoRaWAN versions preceding 1.1) used. Stored in Join Server. | 
last_rj_count_0 | [uint32](#uint32) |  | Last Rejoin counter value used (type 0/2). Stored in Join Server. | 
last_rj_count_1 | [uint32](#uint32) |  | Last Rejoin counter value used (type 1). Stored in Join Server. | 
last_dev_status_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when last DevStatus MAC command was received. Stored in Network Server. | 
power_state | [PowerState](#ttn.lorawan.v3.PowerState) |  | The power state of the device; whether it is battery-powered or connected to an external power source. Received via the DevStatus MAC command at status_received_at. Stored in Network Server. | <p>`enum.defined_only`: `true`</p>
battery_percentage | [google.protobuf.FloatValue](#google.protobuf.FloatValue) |  | Latest-known battery percentage of the device. Received via the DevStatus MAC command at last_dev_status_received_at or earlier. Stored in Network Server. | <p>`float.lte`: `1`</p><p>`float.gte`: `0`</p>
downlink_margin | [int32](#int32) |  | Demodulation signal-to-noise ratio (dB). Received via the DevStatus MAC command at last_dev_status_received_at. Stored in Network Server. | 
recent_adr_uplinks | [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | repeated | Recent uplink messages with ADR bit set to 1 sorted by time. Stored in Network Server. The field is reset each time an uplink message carrying MACPayload is received with ADR bit set to 0. The number of messages stored is in the range [0,20]; | 
recent_uplinks | [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | repeated | Recent uplink messages sorted by time. Stored in Network Server. The number of messages stored may depend on configuration. | 
recent_downlinks | [DownlinkMessage](#ttn.lorawan.v3.DownlinkMessage) | repeated | Recent downlink messages sorted by time. Stored in Network Server. The number of messages stored may depend on configuration. | 
queued_application_downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated | Queued Application downlink messages. Stored in Application Server, which sets them on the Network Server. | 
formatters | [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters) |  | The payload formatters for this end device. Stored in Application Server. Copied on creation from template identified by version_ids. | 
provisioner_id | [string](#string) |  | ID of the provisioner. Stored in Join Server. | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p>
provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  | Vendor-specific provisioning data. Stored in Join Server. | 
multicast | [bool](#bool) |  | Indicates whether this device represents a multicast group. | 

## <a name="ttn.lorawan.v3.EndDevice.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.EndDevice.LocationsEntry">LocationsEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [Location](#ttn.lorawan.v3.Location) |  |  | 

## <a name="ttn.lorawan.v3.EndDeviceBrand">EndDeviceBrand</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
id | [string](#string) |  |  | 
name | [string](#string) |  |  | 
url | [string](#string) |  |  | 
logos | [string](#string) | repeated | Logos contains file names of brand logos. | 

## <a name="ttn.lorawan.v3.EndDeviceModel">EndDeviceModel</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
brand_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
name | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.EndDeviceVersion">EndDeviceVersion</a>

  Template for creating end devices.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | Version identifiers. | <p>`message.required`: `true`</p>
lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. | <p>`enum.defined_only`: `true`</p>
lorawan_phy_version | [PHYVersion](#ttn.lorawan.v3.PHYVersion) |  | LoRaWAN PHY version. | <p>`enum.defined_only`: `true`</p>
frequency_plan_id | [string](#string) |  | ID of the frequency plan used by this device. | <p>`string.max_len`: `64`</p>
photos | [string](#string) | repeated | Photos contains file names of device photos. | 
supports_class_b | [bool](#bool) |  | Whether the device supports class B. | 
supports_class_c | [bool](#bool) |  | Whether the device supports class C. | 
default_mac_settings | [MACSettings](#ttn.lorawan.v3.MACSettings) |  | Default MAC layer settings of the device. | 
min_frequency | [uint64](#uint64) |  | Minimum frequency the device is capable of using (Hz). | 
max_frequency | [uint64](#uint64) |  | Maximum frequency the device is capable of using (Hz). | 
supports_join | [bool](#bool) |  | The device supports join (it's OTAA). | 
resets_join_nonces | [bool](#bool) |  | Whether the device resets the join and dev nonces (not LoRaWAN 1.1 compliant). | 
default_formatters | [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters) |  | Default formatters defining the payload formats for this end device. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.EndDeviceVersionIdentifiers">EndDeviceVersionIdentifiers</a>

  Identifies an end device model with version information.

Field | Type | Label | Description | Validation
---|---|---|---|---
brand_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
model_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
hardware_version | [string](#string) |  |  | 
firmware_version | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.EndDevices">EndDevices</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
end_devices | [EndDevice](#ttn.lorawan.v3.EndDevice) | repeated |  | 

## <a name="ttn.lorawan.v3.GetEndDeviceRequest">GetEndDeviceRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.ListEndDevicesRequest">ListEndDevicesRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.MACParameters">MACParameters</a>

  MACParameters represent the parameters of the device's MAC layer (active or desired).
This is used internally by the Network Server and is read only.

Field | Type | Label | Description | Validation
---|---|---|---|---
max_eirp | [float](#float) |  | Maximum EIRP (dBm). | 
adr_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | ADR: data rate index to use. | <p>`enum.defined_only`: `true`</p>
adr_tx_power_index | [uint32](#uint32) |  | ADR: transmission power index to use. | <p>`uint32.lte`: `15`</p>
adr_nb_trans | [uint32](#uint32) |  | ADR: number of retransmissions. | <p>`uint32.lte`: `15`</p>
adr_ack_limit | [uint32](#uint32) |  | ADR: number of messages to wait before setting ADRAckReq. | <p>`uint32.lte`: `32768`</p><p>`uint32.gte`: `1`</p>
adr_ack_delay | [uint32](#uint32) |  | ADR: number of messages to wait after setting ADRAckReq and before changing TxPower or DataRate. | <p>`uint32.lte`: `32768`</p><p>`uint32.gte`: `1`</p>
rx1_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  | Rx1 delay (Rx2 delay is Rx1 delay + 1 second). | <p>`enum.defined_only`: `true`</p>
rx1_data_rate_offset | [uint32](#uint32) |  | Data rate offset for Rx1. | <p>`uint32.lte`: `7`</p>
rx2_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Data rate index for Rx2. | <p>`enum.defined_only`: `true`</p>
rx2_frequency | [uint64](#uint64) |  | Frequency for Rx2 (Hz). | <p>`uint64.gte`: `100000`</p>
max_duty_cycle | [AggregatedDutyCycle](#ttn.lorawan.v3.AggregatedDutyCycle) |  | Maximum uplink duty cycle (of all channels). | <p>`enum.defined_only`: `true`</p>
rejoin_time_periodicity | [RejoinTimeExponent](#ttn.lorawan.v3.RejoinTimeExponent) |  | Time within which a rejoin-request must be sent. | <p>`enum.defined_only`: `true`</p>
rejoin_count_periodicity | [RejoinCountExponent](#ttn.lorawan.v3.RejoinCountExponent) |  | Message count within which a rejoin-request must be sent. | <p>`enum.defined_only`: `true`</p>
ping_slot_frequency | [uint64](#uint64) |  | Frequency of the class B ping slot (Hz). | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p>
ping_slot_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Data rate index of the class B ping slot. | <p>`enum.defined_only`: `true`</p>
beacon_frequency | [uint64](#uint64) |  | Frequency of the class B beacon (Hz). | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p>
channels | [MACParameters.Channel](#ttn.lorawan.v3.MACParameters.Channel) | repeated | Configured uplink channels and optionally Rx1 frequency. | <p>`repeated.min_items`: `1`</p>
uplink_dwell_time | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether uplink dwell time is set (400ms). If this field is not set, then the value is either unknown or irrelevant(Network Server cannot modify it). | 
downlink_dwell_time | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether downlink dwell time is set (400ms). If this field is not set, then the value is either unknown or irrelevant(Network Server cannot modify it). | 

## <a name="ttn.lorawan.v3.MACParameters.Channel">Channel</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
uplink_frequency | [uint64](#uint64) |  | Uplink frequency of the channel (Hz). | <p>`uint64.gte`: `100000`</p>
downlink_frequency | [uint64](#uint64) |  | Downlink frequency of the channel (Hz). | <p>`uint64.gte`: `100000`</p>
min_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Index of the minimum data rate for uplink. | <p>`enum.defined_only`: `true`</p>
max_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Index of the maximum data rate for uplink. | <p>`enum.defined_only`: `true`</p>
enable_uplink | [bool](#bool) |  | Channel can be used by device for uplink. | 

## <a name="ttn.lorawan.v3.MACSettings">MACSettings</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
class_b_timeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | Maximum delay for the device to answer a MAC request or a confirmed downlink frame. If unset, the default value from Network Server configuration will be used. | 
ping_slot_periodicity | [MACSettings.PingSlotPeriodValue](#ttn.lorawan.v3.MACSettings.PingSlotPeriodValue) |  | Periodicity of the class B ping slot. If unset, the default value from Network Server configuration will be used. | 
ping_slot_data_rate_index | [MACSettings.DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue) |  | Data rate index of the class B ping slot. If unset, the default value from Network Server configuration will be used. | 
ping_slot_frequency | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  | Frequency of the class B ping slot (Hz). If unset, the default value from Network Server configuration will be used. | <p>`uint64.gte`: `100000`</p>
class_c_timeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | Maximum delay for the device to answer a MAC request or a confirmed downlink frame. If unset, the default value from Network Server configuration will be used. | 
rx1_delay | [MACSettings.RxDelayValue](#ttn.lorawan.v3.MACSettings.RxDelayValue) |  | Class A Rx1 delay. If unset, the default value from Network Server configuration or regional parameters specification will be used. | 
rx1_data_rate_offset | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  | Rx1 data rate offset. If unset, the default value from Network Server configuration will be used. | <p>`uint32.lte`: `7`</p>
rx2_data_rate_index | [MACSettings.DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue) |  | Data rate index for Rx2. If unset, the default value from Network Server configuration or regional parameters specification will be used. | 
rx2_frequency | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  | Frequency for Rx2 (Hz). If unset, the default value from Network Server configuration or regional parameters specification will be used. | <p>`uint64.gte`: `100000`</p>
factory_preset_frequencies | [uint64](#uint64) | repeated | List of factory-preset frequencies. If unset, the default value from Network Server configuration or regional parameters specification will be used. | 
max_duty_cycle | [MACSettings.AggregatedDutyCycleValue](#ttn.lorawan.v3.MACSettings.AggregatedDutyCycleValue) |  | Maximum uplink duty cycle (of all channels). | 
supports_32_bit_f_cnt | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether the device supports 32-bit frame counters. If unset, the default value from Network Server configuration will be used. | 
use_adr | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether the Network Server should use ADR for the device. If unset, the default value from Network Server configuration will be used. | 
adr_margin | [google.protobuf.FloatValue](#google.protobuf.FloatValue) |  | The ADR margin tells the network server how much margin it should add in ADR requests. A bigger margin is less efficient, but gives a better chance of successful reception. If unset, the default value from Network Server configuration will be used. | 
resets_f_cnt | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether the device resets the frame counters (not LoRaWAN compliant). If unset, the default value from Network Server configuration will be used. | 
status_time_periodicity | [google.protobuf.Duration](#google.protobuf.Duration) |  | The interval after which a DevStatusReq MACCommand shall be sent. If unset, the default value from Network Server configuration will be used. | 
status_count_periodicity | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  | Number of uplink messages after which a DevStatusReq MACCommand shall be sent. If unset, the default value from Network Server configuration will be used. | 
desired_rx1_delay | [MACSettings.RxDelayValue](#ttn.lorawan.v3.MACSettings.RxDelayValue) |  | The Rx1 delay Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from Network Server configuration or regional parameters specification will be used. | 
desired_rx1_data_rate_offset | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  | The Rx1 data rate offset Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from Network Server configuration will be used. | 
desired_rx2_data_rate_index | [MACSettings.DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue) |  | The Rx2 data rate index Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from frequency plan, Network Server configuration or regional parameters specification will be used. | 
desired_rx2_frequency | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  | The Rx2 frequency index Network Server should configure device to use via MAC commands. If unset, the default value from frequency plan, Network Server configuration or regional parameters specification will be used. | <p>`uint64.gte`: `100000`</p>

## <a name="ttn.lorawan.v3.MACSettings.AggregatedDutyCycleValue">AggregatedDutyCycleValue</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
value | [AggregatedDutyCycle](#ttn.lorawan.v3.AggregatedDutyCycle) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACSettings.DataRateIndexValue">DataRateIndexValue</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
value | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACSettings.PingSlotPeriodValue">PingSlotPeriodValue</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
value | [PingSlotPeriod](#ttn.lorawan.v3.PingSlotPeriod) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACSettings.RxDelayValue">RxDelayValue</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
value | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACState">MACState</a>

  MACState represents the state of MAC layer of the device.
MACState is reset on each join for OTAA or ResetInd for ABP devices.
This is used internally by the Network Server and is read only.

Field | Type | Label | Description | Validation
---|---|---|---|---
current_parameters | [MACParameters](#ttn.lorawan.v3.MACParameters) |  | Current LoRaWAN MAC parameters. | <p>`message.required`: `true`</p>
desired_parameters | [MACParameters](#ttn.lorawan.v3.MACParameters) |  | Desired LoRaWAN MAC parameters. | <p>`message.required`: `true`</p>
device_class | [Class](#ttn.lorawan.v3.Class) |  | Currently active LoRaWAN device class - Device class is A by default - If device sets ClassB bit in uplink, this will be set to B - If device sent DeviceModeInd MAC message, this will be set to that value | <p>`enum.defined_only`: `true`</p>
lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. | <p>`enum.defined_only`: `true`</p>
last_confirmed_downlink_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when the last confirmed downlink message or MAC command was scheduled. | 
last_dev_status_f_cnt_up | [uint32](#uint32) |  | Frame counter value of last uplink containing DevStatusAns. | 
ping_slot_periodicity | [PingSlotPeriod](#ttn.lorawan.v3.PingSlotPeriod) |  | Periodicity of the class B ping slot. | <p>`enum.defined_only`: `true`</p>
pending_application_downlink | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  | A confirmed application downlink, for which an acknowledgment is expected to arrive. | 
queued_responses | [MACCommand](#ttn.lorawan.v3.MACCommand) | repeated | Queued MAC responses. Regenerated on each uplink. | 
pending_requests | [MACCommand](#ttn.lorawan.v3.MACCommand) | repeated | Pending MAC requests(i.e. sent requests, for which no response has been received yet). Regenerated on each downlink. | 
queued_join_accept | [MACState.JoinAccept](#ttn.lorawan.v3.MACState.JoinAccept) |  | Queued join-accept. Set each time a (re-)join request accept is received from Join Server and removed each time a downlink is scheduled. | 
pending_join_request | [JoinRequest](#ttn.lorawan.v3.JoinRequest) |  | Pending join request. Set each time a join accept is scheduled and removed each time an uplink is received from the device. | 
rx_windows_available | [bool](#bool) |  | Whether or not Rx windows are expected to be open. Set to true every time an uplink is received. Set to false every time a successful downlink scheduling attempt is made. | 

## <a name="ttn.lorawan.v3.MACState.JoinAccept">JoinAccept</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
payload | [bytes](#bytes) |  | Payload of the join-accept received from Join Server. | <p>`bytes.min_len`: `17`</p><p>`bytes.max_len`: `33`</p>
request | [JoinRequest](#ttn.lorawan.v3.JoinRequest) |  | JoinRequest sent to Join Server. | <p>`message.required`: `true`</p>
keys | [SessionKeys](#ttn.lorawan.v3.SessionKeys) |  | Network session keys associated with the join. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.Session">Session</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
dev_addr | [bytes](#bytes) |  | Device Address, issued by the Network Server or chosen by device manufacturer in case of testing range (beginning with 00-03). Known by Network Server, Application Server and Join Server. Owned by Network Server. | 
keys | [SessionKeys](#ttn.lorawan.v3.SessionKeys) |  |  | <p>`message.required`: `true`</p>
last_f_cnt_up | [uint32](#uint32) |  | Last uplink frame counter value used. Network Server only. Application Server assumes the Network Server checked it. | 
last_n_f_cnt_down | [uint32](#uint32) |  | Last network downlink frame counter value used. Network Server only. | 
last_a_f_cnt_down | [uint32](#uint32) |  | Last application downlink frame counter value used. Application Server only. | 
last_conf_f_cnt_down | [uint32](#uint32) |  | Frame counter of the last confirmed downlink message sent. Network Server only. | 
started_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when the session started. Network Server only. | 

## <a name="ttn.lorawan.v3.SetEndDeviceRequest">SetEndDeviceRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
end_device | [EndDevice](#ttn.lorawan.v3.EndDevice) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.UpdateEndDeviceRequest">UpdateEndDeviceRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
end_device | [EndDevice](#ttn.lorawan.v3.EndDevice) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 
 
 

## <a name="ttn.lorawan.v3.ErrorDetails">ErrorDetails</a>

  Error details that are communicated over gRPC (and HTTP) APIs.
The messages (for translation) are stored as "error:<namespace>:<name>".

Field | Type | Label | Description | Validation
---|---|---|---|---
namespace | [string](#string) |  | Namespace of the error (typically the package name in the stack). | 
name | [string](#string) |  | Name of the error. | 
message_format | [string](#string) |  | The default (fallback) message format that should be used for the error. This is also used if the client does not have a translation for the error. | 
attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | Attributes that should be filled into the message format. Any extra attributes can be displayed as error details. | 
correlation_id | [string](#string) |  | The correlation ID of the error can be used to correlate the error to stack traces the network may (or may not) store about recent errors. | 
cause | [ErrorDetails](#ttn.lorawan.v3.ErrorDetails) |  | The error that caused this error. | 
code | [uint32](#uint32) |  | The status code of the error. | 
details | [google.protobuf.Any](#google.protobuf.Any) | repeated | The details of the error. | 
 

## <a name="ttn.lorawan.v3.Event">Event</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
name | [string](#string) |  |  | 
time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | <p>`message.required`: `true`</p>
identifiers | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | repeated |  | 
data | [google.protobuf.Any](#google.protobuf.Any) |  |  | 
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>
origin | [string](#string) |  |  | 
context | [Event.ContextEntry](#ttn.lorawan.v3.Event.ContextEntry) | repeated |  | 

## <a name="ttn.lorawan.v3.Event.ContextEntry">ContextEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [bytes](#bytes) |  |  | 

## <a name="ttn.lorawan.v3.StreamEventsRequest">StreamEventsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
identifiers | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | repeated |  | 
tail | [uint32](#uint32) |  | If greater than zero, this will return historical events, up to this maximum when the stream starts. If used in combination with "after", the limit that is reached first, is used. The availability of historical events depends on server support and retention policy. | 
after | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | If not empty, this will return historical events after the given time when the stream starts. If used in combination with "tail", the limit that is reached first, is used. The availability of historical events depends on server support and retention policy. | 
 

## <a name="ttn.lorawan.v3.CreateGatewayAPIKeyRequest">CreateGatewayAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | <p>`repeated.items.enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.CreateGatewayRequest">CreateGatewayRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway | [Gateway](#ttn.lorawan.v3.Gateway) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created gateway. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.Gateway">Gateway</a>

  Gateway is the message that defines a gateway on the network.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
description | [string](#string) |  |  | <p>`string.max_len`: `2000`</p>
attributes | [Gateway.AttributesEntry](#ttn.lorawan.v3.Gateway.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  | 
version_ids | [GatewayVersionIdentifiers](#ttn.lorawan.v3.GatewayVersionIdentifiers) |  |  | <p>`message.required`: `true`</p>
gateway_server_address | [string](#string) |  | The address of the Gateway Server to connect to. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p>
auto_update | [bool](#bool) |  |  | 
update_channel | [string](#string) |  |  | 
frequency_plan_id | [string](#string) |  |  | <p>`string.max_len`: `64`</p>
antennas | [GatewayAntenna](#ttn.lorawan.v3.GatewayAntenna) | repeated |  | 
status_public | [bool](#bool) |  | The status of this gateway may be publicly displayed. | 
location_public | [bool](#bool) |  | The location of this gateway may be publicly displayed. | 
schedule_downlink_late | [bool](#bool) |  | Enable server-side buffering of downlink messages. This is recommended for gateways using the Semtech UDP Packet Forwarder v2.x or older, as it does not feature a just-in-time queue. If enabled, the Gateway Server schedules the downlink message late to the gateway so that it does not overwrite previously scheduled downlink messages that have not been transmitted yet. | 
enforce_duty_cycle | [bool](#bool) |  | Enforcing gateway duty cycle is recommended for all gateways to respect spectrum regulations. Disable enforcing the duty cycle only in controlled research and development environments. | 
downlink_path_constraint | [DownlinkPathConstraint](#ttn.lorawan.v3.DownlinkPathConstraint) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.Gateway.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.GatewayAntenna">GatewayAntenna</a>

  GatewayAntenna is the message that defines a gateway antenna.

Field | Type | Label | Description | Validation
---|---|---|---|---
gain | [float](#float) |  | gain is the antenna gain relative to this gateway, in dBi. | 
location | [Location](#ttn.lorawan.v3.Location) |  | location is the antenna's location. | <p>`message.required`: `true`</p>
attributes | [GatewayAntenna.AttributesEntry](#ttn.lorawan.v3.GatewayAntenna.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>

## <a name="ttn.lorawan.v3.GatewayAntenna.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.GatewayBrand">GatewayBrand</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
id | [string](#string) |  |  | 
name | [string](#string) |  |  | 
url | [string](#string) |  |  | 
logos | [string](#string) | repeated | Logos contains file names of brand logos. | 

## <a name="ttn.lorawan.v3.GatewayConnectionStats">GatewayConnectionStats</a>

  Connection stats as monitored by the Gateway Server.

Field | Type | Label | Description | Validation
---|---|---|---|---
connected_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
protocol | [string](#string) |  | Protocol used to connect (for example, udp, mqtt, grpc) | 
last_status_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
last_status | [GatewayStatus](#ttn.lorawan.v3.GatewayStatus) |  |  | 
last_uplink_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
uplink_count | [uint64](#uint64) |  |  | 
last_downlink_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
downlink_count | [uint64](#uint64) |  |  | 
round_trip_times | [GatewayConnectionStats.RoundTripTimes](#ttn.lorawan.v3.GatewayConnectionStats.RoundTripTimes) |  |  | 

## <a name="ttn.lorawan.v3.GatewayConnectionStats.RoundTripTimes">RoundTripTimes</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
min | [google.protobuf.Duration](#google.protobuf.Duration) |  |  | <p>`message.required`: `true`</p>
max | [google.protobuf.Duration](#google.protobuf.Duration) |  |  | <p>`message.required`: `true`</p>
median | [google.protobuf.Duration](#google.protobuf.Duration) |  |  | <p>`message.required`: `true`</p>
count | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.GatewayModel">GatewayModel</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
brand_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
name | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.GatewayRadio">GatewayRadio</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
enable | [bool](#bool) |  |  | 
chip_type | [string](#string) |  |  | 
frequency | [uint64](#uint64) |  |  | 
rssi_offset | [float](#float) |  |  | 
tx_configuration | [GatewayRadio.TxConfiguration](#ttn.lorawan.v3.GatewayRadio.TxConfiguration) |  |  | 

## <a name="ttn.lorawan.v3.GatewayRadio.TxConfiguration">TxConfiguration</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
min_frequency | [uint64](#uint64) |  |  | 
max_frequency | [uint64](#uint64) |  |  | 
notch_frequency | [uint64](#uint64) |  |  | 

## <a name="ttn.lorawan.v3.GatewayStatus">GatewayStatus</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Current time of the gateway | <p>`message.required`: `true`</p>
boot_time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Boot time of the gateway - can be left out to save bandwidth; old value will be kept | 
versions | [GatewayStatus.VersionsEntry](#ttn.lorawan.v3.GatewayStatus.VersionsEntry) | repeated | Versions of gateway subsystems - each field can be left out to save bandwidth; old value will be kept - map keys are written in snake_case - for example: firmware: "2.0.4" forwarder: "v2-3.3.1" fpga: "48" dsp: "27" hal: "v2-3.5.0" | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
antenna_locations | [Location](#ttn.lorawan.v3.Location) | repeated | Location of each gateway's antenna - if left out, server uses registry-set location as fallback | 
ip | [string](#string) | repeated | IP addresses of this gateway. Repeated addresses can be used to communicate addresses of multiple interfaces (LAN, Public IP, ...). | 
metrics | [GatewayStatus.MetricsEntry](#ttn.lorawan.v3.GatewayStatus.MetricsEntry) | repeated | Metrics - can be used for forwarding gateway metrics such as temperatures or performance metrics - map keys are written in snake_case | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
advanced | [google.protobuf.Struct](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case | 

## <a name="ttn.lorawan.v3.GatewayStatus.MetricsEntry">MetricsEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [float](#float) |  |  | 

## <a name="ttn.lorawan.v3.GatewayStatus.VersionsEntry">VersionsEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.GatewayVersion">GatewayVersion</a>

  Template for creating gateways.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [GatewayVersionIdentifiers](#ttn.lorawan.v3.GatewayVersionIdentifiers) |  | Version identifiers. | <p>`message.required`: `true`</p>
photos | [string](#string) | repeated | Photos contains file names of gateway photos. | 
radios | [GatewayRadio](#ttn.lorawan.v3.GatewayRadio) | repeated |  | 
clock_source | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.GatewayVersionIdentifiers">GatewayVersionIdentifiers</a>

  Identifies an end device model with version information.

Field | Type | Label | Description | Validation
---|---|---|---|---
brand_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
model_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
hardware_version | [string](#string) |  |  | 
firmware_version | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.Gateways">Gateways</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateways | [Gateway](#ttn.lorawan.v3.Gateway) | repeated |  | 

## <a name="ttn.lorawan.v3.GetGatewayAPIKeyRequest">GetGatewayAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
key_id | [string](#string) |  | Unique public identifier for the API key. | 

## <a name="ttn.lorawan.v3.GetGatewayCollaboratorRequest">GetGatewayCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest">GetGatewayIdentifiersForEUIRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
eui | [bytes](#bytes) |  |  | 

## <a name="ttn.lorawan.v3.GetGatewayRequest">GetGatewayRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.ListGatewayAPIKeysRequest">ListGatewayAPIKeysRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListGatewayCollaboratorsRequest">ListGatewayCollaboratorsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListGatewaysRequest">ListGatewaysRequest</a>

  By default we list all gateways the caller has rights on.
Set the user or the organization (not both) to instead list the gateways
where the user or organization is collaborator on.

Field | Type | Label | Description | Validation
---|---|---|---|---
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | 
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.SetGatewayCollaboratorRequest">SetGatewayCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateGatewayAPIKeyRequest">UpdateGatewayAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateGatewayRequest">UpdateGatewayRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway | [Gateway](#ttn.lorawan.v3.Gateway) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 

## <a name="ttn.lorawan.v3.PullGatewayConfigurationRequest">PullGatewayConfigurationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | 
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 

## <a name="ttn.lorawan.v3.GatewayDown">GatewayDown</a>

  GatewayDown contains downlink messages for the gateway.

Field | Type | Label | Description | Validation
---|---|---|---|---
downlink_message | [DownlinkMessage](#ttn.lorawan.v3.DownlinkMessage) |  | DownlinkMessage for the gateway. | 

## <a name="ttn.lorawan.v3.GatewayUp">GatewayUp</a>

  GatewayUp may contain zero or more uplink messages and/or a status message for the gateway.

Field | Type | Label | Description | Validation
---|---|---|---|---
uplink_messages | [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | repeated | UplinkMessages received by the gateway. | 
gateway_status | [GatewayStatus](#ttn.lorawan.v3.GatewayStatus) |  |  | 
tx_acknowledgment | [TxAcknowledgment](#ttn.lorawan.v3.TxAcknowledgment) |  |  | 

## <a name="ttn.lorawan.v3.ScheduleDownlinkErrorDetails">ScheduleDownlinkErrorDetails</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
path_errors | [ErrorDetails](#ttn.lorawan.v3.ErrorDetails) | repeated |  | 

## <a name="ttn.lorawan.v3.ScheduleDownlinkResponse">ScheduleDownlinkResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
delay | [google.protobuf.Duration](#google.protobuf.Duration) |  |  | <p>`message.required`: `true`</p>
 

## <a name="ttn.lorawan.v3.ApplicationIdentifiers">ApplicationIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>

## <a name="ttn.lorawan.v3.ClientIdentifiers">ClientIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
client_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>

## <a name="ttn.lorawan.v3.CombinedIdentifiers">CombinedIdentifiers</a>

  Combine the identifiers of multiple entities.
The main purpose of this message is its use in events.

Field | Type | Label | Description | Validation
---|---|---|---|---
entity_identifiers | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | repeated |  | 

## <a name="ttn.lorawan.v3.EndDeviceIdentifiers">EndDeviceIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
device_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
dev_eui | [bytes](#bytes) |  | The LoRaWAN DevEUI. | 
join_eui | [bytes](#bytes) |  | The LoRaWAN JoinEUI (AppEUI until LoRaWAN 1.0.3 end devices). | 
dev_addr | [bytes](#bytes) |  | The LoRaWAN DevAddr. | 

## <a name="ttn.lorawan.v3.EntityIdentifiers">EntityIdentifiers</a>

  EntityIdentifiers contains one of the possible entity identifiers.

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | 
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | 
device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | 
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | 
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | 
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | 

## <a name="ttn.lorawan.v3.GatewayIdentifiers">GatewayIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
eui | [bytes](#bytes) |  | Secondary identifier, which can only be used in specific requests. | 

## <a name="ttn.lorawan.v3.OrganizationIdentifiers">OrganizationIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_id | [string](#string) |  | This ID shares namespace with user IDs. | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>

## <a name="ttn.lorawan.v3.OrganizationOrUserIdentifiers">OrganizationOrUserIdentifiers</a>

  OrganizationOrUserIdentifiers contains either organization or user identifiers.

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | 
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | 

## <a name="ttn.lorawan.v3.UserIdentifiers">UserIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_id | [string](#string) |  | This ID shares namespace with organization IDs. | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
email | [string](#string) |  | Secondary identifier, which can only be used in specific requests. | 
 

## <a name="ttn.lorawan.v3.AuthInfoResponse">AuthInfoResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
api_key | [AuthInfoResponse.APIKeyAccess](#ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess) |  |  | 
oauth_access_token | [OAuthAccessToken](#ttn.lorawan.v3.OAuthAccessToken) |  |  | 
universal_rights | [Rights](#ttn.lorawan.v3.Rights) |  |  | 
is_admin | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess">APIKeyAccess</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  | <p>`message.required`: `true`</p>
entity_ids | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) |  |  | <p>`message.required`: `true`</p>
 

## <a name="ttn.lorawan.v3.JoinRequest">JoinRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
raw_payload | [bytes](#bytes) |  |  | 
payload | [Message](#ttn.lorawan.v3.Message) |  |  | 
dev_addr | [bytes](#bytes) |  |  | 
selected_mac_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  |  | 
net_id | [bytes](#bytes) |  |  | 
downlink_settings | [DLSettings](#ttn.lorawan.v3.DLSettings) |  |  | <p>`message.required`: `true`</p>
rx_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  | <p>`enum.defined_only`: `true`</p>
cf_list | [CFList](#ttn.lorawan.v3.CFList) |  | Optional CFList. | 
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>

## <a name="ttn.lorawan.v3.JoinResponse">JoinResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
raw_payload | [bytes](#bytes) |  |  | <p>`bytes.min_len`: `17`</p><p>`bytes.max_len`: `33`</p>
session_keys | [SessionKeys](#ttn.lorawan.v3.SessionKeys) |  |  | <p>`message.required`: `true`</p>
lifetime | [google.protobuf.Duration](#google.protobuf.Duration) |  |  | 
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>
 

## <a name="ttn.lorawan.v3.AppSKeyResponse">AppSKeyResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
app_s_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Session Key. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.CryptoServicePayloadRequest">CryptoServicePayloadRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  |  | <p>`enum.defined_only`: `true`</p>
payload | [bytes](#bytes) |  |  | 
provisioner_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p>
provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  |  | 

## <a name="ttn.lorawan.v3.CryptoServicePayloadResponse">CryptoServicePayloadResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
payload | [bytes](#bytes) |  |  | 

## <a name="ttn.lorawan.v3.DeriveSessionKeysRequest">DeriveSessionKeysRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  |  | <p>`enum.defined_only`: `true`</p>
join_nonce | [bytes](#bytes) |  |  | 
dev_nonce | [bytes](#bytes) |  |  | 
net_id | [bytes](#bytes) |  |  | 
provisioner_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p>
provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  |  | 

## <a name="ttn.lorawan.v3.GetRootKeysRequest">GetRootKeysRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
provisioner_id | [string](#string) |  |  | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p>
provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  |  | 

## <a name="ttn.lorawan.v3.JoinAcceptMICRequest">JoinAcceptMICRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
payload_request | [CryptoServicePayloadRequest](#ttn.lorawan.v3.CryptoServicePayloadRequest) |  |  | <p>`message.required`: `true`</p>
join_request_type | [RejoinType](#ttn.lorawan.v3.RejoinType) |  |  | <p>`enum.defined_only`: `true`</p>
dev_nonce | [bytes](#bytes) |  |  | 

## <a name="ttn.lorawan.v3.JoinEUIPrefix">JoinEUIPrefix</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
join_eui | [bytes](#bytes) |  |  | 
length | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.JoinEUIPrefixes">JoinEUIPrefixes</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
prefixes | [JoinEUIPrefix](#ttn.lorawan.v3.JoinEUIPrefix) | repeated |  | 

## <a name="ttn.lorawan.v3.NwkSKeysResponse">NwkSKeysResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
f_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Forwarding Network Session Integrity Key (or Network Session Key in 1.0 compatibility mode). | <p>`message.required`: `true`</p>
s_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Serving Network Session Integrity Key. | <p>`message.required`: `true`</p>
nwk_s_enc_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Session Encryption Key. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest">ProvisionEndDevicesRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | 
provisioner_id | [string](#string) |  | ID of the provisioner service as configured in the Join Server. | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
provisioning_data | [bytes](#bytes) |  | Vendor-specific provisioning data. | 
list | [ProvisionEndDevicesRequest.IdentifiersList](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList) |  | List of device identifiers that will be provisioned. The device identifiers must contain device_id and dev_eui. If set, the application_ids must equal the provision request's application_ids. The number of entries in data must match the number of given identifiers. | 
range | [ProvisionEndDevicesRequest.IdentifiersRange](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange) |  | Provision devices in a range. The device_id will be generated by the provisioner from the vendor-specific data. The dev_eui will be issued from the given start_dev_eui. | 
from_data | [ProvisionEndDevicesRequest.IdentifiersFromData](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData) |  | Provision devices with identifiers from the given data. The device_id and dev_eui will be generated by the provisioner from the vendor-specific data. | 

## <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData">IdentifiersFromData</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
join_eui | [bytes](#bytes) |  |  | 

## <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList">IdentifiersList</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
join_eui | [bytes](#bytes) |  |  | 
end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | repeated |  | 

## <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange">IdentifiersRange</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
join_eui | [bytes](#bytes) |  |  | 
start_dev_eui | [bytes](#bytes) |  | DevEUI to start issuing from. | 

## <a name="ttn.lorawan.v3.SessionKeyRequest">SessionKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys. | <p>`bytes.max_len`: `2048`</p>
dev_eui | [bytes](#bytes) |  | LoRaWAN DevEUI. | 
join_eui | [bytes](#bytes) |  | The LoRaWAN JoinEUI (AppEUI until LoRaWAN 1.0.3 end devices). | 
 

## <a name="ttn.lorawan.v3.KeyEnvelope">KeyEnvelope</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [bytes](#bytes) |  | The unencrypted AES key. | 
kek_label | [string](#string) |  | The label of the RFC 3394 key-encryption-key (KEK) that was used to encrypt the key. | 
encrypted_key | [bytes](#bytes) |  |  | 

## <a name="ttn.lorawan.v3.RootKeys">RootKeys</a>

  Root keys for a LoRaWAN device.
These are stored on the Join Server.

Field | Type | Label | Description | Validation
---|---|---|---|---
root_key_id | [string](#string) |  | Join Server issued identifier for the root keys. | <p>`string.max_len`: `2048`</p>
app_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Key. | 
nwk_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Key. | 

## <a name="ttn.lorawan.v3.SessionKeys">SessionKeys</a>

  Session keys for a LoRaWAN session.
Only the components for which the keys were meant, will have the key-encryption-key (KEK) to decrypt the individual keys.

Field | Type | Label | Description | Validation
---|---|---|---|---
session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys. This ID can be used to request the keys from the Join Server in case the are lost. | <p>`bytes.max_len`: `2048`</p>
f_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Forwarding Network Session Integrity Key (or Network Session Key in 1.0 compatibility mode). This key is stored by the (forwarding) Network Server. | 
s_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Serving Network Session Integrity Key. This key is stored by the (serving) Network Server. | 
nwk_s_enc_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Session Encryption Key. This key is stored by the (serving) Network Server. | 
app_s_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Session Key. This key is stored by the Application Server. | 
 

## <a name="ttn.lorawan.v3.CFList">CFList</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
type | [CFListType](#ttn.lorawan.v3.CFListType) |  |  | <p>`enum.defined_only`: `true`</p>
freq | [uint32](#uint32) | repeated | Frequencies to be broadcasted, in hecto-Hz. These values are broadcasted as 24 bits unsigned integers. This field should not contain default values. | 
ch_masks | [bool](#bool) | repeated | ChMasks controlling the channels to be used. Length of this field must be equal to the amount of uplink channels defined by the selected frequency plan. | 

## <a name="ttn.lorawan.v3.DLSettings">DLSettings</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
rx1_dr_offset | [uint32](#uint32) |  |  | <p>`uint32.lte`: `7`</p>
rx2_dr | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>
opt_neg | [bool](#bool) |  | OptNeg is set if Network Server implements LoRaWAN 1.1 or greater. | 

## <a name="ttn.lorawan.v3.DataRate">DataRate</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
lora | [LoRaDataRate](#ttn.lorawan.v3.LoRaDataRate) |  |  | 
fsk | [FSKDataRate](#ttn.lorawan.v3.FSKDataRate) |  |  | 

## <a name="ttn.lorawan.v3.DownlinkPath">DownlinkPath</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
uplink_token | [bytes](#bytes) |  |  | 
fixed | [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers) |  |  | 

## <a name="ttn.lorawan.v3.FCtrl">FCtrl</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
adr | [bool](#bool) |  |  | 
adr_ack_req | [bool](#bool) |  | Only on uplink. | 
ack | [bool](#bool) |  |  | 
f_pending | [bool](#bool) |  | Only on downlink. | 
class_b | [bool](#bool) |  | Only on uplink. | 

## <a name="ttn.lorawan.v3.FHDR">FHDR</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
dev_addr | [bytes](#bytes) |  |  | 
f_ctrl | [FCtrl](#ttn.lorawan.v3.FCtrl) |  |  | <p>`message.required`: `true`</p>
f_cnt | [uint32](#uint32) |  |  | <p>`uint32.lte`: `65535`</p>
f_opts | [bytes](#bytes) |  |  | <p>`bytes.max_len`: `15`</p>

## <a name="ttn.lorawan.v3.FSKDataRate">FSKDataRate</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
bit_rate | [uint32](#uint32) |  | Bit rate (bps). | 

## <a name="ttn.lorawan.v3.GatewayAntennaIdentifiers">GatewayAntennaIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
antenna_index | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.JoinAcceptPayload">JoinAcceptPayload</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
encrypted | [bytes](#bytes) |  |  | 
join_nonce | [bytes](#bytes) |  |  | 
net_id | [bytes](#bytes) |  |  | 
dev_addr | [bytes](#bytes) |  |  | 
dl_settings | [DLSettings](#ttn.lorawan.v3.DLSettings) |  |  | <p>`message.required`: `true`</p>
rx_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  | <p>`enum.defined_only`: `true`</p>
cf_list | [CFList](#ttn.lorawan.v3.CFList) |  |  | 

## <a name="ttn.lorawan.v3.JoinRequestPayload">JoinRequestPayload</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
join_eui | [bytes](#bytes) |  |  | 
dev_eui | [bytes](#bytes) |  |  | 
dev_nonce | [bytes](#bytes) |  |  | 

## <a name="ttn.lorawan.v3.LoRaDataRate">LoRaDataRate</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
bandwidth | [uint32](#uint32) |  | Bandwidth (Hz). | 
spreading_factor | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand">MACCommand</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
cid | [MACCommandIdentifier](#ttn.lorawan.v3.MACCommandIdentifier) |  |  | <p>`enum.defined_only`: `true`</p>
raw_payload | [bytes](#bytes) |  |  | 
reset_ind | [MACCommand.ResetInd](#ttn.lorawan.v3.MACCommand.ResetInd) |  |  | 
reset_conf | [MACCommand.ResetConf](#ttn.lorawan.v3.MACCommand.ResetConf) |  |  | 
link_check_ans | [MACCommand.LinkCheckAns](#ttn.lorawan.v3.MACCommand.LinkCheckAns) |  |  | 
link_adr_req | [MACCommand.LinkADRReq](#ttn.lorawan.v3.MACCommand.LinkADRReq) |  |  | 
link_adr_ans | [MACCommand.LinkADRAns](#ttn.lorawan.v3.MACCommand.LinkADRAns) |  |  | 
duty_cycle_req | [MACCommand.DutyCycleReq](#ttn.lorawan.v3.MACCommand.DutyCycleReq) |  |  | 
rx_param_setup_req | [MACCommand.RxParamSetupReq](#ttn.lorawan.v3.MACCommand.RxParamSetupReq) |  |  | 
rx_param_setup_ans | [MACCommand.RxParamSetupAns](#ttn.lorawan.v3.MACCommand.RxParamSetupAns) |  |  | 
dev_status_ans | [MACCommand.DevStatusAns](#ttn.lorawan.v3.MACCommand.DevStatusAns) |  |  | 
new_channel_req | [MACCommand.NewChannelReq](#ttn.lorawan.v3.MACCommand.NewChannelReq) |  |  | 
new_channel_ans | [MACCommand.NewChannelAns](#ttn.lorawan.v3.MACCommand.NewChannelAns) |  |  | 
dl_channel_req | [MACCommand.DLChannelReq](#ttn.lorawan.v3.MACCommand.DLChannelReq) |  |  | 
dl_channel_ans | [MACCommand.DLChannelAns](#ttn.lorawan.v3.MACCommand.DLChannelAns) |  |  | 
rx_timing_setup_req | [MACCommand.RxTimingSetupReq](#ttn.lorawan.v3.MACCommand.RxTimingSetupReq) |  |  | 
tx_param_setup_req | [MACCommand.TxParamSetupReq](#ttn.lorawan.v3.MACCommand.TxParamSetupReq) |  |  | 
rekey_ind | [MACCommand.RekeyInd](#ttn.lorawan.v3.MACCommand.RekeyInd) |  |  | 
rekey_conf | [MACCommand.RekeyConf](#ttn.lorawan.v3.MACCommand.RekeyConf) |  |  | 
adr_param_setup_req | [MACCommand.ADRParamSetupReq](#ttn.lorawan.v3.MACCommand.ADRParamSetupReq) |  |  | 
device_time_ans | [MACCommand.DeviceTimeAns](#ttn.lorawan.v3.MACCommand.DeviceTimeAns) |  |  | 
force_rejoin_req | [MACCommand.ForceRejoinReq](#ttn.lorawan.v3.MACCommand.ForceRejoinReq) |  |  | 
rejoin_param_setup_req | [MACCommand.RejoinParamSetupReq](#ttn.lorawan.v3.MACCommand.RejoinParamSetupReq) |  |  | 
rejoin_param_setup_ans | [MACCommand.RejoinParamSetupAns](#ttn.lorawan.v3.MACCommand.RejoinParamSetupAns) |  |  | 
ping_slot_info_req | [MACCommand.PingSlotInfoReq](#ttn.lorawan.v3.MACCommand.PingSlotInfoReq) |  |  | 
ping_slot_channel_req | [MACCommand.PingSlotChannelReq](#ttn.lorawan.v3.MACCommand.PingSlotChannelReq) |  |  | 
ping_slot_channel_ans | [MACCommand.PingSlotChannelAns](#ttn.lorawan.v3.MACCommand.PingSlotChannelAns) |  |  | 
beacon_timing_ans | [MACCommand.BeaconTimingAns](#ttn.lorawan.v3.MACCommand.BeaconTimingAns) |  |  | 
beacon_freq_req | [MACCommand.BeaconFreqReq](#ttn.lorawan.v3.MACCommand.BeaconFreqReq) |  |  | 
beacon_freq_ans | [MACCommand.BeaconFreqAns](#ttn.lorawan.v3.MACCommand.BeaconFreqAns) |  |  | 
device_mode_ind | [MACCommand.DeviceModeInd](#ttn.lorawan.v3.MACCommand.DeviceModeInd) |  |  | 
device_mode_conf | [MACCommand.DeviceModeConf](#ttn.lorawan.v3.MACCommand.DeviceModeConf) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.ADRParamSetupReq">ADRParamSetupReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
adr_ack_limit_exponent | [ADRAckLimitExponent](#ttn.lorawan.v3.ADRAckLimitExponent) |  | Exponent e that configures the ADR_ACK_LIMIT = 2^e messages. | <p>`enum.defined_only`: `true`</p>
adr_ack_delay_exponent | [ADRAckDelayExponent](#ttn.lorawan.v3.ADRAckDelayExponent) |  | Exponent e that configures the ADR_ACK_DELAY = 2^e messages. | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.BeaconFreqAns">BeaconFreqAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency_ack | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.BeaconFreqReq">BeaconFreqReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency | [uint64](#uint64) |  | Frequency of the Class B beacons (Hz). | <p>`uint64.gte`: `100000`</p>

## <a name="ttn.lorawan.v3.MACCommand.BeaconTimingAns">BeaconTimingAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
delay | [uint32](#uint32) |  | (uint16) See LoRaWAN specification. | <p>`uint32.lte`: `65535`</p>
channel_index | [uint32](#uint32) |  |  | <p>`uint32.lte`: `255`</p>

## <a name="ttn.lorawan.v3.MACCommand.DLChannelAns">DLChannelAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
channel_index_ack | [bool](#bool) |  |  | 
frequency_ack | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.DLChannelReq">DLChannelReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
channel_index | [uint32](#uint32) |  |  | <p>`uint32.lte`: `255`</p>
frequency | [uint64](#uint64) |  | Downlink channel frequency (Hz). | <p>`uint64.gte`: `100000`</p>

## <a name="ttn.lorawan.v3.MACCommand.DevStatusAns">DevStatusAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
battery | [uint32](#uint32) |  | Device battery status. 0 indicates that the device is connected to an external power source. 1..254 indicates a battery level. 255 indicates that the device was not able to measure the battery level. | <p>`uint32.lte`: `255`</p>
margin | [int32](#int32) |  | SNR of the last downlink (dB; [-32, +31]). | <p>`int32.lte`: `31`</p><p>`int32.gte`: `-32`</p>

## <a name="ttn.lorawan.v3.MACCommand.DeviceModeConf">DeviceModeConf</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
class | [Class](#ttn.lorawan.v3.Class) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.DeviceModeInd">DeviceModeInd</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
class | [Class](#ttn.lorawan.v3.Class) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.DeviceTimeAns">DeviceTimeAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.DutyCycleReq">DutyCycleReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
max_duty_cycle | [AggregatedDutyCycle](#ttn.lorawan.v3.AggregatedDutyCycle) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.ForceRejoinReq">ForceRejoinReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
rejoin_type | [RejoinType](#ttn.lorawan.v3.RejoinType) |  |  | <p>`enum.defined_only`: `true`</p>
data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>
max_retries | [uint32](#uint32) |  |  | <p>`uint32.lte`: `7`</p>
period_exponent | [RejoinPeriodExponent](#ttn.lorawan.v3.RejoinPeriodExponent) |  | Exponent e that configures the rejoin period = 32 * 2^e + rand(0,32) seconds. | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.LinkADRAns">LinkADRAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
channel_mask_ack | [bool](#bool) |  |  | 
data_rate_index_ack | [bool](#bool) |  |  | 
tx_power_index_ack | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.LinkADRReq">LinkADRReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>
tx_power_index | [uint32](#uint32) |  |  | <p>`uint32.lte`: `15`</p>
channel_mask | [bool](#bool) | repeated |  | <p>`repeated.max_items`: `16`</p>
channel_mask_control | [uint32](#uint32) |  |  | <p>`uint32.lte`: `7`</p>
nb_trans | [uint32](#uint32) |  |  | <p>`uint32.lte`: `15`</p>

## <a name="ttn.lorawan.v3.MACCommand.LinkCheckAns">LinkCheckAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
margin | [uint32](#uint32) |  | Indicates the link margin in dB of the received LinkCheckReq, relative to the demodulation floor. | <p>`uint32.lte`: `254`</p>
gateway_count | [uint32](#uint32) |  |  | <p>`uint32.lte`: `255`</p>

## <a name="ttn.lorawan.v3.MACCommand.NewChannelAns">NewChannelAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency_ack | [bool](#bool) |  |  | 
data_rate_ack | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.NewChannelReq">NewChannelReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
channel_index | [uint32](#uint32) |  |  | <p>`uint32.lte`: `255`</p>
frequency | [uint64](#uint64) |  | Channel frequency (Hz). | <p>`uint64.gte`: `100000`</p>
min_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>
max_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.PingSlotChannelAns">PingSlotChannelAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency_ack | [bool](#bool) |  |  | 
data_rate_index_ack | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.PingSlotChannelReq">PingSlotChannelReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency | [uint64](#uint64) |  | Ping slot channel frequency (Hz). | <p>`uint64.gte`: `100000`</p>
data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.PingSlotInfoReq">PingSlotInfoReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
period | [PingSlotPeriod](#ttn.lorawan.v3.PingSlotPeriod) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.RejoinParamSetupAns">RejoinParamSetupAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
max_time_exponent_ack | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.RejoinParamSetupReq">RejoinParamSetupReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
max_count_exponent | [RejoinCountExponent](#ttn.lorawan.v3.RejoinCountExponent) |  | Exponent e that configures the rejoin counter = 2^(e+4) messages. | <p>`enum.defined_only`: `true`</p>
max_time_exponent | [RejoinTimeExponent](#ttn.lorawan.v3.RejoinTimeExponent) |  | Exponent e that configures the rejoin timer = 2^(e+10) seconds. | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.RekeyConf">RekeyConf</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.RekeyInd">RekeyInd</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.ResetConf">ResetConf</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.ResetInd">ResetInd</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.RxParamSetupAns">RxParamSetupAns</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
rx2_data_rate_index_ack | [bool](#bool) |  |  | 
rx1_data_rate_offset_ack | [bool](#bool) |  |  | 
rx2_frequency_ack | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACCommand.RxParamSetupReq">RxParamSetupReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
rx2_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  | <p>`enum.defined_only`: `true`</p>
rx1_data_rate_offset | [uint32](#uint32) |  |  | <p>`uint32.lte`: `7`</p>
rx2_frequency | [uint64](#uint64) |  | Rx2 frequency (Hz). | <p>`uint64.gte`: `100000`</p>

## <a name="ttn.lorawan.v3.MACCommand.RxTimingSetupReq">RxTimingSetupReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.MACCommand.TxParamSetupReq">TxParamSetupReq</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
max_eirp_index | [DeviceEIRP](#ttn.lorawan.v3.DeviceEIRP) |  | Indicates the maximum EIRP value in dBm, indexed by the following vector: [ 8 10 12 13 14 16 18 20 21 24 26 27 29 30 33 36 ] | <p>`enum.defined_only`: `true`</p>
uplink_dwell_time | [bool](#bool) |  |  | 
downlink_dwell_time | [bool](#bool) |  |  | 

## <a name="ttn.lorawan.v3.MACPayload">MACPayload</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
f_hdr | [FHDR](#ttn.lorawan.v3.FHDR) |  |  | <p>`message.required`: `true`</p>
f_port | [uint32](#uint32) |  |  | <p>`uint32.lte`: `255`</p>
frm_payload | [bytes](#bytes) |  |  | 
decoded_payload | [google.protobuf.Struct](#google.protobuf.Struct) |  |  | 

## <a name="ttn.lorawan.v3.MHDR">MHDR</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
m_type | [MType](#ttn.lorawan.v3.MType) |  |  | <p>`enum.defined_only`: `true`</p>
major | [Major](#ttn.lorawan.v3.Major) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.Message">Message</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
m_hdr | [MHDR](#ttn.lorawan.v3.MHDR) |  |  | <p>`message.required`: `true`</p>
mic | [bytes](#bytes) |  |  | 
mac_payload | [MACPayload](#ttn.lorawan.v3.MACPayload) |  |  | 
join_request_payload | [JoinRequestPayload](#ttn.lorawan.v3.JoinRequestPayload) |  |  | 
join_accept_payload | [JoinAcceptPayload](#ttn.lorawan.v3.JoinAcceptPayload) |  |  | 
rejoin_request_payload | [RejoinRequestPayload](#ttn.lorawan.v3.RejoinRequestPayload) |  |  | 

## <a name="ttn.lorawan.v3.RejoinRequestPayload">RejoinRequestPayload</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
rejoin_type | [RejoinType](#ttn.lorawan.v3.RejoinType) |  |  | <p>`enum.defined_only`: `true`</p>
net_id | [bytes](#bytes) |  |  | 
join_eui | [bytes](#bytes) |  |  | 
dev_eui | [bytes](#bytes) |  |  | 
rejoin_cnt | [uint32](#uint32) |  | Contains RJCount0 or RJCount1 depending on rejoin_type. | 

## <a name="ttn.lorawan.v3.TxRequest">TxRequest</a>

  TxRequest is a request for transmission.
If sent to a roaming partner, this request is used to generate the DLMetadata Object (see Backend Interfaces 1.0, Table 22).
If the gateway has a scheduler, this request is sent to the gateway, in the order of gateway_ids.
Otherwise, the Gateway Server attempts to schedule the request and creates the TxSettings.

Field | Type | Label | Description | Validation
---|---|---|---|---
class | [Class](#ttn.lorawan.v3.Class) |  |  | 
downlink_paths | [DownlinkPath](#ttn.lorawan.v3.DownlinkPath) | repeated | Downlink paths used to select a gateway for downlink. In class A, the downlink paths are required to only contain uplink tokens. In class B and C, the downlink paths may contain uplink tokens and fixed gateways antenna identifiers. | 
rx1_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  | Rx1 delay (Rx2 delay is Rx1 delay + 1 second). | <p>`enum.defined_only`: `true`</p>
rx1_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | LoRaWAN data rate index for Rx1. | <p>`enum.defined_only`: `true`</p>
rx1_frequency | [uint64](#uint64) |  | Frequency (Hz) for Rx1. | 
rx2_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | LoRaWAN data rate index for Rx2. | <p>`enum.defined_only`: `true`</p>
rx2_frequency | [uint64](#uint64) |  | Frequency (Hz) for Rx2. | 
priority | [TxSchedulePriority](#ttn.lorawan.v3.TxSchedulePriority) |  | Priority for scheduling. Requests with a higher priority are allocated more channel time than messages with a lower priority, in duty-cycle limited regions. A priority of HIGH or higher sets the HiPriorityFlag in the DLMetadata Object. | <p>`enum.defined_only`: `true`</p>
absolute_time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when the downlink message should be transmitted. This value is only valid for class C downlink; class A downlink uses uplink tokens and class B downlink is scheduled on ping slots. This requires the gateway to have GPS time sychronization. If the absolute time is not set, the first available time will be used that does not conflict or violate regional limitations. | 
advanced | [google.protobuf.Struct](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case | 

## <a name="ttn.lorawan.v3.TxSettings">TxSettings</a>

  TxSettings contains the settings for a transmission.
This message is used on both uplink and downlink.
On downlink, this is a scheduled transmission.

Field | Type | Label | Description | Validation
---|---|---|---|---
data_rate | [DataRate](#ttn.lorawan.v3.DataRate) |  | Data rate. | <p>`message.required`: `true`</p>
data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | LoRaWAN data rate index. | <p>`enum.defined_only`: `true`</p>
coding_rate | [string](#string) |  | LoRa coding rate. | 
frequency | [uint64](#uint64) |  | Frequency (Hz). | 
enable_crc | [bool](#bool) |  | Send a CRC in the packet; only on uplink; on downlink, CRC should not be enabled. | 
timestamp | [uint32](#uint32) |  | Timestamp of the gateway concentrator when the uplink message was received, or when the downlink message should be transmitted (microseconds). On downlink, set timestamp to 0 and time to null to use immediate scheduling. | 
time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time of the gateway when the uplink message was received, or when the downlink message should be transmitted. For downlink, this requires the gateway to have GPS time synchronization. | 
downlink | [TxSettings.Downlink](#ttn.lorawan.v3.TxSettings.Downlink) |  | Transmission settings for downlink. | 

## <a name="ttn.lorawan.v3.TxSettings.Downlink">Downlink</a>

  Transmission settings for downlink.

Field | Type | Label | Description | Validation
---|---|---|---|---
antenna_index | [uint32](#uint32) |  | Index of the antenna on which the uplink was received and/or downlink must be sent. | 
tx_power | [float](#float) |  | Transmission power (dBm). Only on downlink. | 
invert_polarization | [bool](#bool) |  | Invert LoRa polarization; false for LoRaWAN uplink, true for downlink. | 

## <a name="ttn.lorawan.v3.UplinkToken">UplinkToken</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers) |  |  | <p>`message.required`: `true`</p>
timestamp | [uint32](#uint32) |  |  | 
 

## <a name="ttn.lorawan.v3.ProcessDownlinkMessageRequest">ProcessDownlinkMessageRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
end_device_version_ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  |  | <p>`message.required`: `true`</p>
message | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  | <p>`message.required`: `true`</p>
parameter | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.ProcessUplinkMessageRequest">ProcessUplinkMessageRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
end_device_version_ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  |  | <p>`message.required`: `true`</p>
message | [ApplicationUplink](#ttn.lorawan.v3.ApplicationUplink) |  |  | <p>`message.required`: `true`</p>
parameter | [string](#string) |  |  | 
 

## <a name="ttn.lorawan.v3.ApplicationDownlink">ApplicationDownlink</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys used by this downlink. | <p>`bytes.max_len`: `2048`</p>
f_port | [uint32](#uint32) |  |  | <p>`uint32.lte`: `255`</p><p>`uint32.gte`: `1`</p>
f_cnt | [uint32](#uint32) |  |  | 
frm_payload | [bytes](#bytes) |  |  | 
decoded_payload | [google.protobuf.Struct](#google.protobuf.Struct) |  |  | 
confirmed | [bool](#bool) |  |  | 
class_b_c | [ApplicationDownlink.ClassBC](#ttn.lorawan.v3.ApplicationDownlink.ClassBC) |  | Optional gateway and timing information for class B and C. If set, this downlink message will only be transmitted as class B or C downlink. If not set, this downlink message may be transmitted in class A, B and C. | 
priority | [TxSchedulePriority](#ttn.lorawan.v3.TxSchedulePriority) |  | Priority for scheduling the downlink message. | <p>`enum.defined_only`: `true`</p>
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>

## <a name="ttn.lorawan.v3.ApplicationDownlink.ClassBC">ClassBC</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
gateways | [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers) | repeated | Possible gateway identifiers and antenna index to use for this downlink message. The Network Server selects one of these gateways for downlink, based on connectivity, signal quality, channel utilization and an available slot. If none of the gateways can be selected, the downlink message fails. If empty, a gateway and antenna is selected automatically from the gateways seen in recent uplinks. | 
absolute_time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Absolute time when the downlink message should be transmitted. This requires the gateway to have GPS time synchronization. If the time is in the past or if there is a scheduling conflict, the downlink message fails. If null, the time is selected based on slot availability. This is recommended in class B mode. | 

## <a name="ttn.lorawan.v3.ApplicationDownlinkFailed">ApplicationDownlinkFailed</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
downlink | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  | <p>`message.required`: `true`</p>
error | [ErrorDetails](#ttn.lorawan.v3.ErrorDetails) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.ApplicationDownlinks">ApplicationDownlinks</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  | 

## <a name="ttn.lorawan.v3.ApplicationInvalidatedDownlinks">ApplicationInvalidatedDownlinks</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  | <p>`repeated.min_items`: `1`</p>
last_f_cnt_down | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationJoinAccept">ApplicationJoinAccept</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys negotiated in this join. | <p>`bytes.max_len`: `2048`</p>
app_s_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | Encrypted Application Session Key (if Join Server sent it to Network Server). | 
invalidated_downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated | Downlink messages in the queue that got invalidated because of the session change. | 
pending_session | [bool](#bool) |  | Indicates whether the security context refers to the pending session, i.e. when this join-accept is an answer to a rejoin-request. | 

## <a name="ttn.lorawan.v3.ApplicationLocation">ApplicationLocation</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
service | [string](#string) |  |  | 
location | [Location](#ttn.lorawan.v3.Location) |  |  | <p>`message.required`: `true`</p>
attributes | [ApplicationLocation.AttributesEntry](#ttn.lorawan.v3.ApplicationLocation.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>

## <a name="ttn.lorawan.v3.ApplicationLocation.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationUp">ApplicationUp</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | <p>`message.required`: `true`</p>
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>
received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
uplink_message | [ApplicationUplink](#ttn.lorawan.v3.ApplicationUplink) |  |  | 
join_accept | [ApplicationJoinAccept](#ttn.lorawan.v3.ApplicationJoinAccept) |  |  | 
downlink_ack | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  | 
downlink_nack | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  | 
downlink_sent | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  | 
downlink_failed | [ApplicationDownlinkFailed](#ttn.lorawan.v3.ApplicationDownlinkFailed) |  |  | 
downlink_queued | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  | 
downlink_queue_invalidated | [ApplicationInvalidatedDownlinks](#ttn.lorawan.v3.ApplicationInvalidatedDownlinks) |  |  | 
location_solved | [ApplicationLocation](#ttn.lorawan.v3.ApplicationLocation) |  |  | 

## <a name="ttn.lorawan.v3.ApplicationUplink">ApplicationUplink</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys used by this uplink. | <p>`bytes.max_len`: `2048`</p>
f_port | [uint32](#uint32) |  |  | <p>`uint32.lte`: `255`</p><p>`uint32.gte`: `1`</p>
f_cnt | [uint32](#uint32) |  |  | 
frm_payload | [bytes](#bytes) |  |  | 
decoded_payload | [google.protobuf.Struct](#google.protobuf.Struct) |  |  | 
rx_metadata | [RxMetadata](#ttn.lorawan.v3.RxMetadata) | repeated |  | <p>`repeated.min_items`: `1`</p>
settings | [TxSettings](#ttn.lorawan.v3.TxSettings) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.DownlinkMessage">DownlinkMessage</a>

  Downlink message from the network to the end device

Field | Type | Label | Description | Validation
---|---|---|---|---
raw_payload | [bytes](#bytes) |  |  | 
payload | [Message](#ttn.lorawan.v3.Message) |  |  | 
end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | 
request | [TxRequest](#ttn.lorawan.v3.TxRequest) |  |  | 
scheduled | [TxSettings](#ttn.lorawan.v3.TxSettings) |  |  | 
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>

## <a name="ttn.lorawan.v3.DownlinkQueueRequest">DownlinkQueueRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  | 
downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  | 

## <a name="ttn.lorawan.v3.MessagePayloadFormatters">MessagePayloadFormatters</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
up_formatter | [PayloadFormatter](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter for uplink messages, must be set together with its parameter. | <p>`enum.defined_only`: `true`</p>
up_formatter_parameter | [string](#string) |  | Parameter for the up_formatter, must be set together. | 
down_formatter | [PayloadFormatter](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter for downlink messages, must be set together with its parameter. | <p>`enum.defined_only`: `true`</p>
down_formatter_parameter | [string](#string) |  | Parameter for the down_formatter, must be set together. | 

## <a name="ttn.lorawan.v3.TxAcknowledgment">TxAcknowledgment</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>
result | [TxAcknowledgment.Result](#ttn.lorawan.v3.TxAcknowledgment.Result) |  |  | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.UplinkMessage">UplinkMessage</a>

  Uplink message from the end device to the network

Field | Type | Label | Description | Validation
---|---|---|---|---
raw_payload | [bytes](#bytes) |  |  | 
payload | [Message](#ttn.lorawan.v3.Message) |  |  | 
settings | [TxSettings](#ttn.lorawan.v3.TxSettings) |  |  | <p>`message.required`: `true`</p>
rx_metadata | [RxMetadata](#ttn.lorawan.v3.RxMetadata) | repeated |  | <p>`repeated.min_items`: `1`</p>
received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Server time when a component received the message. The Gateway Server and Network Server set this value to their local server time of reception. | 
correlation_ids | [string](#string) | repeated |  | <p>`repeated.items.string.max_len`: `100`</p>
device_channel_index | [uint32](#uint32) |  | Index of the device channel that received the message. Set by Network Server. | <p>`uint32.lte`: `255`</p>
 

## <a name="ttn.lorawan.v3.Location">Location</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
latitude | [double](#double) |  | The North–South position (degrees; -90 to +90), where 0 is the equator, North pole is positive, South pole is negative. | <p>`double.lte`: `90`</p><p>`double.gte`: `-90`</p>
longitude | [double](#double) |  | The East-West position (degrees; -180 to +180), where 0 is the Prime Meridian (Greenwich), East is positive , West is negative. | <p>`double.lte`: `180`</p><p>`double.gte`: `-180`</p>
altitude | [int32](#int32) |  | The altitude (meters), where 0 is the mean sea level. | 
accuracy | [int32](#int32) |  | The accuracy of the location (meters). | 
source | [LocationSource](#ttn.lorawan.v3.LocationSource) |  | Source of the location information. | <p>`enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.RxMetadata">RxMetadata</a>

  Contains metadata for a received message. Each antenna that receives
a message corresponds to one RxMetadata.

Field | Type | Label | Description | Validation
---|---|---|---|---
gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  | <p>`message.required`: `true`</p>
antenna_index | [uint32](#uint32) |  |  | 
time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
timestamp | [uint32](#uint32) |  | Gateway concentrator timestamp when the Rx finished (microseconds). | 
fine_timestamp | [uint64](#uint64) |  | Gateway's internal fine timestamp when the Rx finished (nanoseconds). | 
encrypted_fine_timestamp | [bytes](#bytes) |  | Encrypted gateway's internal fine timestamp when the Rx finished (nanoseconds). | 
encrypted_fine_timestamp_key_id | [string](#string) |  |  | 
rssi | [float](#float) |  | Received signal strength indicator (dBm). This value equals `channel_rssi`. | 
signal_rssi | [google.protobuf.FloatValue](#google.protobuf.FloatValue) |  | Received signal strength indicator of the signal (dBm). | 
channel_rssi | [float](#float) |  | Received signal strength indicator of the channel (dBm). | 
rssi_standard_deviation | [float](#float) |  | Standard deviation of the RSSI during preamble. | 
snr | [float](#float) |  | Signal-to-noise ratio (dB). | 
frequency_offset | [int64](#int64) |  | Frequency offset (Hz). | 
location | [Location](#ttn.lorawan.v3.Location) |  | Antenna location; injected by the Gateway Server. | 
downlink_path_constraint | [DownlinkPathConstraint](#ttn.lorawan.v3.DownlinkPathConstraint) |  | Gateway downlink path constraint; injected by the Gateway Server. | <p>`enum.defined_only`: `true`</p>
uplink_token | [bytes](#bytes) |  | Uplink token to be included in the Tx request in class A downlink; injected by gateway, Gateway Server or fNS. | 
channel_index | [uint32](#uint32) |  | Index of the gateway channel that received the message. | <p>`uint32.lte`: `255`</p>
advanced | [google.protobuf.Struct](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case | 
 

## <a name="ttn.lorawan.v3.GenerateDevAddrResponse">GenerateDevAddrResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
dev_addr | [bytes](#bytes) |  |  | 
 

## <a name="ttn.lorawan.v3.ListOAuthAccessTokensRequest">ListOAuthAccessTokensRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest">ListOAuthClientAuthorizationsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.OAuthAccessToken">OAuthAccessToken</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
id | [string](#string) |  |  | 
access_token | [string](#string) |  |  | 
refresh_token | [string](#string) |  |  | 
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | 
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 

## <a name="ttn.lorawan.v3.OAuthAccessTokenIdentifiers">OAuthAccessTokenIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
id | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.OAuthAccessTokens">OAuthAccessTokens</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
tokens | [OAuthAccessToken](#ttn.lorawan.v3.OAuthAccessToken) | repeated |  | 

## <a name="ttn.lorawan.v3.OAuthAuthorizationCode">OAuthAuthorizationCode</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | 
code | [string](#string) |  |  | 
redirect_uri | [string](#string) |  |  | <p>`string.uri_ref`: `true`</p>
state | [string](#string) |  |  | 
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 

## <a name="ttn.lorawan.v3.OAuthClientAuthorization">OAuthClientAuthorization</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | 
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 

## <a name="ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers">OAuthClientAuthorizationIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.OAuthClientAuthorizations">OAuthClientAuthorizations</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
authorizations | [OAuthClientAuthorization](#ttn.lorawan.v3.OAuthClientAuthorization) | repeated |  | 
 
 

## <a name="ttn.lorawan.v3.CreateOrganizationAPIKeyRequest">CreateOrganizationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | <p>`repeated.items.enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.CreateOrganizationRequest">CreateOrganizationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization | [Organization](#ttn.lorawan.v3.Organization) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created application. NOTE: It is currently not possible to have organizations collaborating on other organizations. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.GetOrganizationAPIKeyRequest">GetOrganizationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
key_id | [string](#string) |  | Unique public identifier for the API key. | 

## <a name="ttn.lorawan.v3.GetOrganizationCollaboratorRequest">GetOrganizationCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | NOTE: It is currently not possible to have organizations collaborating on other organizations. | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.GetOrganizationRequest">GetOrganizationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.ListOrganizationAPIKeysRequest">ListOrganizationAPIKeysRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListOrganizationCollaboratorsRequest">ListOrganizationCollaboratorsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListOrganizationsRequest">ListOrganizationsRequest</a>

  By default we list all organizations the caller has rights on.
Set the user to instead list the organizations
where the user or organization is collaborator on.

Field | Type | Label | Description | Validation
---|---|---|---|---
collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | NOTE: It is currently not possible to have organizations collaborating on other organizations. | 
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.Organization">Organization</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
description | [string](#string) |  |  | <p>`string.max_len`: `2000`</p>
attributes | [Organization.AttributesEntry](#ttn.lorawan.v3.Organization.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  | 

## <a name="ttn.lorawan.v3.Organization.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.Organizations">Organizations</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organizations | [Organization](#ttn.lorawan.v3.Organization) | repeated |  | 

## <a name="ttn.lorawan.v3.SetOrganizationCollaboratorRequest">SetOrganizationCollaboratorRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest">UpdateOrganizationAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  | <p>`message.required`: `true`</p>
api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateOrganizationRequest">UpdateOrganizationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
organization | [Organization](#ttn.lorawan.v3.Organization) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 
 
 

## <a name="ttn.lorawan.v3.ConcentratorConfig">ConcentratorConfig</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
channels | [ConcentratorConfig.Channel](#ttn.lorawan.v3.ConcentratorConfig.Channel) | repeated |  | 
lora_standard_channel | [ConcentratorConfig.LoRaStandardChannel](#ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel) |  |  | 
fsk_channel | [ConcentratorConfig.FSKChannel](#ttn.lorawan.v3.ConcentratorConfig.FSKChannel) |  |  | 
lbt | [ConcentratorConfig.LBTConfiguration](#ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration) |  |  | 
ping_slot | [ConcentratorConfig.Channel](#ttn.lorawan.v3.ConcentratorConfig.Channel) |  |  | 
radios | [GatewayRadio](#ttn.lorawan.v3.GatewayRadio) | repeated |  | 
clock_source | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.ConcentratorConfig.Channel">Channel</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency | [uint64](#uint64) |  | Frequency (Hz). | 
radio | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.ConcentratorConfig.FSKChannel">FSKChannel</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency | [uint64](#uint64) |  | Frequency (Hz). | 
radio | [uint32](#uint32) |  |  | 

## <a name="ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration">LBTConfiguration</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
rssi_target | [float](#float) |  | Received signal strength (dBm). | 
rssi_offset | [float](#float) |  | Received signal strength offset (dBm). | 
scan_time | [google.protobuf.Duration](#google.protobuf.Duration) |  |  | 

## <a name="ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel">LoRaStandardChannel</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
frequency | [uint64](#uint64) |  | Frequency (Hz). | 
radio | [uint32](#uint32) |  |  | 
bandwidth | [uint32](#uint32) |  | Bandwidth (Hz). | 
spreading_factor | [uint32](#uint32) |  |  | 
 

## <a name="ttn.lorawan.v3.APIKey">APIKey</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
id | [string](#string) |  | Immutable and unique public identifier for the API key. Generated by the Access Server. | 
key | [string](#string) |  | Immutable and unique secret value of the API key. Generated by the Access Server. | 
name | [string](#string) |  | User-defined (friendly) name for the API key. | <p>`string.max_len`: `50`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated | Rights that are granted to this API key. | <p>`repeated.items.enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.APIKeys">APIKeys</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
api_keys | [APIKey](#ttn.lorawan.v3.APIKey) | repeated |  | 

## <a name="ttn.lorawan.v3.Collaborator">Collaborator</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | <p>`message.required`: `true`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | <p>`repeated.items.enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.Collaborators">Collaborators</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
collaborators | [Collaborator](#ttn.lorawan.v3.Collaborator) | repeated |  | 

## <a name="ttn.lorawan.v3.GetCollaboratorResponse">GetCollaboratorResponse</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  | 
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | 

## <a name="ttn.lorawan.v3.Rights">Rights</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | <p>`repeated.items.enum.defined_only`: `true`</p>
 

## <a name="ttn.lorawan.v3.SearchEndDevicesRequest">SearchEndDevicesRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  | <p>`message.required`: `true`</p>
id_contains | [string](#string) |  | Find end devices where the ID contains this substring. | 
name_contains | [string](#string) |  | Find end devices where the name contains this substring. | 
description_contains | [string](#string) |  | Find end devices where the description contains this substring. | 
attributes_contain | [SearchEndDevicesRequest.AttributesContainEntry](#ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry) | repeated | Find end devices where the given attributes contain these substrings. | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
dev_eui_contains | [string](#string) |  | Find end devices where the (hexadecimal) DevEUI contains this substring. | 
join_eui_contains | [string](#string) |  | Find end devices where the (hexadecimal) JoinEUI contains this substring. | 
dev_addr_contains | [string](#string) |  | Find end devices where the (hexadecimal) DevAddr contains this substring. | 
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry">AttributesContainEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.SearchEntitiesRequest">SearchEntitiesRequest</a>

  This message is used for finding entities in the EntityRegistrySearch service.

Field | Type | Label | Description | Validation
---|---|---|---|---
id_contains | [string](#string) |  | Find entities where the ID contains this substring. | 
name_contains | [string](#string) |  | Find entities where the name contains this substring. | 
description_contains | [string](#string) |  | Find entities where the description contains this substring. | 
attributes_contain | [SearchEntitiesRequest.AttributesContainEntry](#ttn.lorawan.v3.SearchEntitiesRequest.AttributesContainEntry) | repeated | Find entities where the given attributes contain these substrings. | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.SearchEntitiesRequest.AttributesContainEntry">AttributesContainEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 
 

## <a name="ttn.lorawan.v3.CreateTemporaryPasswordRequest">CreateTemporaryPasswordRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.CreateUserAPIKeyRequest">CreateUserAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
rights | [Right](#ttn.lorawan.v3.Right) | repeated |  | <p>`repeated.items.enum.defined_only`: `true`</p>

## <a name="ttn.lorawan.v3.CreateUserRequest">CreateUserRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user | [User](#ttn.lorawan.v3.User) |  |  | <p>`message.required`: `true`</p>
invitation_token | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.DeleteInvitationRequest">DeleteInvitationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
email | [string](#string) |  |  | <p>`string.email`: `true`</p>

## <a name="ttn.lorawan.v3.GetUserAPIKeyRequest">GetUserAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
key_id | [string](#string) |  | Unique public identifier for the API key. | 

## <a name="ttn.lorawan.v3.GetUserRequest">GetUserRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.Invitation">Invitation</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
email | [string](#string) |  |  | <p>`string.email`: `true`</p>
token | [string](#string) |  |  | 
expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
accepted_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
accepted_by | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | 

## <a name="ttn.lorawan.v3.Invitations">Invitations</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
invitations | [Invitation](#ttn.lorawan.v3.Invitation) | repeated |  | 

## <a name="ttn.lorawan.v3.ListInvitationsRequest">ListInvitationsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListUserAPIKeysRequest">ListUserAPIKeysRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.ListUserSessionsRequest">ListUserSessionsRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. | 
limit | [uint32](#uint32) |  | Limit the number of results per page. | <p>`uint32.lte`: `1000`</p>
page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. | 

## <a name="ttn.lorawan.v3.Picture">Picture</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
embedded | [Picture.Embedded](#ttn.lorawan.v3.Picture.Embedded) |  | Embedded picture, always maximum 128px in size. Omitted if there are external URLs available (in sizes). | 
sizes | [Picture.SizesEntry](#ttn.lorawan.v3.Picture.SizesEntry) | repeated | URLs of the picture for different sizes, if available on a CDN. | <p>`map.values.string.uri_ref`: `true`</p>

## <a name="ttn.lorawan.v3.Picture.Embedded">Embedded</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
mime_type | [string](#string) |  | MIME type of the picture. | 
data | [bytes](#bytes) |  | Picture data. A data URI can be constructed as follows: `data:<mime_type>;base64,<data>`. | 

## <a name="ttn.lorawan.v3.Picture.SizesEntry">SizesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [uint32](#uint32) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.SendInvitationRequest">SendInvitationRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
email | [string](#string) |  |  | <p>`string.email`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateUserAPIKeyRequest">UpdateUserAPIKeyRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  | <p>`message.required`: `true`</p>

## <a name="ttn.lorawan.v3.UpdateUserPasswordRequest">UpdateUserPasswordRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
new | [string](#string) |  |  | 
old | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.UpdateUserRequest">UpdateUserRequest</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user | [User](#ttn.lorawan.v3.User) |  |  | <p>`message.required`: `true`</p>
field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  | 

## <a name="ttn.lorawan.v3.User">User</a>

  User is the message that defines an user on the network.

Field | Type | Label | Description | Validation
---|---|---|---|---
ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
name | [string](#string) |  |  | <p>`string.max_len`: `50`</p>
description | [string](#string) |  |  | <p>`string.max_len`: `2000`</p>
attributes | [User.AttributesEntry](#ttn.lorawan.v3.User.AttributesEntry) | repeated |  | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p>
contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  | 
primary_email_address | [string](#string) |  | Primary email address that can be used for logging in. This address is not public, use contact_info for that. | <p>`string.email`: `true`</p>
primary_email_address_validated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
password | [string](#string) |  | Only used on create; never returned on API calls. | 
password_updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
require_password_update | [bool](#bool) |  |  | 
state | [State](#ttn.lorawan.v3.State) |  | The reviewing state of the user. This field can only be modified by admins. | <p>`enum.defined_only`: `true`</p>
admin | [bool](#bool) |  | This user is an admin. This field can only be modified by other admins. | 
temporary_password | [string](#string) |  | The temporary password can only be used to update a user's password; never returned on API calls. | 
temporary_password_created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
temporary_password_expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
profile_picture | [Picture](#ttn.lorawan.v3.Picture) |  |  | 

## <a name="ttn.lorawan.v3.User.AttributesEntry">AttributesEntry</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
key | [string](#string) |  |  | 
value | [string](#string) |  |  | 

## <a name="ttn.lorawan.v3.UserSession">UserSession</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
session_id | [string](#string) |  |  | <p>`string.max_len`: `64`</p>
created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 
expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  | 

## <a name="ttn.lorawan.v3.UserSessionIdentifiers">UserSessionIdentifiers</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  | <p>`message.required`: `true`</p>
session_id | [string](#string) |  |  | <p>`string.max_len`: `64`</p>

## <a name="ttn.lorawan.v3.UserSessions">UserSessions</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
sessions | [UserSession](#ttn.lorawan.v3.UserSession) | repeated |  | 

## <a name="ttn.lorawan.v3.Users">Users</a>

  

Field | Type | Label | Description | Validation
---|---|---|---|---
users | [User](#ttn.lorawan.v3.User) | repeated |  | 
 
