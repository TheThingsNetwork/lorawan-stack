# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [lorawan-stack/api/_api.proto](#lorawan-stack/api/_api.proto)
  
  
  
  

- [lorawan-stack/api/application.proto](#lorawan-stack/api/application.proto)
    - [Application](#ttn.lorawan.v3.Application)
    - [Application.AttributesEntry](#ttn.lorawan.v3.Application.AttributesEntry)
    - [Applications](#ttn.lorawan.v3.Applications)
    - [CreateApplicationAPIKeyRequest](#ttn.lorawan.v3.CreateApplicationAPIKeyRequest)
    - [CreateApplicationRequest](#ttn.lorawan.v3.CreateApplicationRequest)
    - [GetApplicationRequest](#ttn.lorawan.v3.GetApplicationRequest)
    - [ListApplicationAPIKeysRequest](#ttn.lorawan.v3.ListApplicationAPIKeysRequest)
    - [ListApplicationCollaboratorsRequest](#ttn.lorawan.v3.ListApplicationCollaboratorsRequest)
    - [ListApplicationsRequest](#ttn.lorawan.v3.ListApplicationsRequest)
    - [SetApplicationCollaboratorRequest](#ttn.lorawan.v3.SetApplicationCollaboratorRequest)
    - [UpdateApplicationAPIKeyRequest](#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest)
    - [UpdateApplicationRequest](#ttn.lorawan.v3.UpdateApplicationRequest)
  
  
  
  

- [lorawan-stack/api/application_services.proto](#lorawan-stack/api/application_services.proto)
  
  
  
    - [ApplicationAccess](#ttn.lorawan.v3.ApplicationAccess)
    - [ApplicationRegistry](#ttn.lorawan.v3.ApplicationRegistry)
  

- [lorawan-stack/api/applicationserver.proto](#lorawan-stack/api/applicationserver.proto)
    - [ApplicationLink](#ttn.lorawan.v3.ApplicationLink)
    - [ApplicationLinkStats](#ttn.lorawan.v3.ApplicationLinkStats)
    - [GetApplicationLinkRequest](#ttn.lorawan.v3.GetApplicationLinkRequest)
    - [SetApplicationLinkRequest](#ttn.lorawan.v3.SetApplicationLinkRequest)
  
  
  
    - [AppAs](#ttn.lorawan.v3.AppAs)
    - [As](#ttn.lorawan.v3.As)
    - [AsEndDeviceRegistry](#ttn.lorawan.v3.AsEndDeviceRegistry)
  

- [lorawan-stack/api/applicationserver_web.proto](#lorawan-stack/api/applicationserver_web.proto)
    - [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook)
    - [ApplicationWebhook.HeadersEntry](#ttn.lorawan.v3.ApplicationWebhook.HeadersEntry)
    - [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message)
    - [ApplicationWebhookFormats](#ttn.lorawan.v3.ApplicationWebhookFormats)
    - [ApplicationWebhookFormats.FormatsEntry](#ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry)
    - [ApplicationWebhookIdentifiers](#ttn.lorawan.v3.ApplicationWebhookIdentifiers)
    - [ApplicationWebhooks](#ttn.lorawan.v3.ApplicationWebhooks)
    - [GetApplicationWebhookRequest](#ttn.lorawan.v3.GetApplicationWebhookRequest)
    - [ListApplicationWebhooksRequest](#ttn.lorawan.v3.ListApplicationWebhooksRequest)
    - [SetApplicationWebhookRequest](#ttn.lorawan.v3.SetApplicationWebhookRequest)
  
  
  
    - [ApplicationWebhookRegistry](#ttn.lorawan.v3.ApplicationWebhookRegistry)
  

- [lorawan-stack/api/client.proto](#lorawan-stack/api/client.proto)
    - [Client](#ttn.lorawan.v3.Client)
    - [Client.AttributesEntry](#ttn.lorawan.v3.Client.AttributesEntry)
    - [Clients](#ttn.lorawan.v3.Clients)
    - [CreateClientRequest](#ttn.lorawan.v3.CreateClientRequest)
    - [GetClientRequest](#ttn.lorawan.v3.GetClientRequest)
    - [ListClientCollaboratorsRequest](#ttn.lorawan.v3.ListClientCollaboratorsRequest)
    - [ListClientsRequest](#ttn.lorawan.v3.ListClientsRequest)
    - [SetClientCollaboratorRequest](#ttn.lorawan.v3.SetClientCollaboratorRequest)
    - [UpdateClientRequest](#ttn.lorawan.v3.UpdateClientRequest)
  
    - [GrantType](#ttn.lorawan.v3.GrantType)
  
  
  

- [lorawan-stack/api/client_services.proto](#lorawan-stack/api/client_services.proto)
  
  
  
    - [ClientAccess](#ttn.lorawan.v3.ClientAccess)
    - [ClientRegistry](#ttn.lorawan.v3.ClientRegistry)
  

- [lorawan-stack/api/cluster.proto](#lorawan-stack/api/cluster.proto)
    - [PeerInfo](#ttn.lorawan.v3.PeerInfo)
    - [PeerInfo.TagsEntry](#ttn.lorawan.v3.PeerInfo.TagsEntry)
  
    - [PeerInfo.Role](#ttn.lorawan.v3.PeerInfo.Role)
  
  
  

- [lorawan-stack/api/configuration_services.proto](#lorawan-stack/api/configuration_services.proto)
    - [FrequencyPlanDescription](#ttn.lorawan.v3.FrequencyPlanDescription)
    - [ListFrequencyPlansRequest](#ttn.lorawan.v3.ListFrequencyPlansRequest)
    - [ListFrequencyPlansResponse](#ttn.lorawan.v3.ListFrequencyPlansResponse)
  
  
  
    - [Configuration](#ttn.lorawan.v3.Configuration)
  

- [lorawan-stack/api/contact_info.proto](#lorawan-stack/api/contact_info.proto)
    - [ContactInfo](#ttn.lorawan.v3.ContactInfo)
    - [ContactInfoValidation](#ttn.lorawan.v3.ContactInfoValidation)
  
    - [ContactMethod](#ttn.lorawan.v3.ContactMethod)
    - [ContactType](#ttn.lorawan.v3.ContactType)
  
  
    - [ContactInfoRegistry](#ttn.lorawan.v3.ContactInfoRegistry)
  

- [lorawan-stack/api/end_device.proto](#lorawan-stack/api/end_device.proto)
    - [CreateEndDeviceRequest](#ttn.lorawan.v3.CreateEndDeviceRequest)
    - [EndDevice](#ttn.lorawan.v3.EndDevice)
    - [EndDevice.AttributesEntry](#ttn.lorawan.v3.EndDevice.AttributesEntry)
    - [EndDevice.LocationsEntry](#ttn.lorawan.v3.EndDevice.LocationsEntry)
    - [EndDeviceBrand](#ttn.lorawan.v3.EndDeviceBrand)
    - [EndDeviceModel](#ttn.lorawan.v3.EndDeviceModel)
    - [EndDeviceVersion](#ttn.lorawan.v3.EndDeviceVersion)
    - [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers)
    - [EndDevices](#ttn.lorawan.v3.EndDevices)
    - [GetEndDeviceRequest](#ttn.lorawan.v3.GetEndDeviceRequest)
    - [ListEndDevicesRequest](#ttn.lorawan.v3.ListEndDevicesRequest)
    - [MACParameters](#ttn.lorawan.v3.MACParameters)
    - [MACParameters.Channel](#ttn.lorawan.v3.MACParameters.Channel)
    - [MACSettings](#ttn.lorawan.v3.MACSettings)
    - [MACSettings.AggregatedDutyCycleValue](#ttn.lorawan.v3.MACSettings.AggregatedDutyCycleValue)
    - [MACSettings.DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue)
    - [MACSettings.PingSlotPeriodValue](#ttn.lorawan.v3.MACSettings.PingSlotPeriodValue)
    - [MACSettings.RxDelayValue](#ttn.lorawan.v3.MACSettings.RxDelayValue)
    - [MACState](#ttn.lorawan.v3.MACState)
    - [MACState.JoinAccept](#ttn.lorawan.v3.MACState.JoinAccept)
    - [Session](#ttn.lorawan.v3.Session)
    - [SetEndDeviceRequest](#ttn.lorawan.v3.SetEndDeviceRequest)
    - [UpdateEndDeviceRequest](#ttn.lorawan.v3.UpdateEndDeviceRequest)
  
    - [PowerState](#ttn.lorawan.v3.PowerState)
  
  
  

- [lorawan-stack/api/end_device_services.proto](#lorawan-stack/api/end_device_services.proto)
  
  
  
    - [EndDeviceRegistry](#ttn.lorawan.v3.EndDeviceRegistry)
  

- [lorawan-stack/api/enums.proto](#lorawan-stack/api/enums.proto)
  
    - [DownlinkPathConstraint](#ttn.lorawan.v3.DownlinkPathConstraint)
    - [State](#ttn.lorawan.v3.State)
  
  
  

- [lorawan-stack/api/error.proto](#lorawan-stack/api/error.proto)
    - [ErrorDetails](#ttn.lorawan.v3.ErrorDetails)
  
  
  
  

- [lorawan-stack/api/events.proto](#lorawan-stack/api/events.proto)
    - [Event](#ttn.lorawan.v3.Event)
    - [Event.ContextEntry](#ttn.lorawan.v3.Event.ContextEntry)
    - [StreamEventsRequest](#ttn.lorawan.v3.StreamEventsRequest)
  
  
  
    - [Events](#ttn.lorawan.v3.Events)
  

- [lorawan-stack/api/gateway.proto](#lorawan-stack/api/gateway.proto)
    - [CreateGatewayAPIKeyRequest](#ttn.lorawan.v3.CreateGatewayAPIKeyRequest)
    - [CreateGatewayRequest](#ttn.lorawan.v3.CreateGatewayRequest)
    - [Gateway](#ttn.lorawan.v3.Gateway)
    - [Gateway.AttributesEntry](#ttn.lorawan.v3.Gateway.AttributesEntry)
    - [GatewayAntenna](#ttn.lorawan.v3.GatewayAntenna)
    - [GatewayAntenna.AttributesEntry](#ttn.lorawan.v3.GatewayAntenna.AttributesEntry)
    - [GatewayBrand](#ttn.lorawan.v3.GatewayBrand)
    - [GatewayConnectionStats](#ttn.lorawan.v3.GatewayConnectionStats)
    - [GatewayModel](#ttn.lorawan.v3.GatewayModel)
    - [GatewayRadio](#ttn.lorawan.v3.GatewayRadio)
    - [GatewayRadio.TxConfiguration](#ttn.lorawan.v3.GatewayRadio.TxConfiguration)
    - [GatewayStatus](#ttn.lorawan.v3.GatewayStatus)
    - [GatewayStatus.MetricsEntry](#ttn.lorawan.v3.GatewayStatus.MetricsEntry)
    - [GatewayStatus.VersionsEntry](#ttn.lorawan.v3.GatewayStatus.VersionsEntry)
    - [GatewayVersion](#ttn.lorawan.v3.GatewayVersion)
    - [GatewayVersionIdentifiers](#ttn.lorawan.v3.GatewayVersionIdentifiers)
    - [Gateways](#ttn.lorawan.v3.Gateways)
    - [GetGatewayIdentifiersForEUIRequest](#ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest)
    - [GetGatewayRequest](#ttn.lorawan.v3.GetGatewayRequest)
    - [ListGatewayAPIKeysRequest](#ttn.lorawan.v3.ListGatewayAPIKeysRequest)
    - [ListGatewayCollaboratorsRequest](#ttn.lorawan.v3.ListGatewayCollaboratorsRequest)
    - [ListGatewaysRequest](#ttn.lorawan.v3.ListGatewaysRequest)
    - [SetGatewayCollaboratorRequest](#ttn.lorawan.v3.SetGatewayCollaboratorRequest)
    - [UpdateGatewayAPIKeyRequest](#ttn.lorawan.v3.UpdateGatewayAPIKeyRequest)
    - [UpdateGatewayRequest](#ttn.lorawan.v3.UpdateGatewayRequest)
  
  
  
  

- [lorawan-stack/api/gateway_services.proto](#lorawan-stack/api/gateway_services.proto)
    - [PullGatewayConfigurationRequest](#ttn.lorawan.v3.PullGatewayConfigurationRequest)
  
  
  
    - [GatewayAccess](#ttn.lorawan.v3.GatewayAccess)
    - [GatewayConfigurator](#ttn.lorawan.v3.GatewayConfigurator)
    - [GatewayRegistry](#ttn.lorawan.v3.GatewayRegistry)
  

- [lorawan-stack/api/gatewayserver.proto](#lorawan-stack/api/gatewayserver.proto)
    - [GatewayDown](#ttn.lorawan.v3.GatewayDown)
    - [GatewayUp](#ttn.lorawan.v3.GatewayUp)
    - [ScheduleDownlinkResponse](#ttn.lorawan.v3.ScheduleDownlinkResponse)
  
  
  
    - [Gs](#ttn.lorawan.v3.Gs)
    - [GtwGs](#ttn.lorawan.v3.GtwGs)
    - [NsGs](#ttn.lorawan.v3.NsGs)
  

- [lorawan-stack/api/identifiers.proto](#lorawan-stack/api/identifiers.proto)
    - [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers)
    - [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers)
    - [CombinedIdentifiers](#ttn.lorawan.v3.CombinedIdentifiers)
    - [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers)
    - [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers)
    - [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers)
    - [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers)
    - [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers)
    - [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers)
  
  
  
  

- [lorawan-stack/api/identityserver.proto](#lorawan-stack/api/identityserver.proto)
    - [AuthInfoResponse](#ttn.lorawan.v3.AuthInfoResponse)
    - [AuthInfoResponse.APIKeyAccess](#ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess)
  
  
  
    - [EntityAccess](#ttn.lorawan.v3.EntityAccess)
  

- [lorawan-stack/api/join.proto](#lorawan-stack/api/join.proto)
    - [JoinRequest](#ttn.lorawan.v3.JoinRequest)
    - [JoinResponse](#ttn.lorawan.v3.JoinResponse)
  
  
  
  

- [lorawan-stack/api/joinserver.proto](#lorawan-stack/api/joinserver.proto)
    - [AppSKeyResponse](#ttn.lorawan.v3.AppSKeyResponse)
    - [CryptoServicePayloadRequest](#ttn.lorawan.v3.CryptoServicePayloadRequest)
    - [CryptoServicePayloadResponse](#ttn.lorawan.v3.CryptoServicePayloadResponse)
    - [DeriveSessionKeysRequest](#ttn.lorawan.v3.DeriveSessionKeysRequest)
    - [GetRootKeysRequest](#ttn.lorawan.v3.GetRootKeysRequest)
    - [JoinAcceptMICRequest](#ttn.lorawan.v3.JoinAcceptMICRequest)
    - [NwkSKeysResponse](#ttn.lorawan.v3.NwkSKeysResponse)
    - [ProvisionEndDevicesRequest](#ttn.lorawan.v3.ProvisionEndDevicesRequest)
    - [ProvisionEndDevicesRequest.IdentifiersFromData](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData)
    - [ProvisionEndDevicesRequest.IdentifiersList](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList)
    - [ProvisionEndDevicesRequest.IdentifiersRange](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange)
    - [SessionKeyRequest](#ttn.lorawan.v3.SessionKeyRequest)
  
  
  
    - [ApplicationCryptoService](#ttn.lorawan.v3.ApplicationCryptoService)
    - [AsJs](#ttn.lorawan.v3.AsJs)
    - [JsEndDeviceRegistry](#ttn.lorawan.v3.JsEndDeviceRegistry)
    - [NetworkCryptoService](#ttn.lorawan.v3.NetworkCryptoService)
    - [NsJs](#ttn.lorawan.v3.NsJs)
  

- [lorawan-stack/api/keys.proto](#lorawan-stack/api/keys.proto)
    - [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope)
    - [RootKeys](#ttn.lorawan.v3.RootKeys)
    - [SessionKeys](#ttn.lorawan.v3.SessionKeys)
  
  
  
  

- [lorawan-stack/api/lorawan.proto](#lorawan-stack/api/lorawan.proto)
    - [CFList](#ttn.lorawan.v3.CFList)
    - [DLSettings](#ttn.lorawan.v3.DLSettings)
    - [DataRate](#ttn.lorawan.v3.DataRate)
    - [DownlinkPath](#ttn.lorawan.v3.DownlinkPath)
    - [FCtrl](#ttn.lorawan.v3.FCtrl)
    - [FHDR](#ttn.lorawan.v3.FHDR)
    - [FSKDataRate](#ttn.lorawan.v3.FSKDataRate)
    - [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers)
    - [JoinAcceptPayload](#ttn.lorawan.v3.JoinAcceptPayload)
    - [JoinRequestPayload](#ttn.lorawan.v3.JoinRequestPayload)
    - [LoRaDataRate](#ttn.lorawan.v3.LoRaDataRate)
    - [MACCommand](#ttn.lorawan.v3.MACCommand)
    - [MACCommand.ADRParamSetupReq](#ttn.lorawan.v3.MACCommand.ADRParamSetupReq)
    - [MACCommand.BeaconFreqAns](#ttn.lorawan.v3.MACCommand.BeaconFreqAns)
    - [MACCommand.BeaconFreqReq](#ttn.lorawan.v3.MACCommand.BeaconFreqReq)
    - [MACCommand.BeaconTimingAns](#ttn.lorawan.v3.MACCommand.BeaconTimingAns)
    - [MACCommand.DLChannelAns](#ttn.lorawan.v3.MACCommand.DLChannelAns)
    - [MACCommand.DLChannelReq](#ttn.lorawan.v3.MACCommand.DLChannelReq)
    - [MACCommand.DevStatusAns](#ttn.lorawan.v3.MACCommand.DevStatusAns)
    - [MACCommand.DeviceModeConf](#ttn.lorawan.v3.MACCommand.DeviceModeConf)
    - [MACCommand.DeviceModeInd](#ttn.lorawan.v3.MACCommand.DeviceModeInd)
    - [MACCommand.DeviceTimeAns](#ttn.lorawan.v3.MACCommand.DeviceTimeAns)
    - [MACCommand.DutyCycleReq](#ttn.lorawan.v3.MACCommand.DutyCycleReq)
    - [MACCommand.ForceRejoinReq](#ttn.lorawan.v3.MACCommand.ForceRejoinReq)
    - [MACCommand.LinkADRAns](#ttn.lorawan.v3.MACCommand.LinkADRAns)
    - [MACCommand.LinkADRReq](#ttn.lorawan.v3.MACCommand.LinkADRReq)
    - [MACCommand.LinkCheckAns](#ttn.lorawan.v3.MACCommand.LinkCheckAns)
    - [MACCommand.NewChannelAns](#ttn.lorawan.v3.MACCommand.NewChannelAns)
    - [MACCommand.NewChannelReq](#ttn.lorawan.v3.MACCommand.NewChannelReq)
    - [MACCommand.PingSlotChannelAns](#ttn.lorawan.v3.MACCommand.PingSlotChannelAns)
    - [MACCommand.PingSlotChannelReq](#ttn.lorawan.v3.MACCommand.PingSlotChannelReq)
    - [MACCommand.PingSlotInfoReq](#ttn.lorawan.v3.MACCommand.PingSlotInfoReq)
    - [MACCommand.RejoinParamSetupAns](#ttn.lorawan.v3.MACCommand.RejoinParamSetupAns)
    - [MACCommand.RejoinParamSetupReq](#ttn.lorawan.v3.MACCommand.RejoinParamSetupReq)
    - [MACCommand.RekeyConf](#ttn.lorawan.v3.MACCommand.RekeyConf)
    - [MACCommand.RekeyInd](#ttn.lorawan.v3.MACCommand.RekeyInd)
    - [MACCommand.ResetConf](#ttn.lorawan.v3.MACCommand.ResetConf)
    - [MACCommand.ResetInd](#ttn.lorawan.v3.MACCommand.ResetInd)
    - [MACCommand.RxParamSetupAns](#ttn.lorawan.v3.MACCommand.RxParamSetupAns)
    - [MACCommand.RxParamSetupReq](#ttn.lorawan.v3.MACCommand.RxParamSetupReq)
    - [MACCommand.RxTimingSetupReq](#ttn.lorawan.v3.MACCommand.RxTimingSetupReq)
    - [MACCommand.TxParamSetupReq](#ttn.lorawan.v3.MACCommand.TxParamSetupReq)
    - [MACPayload](#ttn.lorawan.v3.MACPayload)
    - [MHDR](#ttn.lorawan.v3.MHDR)
    - [Message](#ttn.lorawan.v3.Message)
    - [RejoinRequestPayload](#ttn.lorawan.v3.RejoinRequestPayload)
    - [TxRequest](#ttn.lorawan.v3.TxRequest)
    - [TxSettings](#ttn.lorawan.v3.TxSettings)
    - [TxSettings.Downlink](#ttn.lorawan.v3.TxSettings.Downlink)
    - [UplinkToken](#ttn.lorawan.v3.UplinkToken)
  
    - [ADRAckDelayExponent](#ttn.lorawan.v3.ADRAckDelayExponent)
    - [ADRAckLimitExponent](#ttn.lorawan.v3.ADRAckLimitExponent)
    - [AggregatedDutyCycle](#ttn.lorawan.v3.AggregatedDutyCycle)
    - [CFListType](#ttn.lorawan.v3.CFListType)
    - [Class](#ttn.lorawan.v3.Class)
    - [DataRateIndex](#ttn.lorawan.v3.DataRateIndex)
    - [DeviceEIRP](#ttn.lorawan.v3.DeviceEIRP)
    - [MACCommandIdentifier](#ttn.lorawan.v3.MACCommandIdentifier)
    - [MACVersion](#ttn.lorawan.v3.MACVersion)
    - [MType](#ttn.lorawan.v3.MType)
    - [Major](#ttn.lorawan.v3.Major)
    - [Minor](#ttn.lorawan.v3.Minor)
    - [PHYVersion](#ttn.lorawan.v3.PHYVersion)
    - [PingSlotPeriod](#ttn.lorawan.v3.PingSlotPeriod)
    - [RejoinCountExponent](#ttn.lorawan.v3.RejoinCountExponent)
    - [RejoinPeriodExponent](#ttn.lorawan.v3.RejoinPeriodExponent)
    - [RejoinTimeExponent](#ttn.lorawan.v3.RejoinTimeExponent)
    - [RejoinType](#ttn.lorawan.v3.RejoinType)
    - [RxDelay](#ttn.lorawan.v3.RxDelay)
    - [TxSchedulePriority](#ttn.lorawan.v3.TxSchedulePriority)
  
  
  

- [lorawan-stack/api/message_services.proto](#lorawan-stack/api/message_services.proto)
    - [ProcessDownlinkMessageRequest](#ttn.lorawan.v3.ProcessDownlinkMessageRequest)
    - [ProcessUplinkMessageRequest](#ttn.lorawan.v3.ProcessUplinkMessageRequest)
  
  
  
    - [DownlinkMessageProcessor](#ttn.lorawan.v3.DownlinkMessageProcessor)
    - [UplinkMessageProcessor](#ttn.lorawan.v3.UplinkMessageProcessor)
  

- [lorawan-stack/api/messages.proto](#lorawan-stack/api/messages.proto)
    - [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink)
    - [ApplicationDownlink.ClassBC](#ttn.lorawan.v3.ApplicationDownlink.ClassBC)
    - [ApplicationDownlinkFailed](#ttn.lorawan.v3.ApplicationDownlinkFailed)
    - [ApplicationDownlinks](#ttn.lorawan.v3.ApplicationDownlinks)
    - [ApplicationInvalidatedDownlinks](#ttn.lorawan.v3.ApplicationInvalidatedDownlinks)
    - [ApplicationJoinAccept](#ttn.lorawan.v3.ApplicationJoinAccept)
    - [ApplicationLocation](#ttn.lorawan.v3.ApplicationLocation)
    - [ApplicationLocation.AttributesEntry](#ttn.lorawan.v3.ApplicationLocation.AttributesEntry)
    - [ApplicationUp](#ttn.lorawan.v3.ApplicationUp)
    - [ApplicationUplink](#ttn.lorawan.v3.ApplicationUplink)
    - [DownlinkMessage](#ttn.lorawan.v3.DownlinkMessage)
    - [DownlinkQueueRequest](#ttn.lorawan.v3.DownlinkQueueRequest)
    - [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters)
    - [TxAcknowledgment](#ttn.lorawan.v3.TxAcknowledgment)
    - [UplinkMessage](#ttn.lorawan.v3.UplinkMessage)
  
    - [PayloadFormatter](#ttn.lorawan.v3.PayloadFormatter)
    - [TxAcknowledgment.Result](#ttn.lorawan.v3.TxAcknowledgment.Result)
  
  
  

- [lorawan-stack/api/metadata.proto](#lorawan-stack/api/metadata.proto)
    - [Location](#ttn.lorawan.v3.Location)
    - [RxMetadata](#ttn.lorawan.v3.RxMetadata)
  
    - [LocationSource](#ttn.lorawan.v3.LocationSource)
  
  
  

- [lorawan-stack/api/networkserver.proto](#lorawan-stack/api/networkserver.proto)
  
  
  
    - [AsNs](#ttn.lorawan.v3.AsNs)
    - [GsNs](#ttn.lorawan.v3.GsNs)
    - [NsEndDeviceRegistry](#ttn.lorawan.v3.NsEndDeviceRegistry)
  

- [lorawan-stack/api/oauth.proto](#lorawan-stack/api/oauth.proto)
    - [ListOAuthAccessTokensRequest](#ttn.lorawan.v3.ListOAuthAccessTokensRequest)
    - [ListOAuthClientAuthorizationsRequest](#ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest)
    - [OAuthAccessToken](#ttn.lorawan.v3.OAuthAccessToken)
    - [OAuthAccessTokenIdentifiers](#ttn.lorawan.v3.OAuthAccessTokenIdentifiers)
    - [OAuthAccessTokens](#ttn.lorawan.v3.OAuthAccessTokens)
    - [OAuthAuthorizationCode](#ttn.lorawan.v3.OAuthAuthorizationCode)
    - [OAuthClientAuthorization](#ttn.lorawan.v3.OAuthClientAuthorization)
    - [OAuthClientAuthorizationIdentifiers](#ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers)
    - [OAuthClientAuthorizations](#ttn.lorawan.v3.OAuthClientAuthorizations)
  
  
  
  

- [lorawan-stack/api/oauth_services.proto](#lorawan-stack/api/oauth_services.proto)
  
  
  
    - [OAuthAuthorizationRegistry](#ttn.lorawan.v3.OAuthAuthorizationRegistry)
  

- [lorawan-stack/api/organization.proto](#lorawan-stack/api/organization.proto)
    - [CreateOrganizationAPIKeyRequest](#ttn.lorawan.v3.CreateOrganizationAPIKeyRequest)
    - [CreateOrganizationRequest](#ttn.lorawan.v3.CreateOrganizationRequest)
    - [GetOrganizationRequest](#ttn.lorawan.v3.GetOrganizationRequest)
    - [ListOrganizationAPIKeysRequest](#ttn.lorawan.v3.ListOrganizationAPIKeysRequest)
    - [ListOrganizationCollaboratorsRequest](#ttn.lorawan.v3.ListOrganizationCollaboratorsRequest)
    - [ListOrganizationsRequest](#ttn.lorawan.v3.ListOrganizationsRequest)
    - [Organization](#ttn.lorawan.v3.Organization)
    - [Organization.AttributesEntry](#ttn.lorawan.v3.Organization.AttributesEntry)
    - [Organizations](#ttn.lorawan.v3.Organizations)
    - [SetOrganizationCollaboratorRequest](#ttn.lorawan.v3.SetOrganizationCollaboratorRequest)
    - [UpdateOrganizationAPIKeyRequest](#ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest)
    - [UpdateOrganizationRequest](#ttn.lorawan.v3.UpdateOrganizationRequest)
  
  
  
  

- [lorawan-stack/api/organization_services.proto](#lorawan-stack/api/organization_services.proto)
  
  
  
    - [OrganizationAccess](#ttn.lorawan.v3.OrganizationAccess)
    - [OrganizationRegistry](#ttn.lorawan.v3.OrganizationRegistry)
  

- [lorawan-stack/api/regional.proto](#lorawan-stack/api/regional.proto)
    - [ConcentratorConfig](#ttn.lorawan.v3.ConcentratorConfig)
    - [ConcentratorConfig.Channel](#ttn.lorawan.v3.ConcentratorConfig.Channel)
    - [ConcentratorConfig.FSKChannel](#ttn.lorawan.v3.ConcentratorConfig.FSKChannel)
    - [ConcentratorConfig.LBTConfiguration](#ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration)
    - [ConcentratorConfig.LoRaStandardChannel](#ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel)
  
  
  
  

- [lorawan-stack/api/rights.proto](#lorawan-stack/api/rights.proto)
    - [APIKey](#ttn.lorawan.v3.APIKey)
    - [APIKeys](#ttn.lorawan.v3.APIKeys)
    - [Collaborator](#ttn.lorawan.v3.Collaborator)
    - [Collaborators](#ttn.lorawan.v3.Collaborators)
    - [Rights](#ttn.lorawan.v3.Rights)
  
    - [Right](#ttn.lorawan.v3.Right)
  
  
  

- [lorawan-stack/api/search_services.proto](#lorawan-stack/api/search_services.proto)
    - [SearchEndDevicesRequest](#ttn.lorawan.v3.SearchEndDevicesRequest)
    - [SearchEndDevicesRequest.AttributesContainEntry](#ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry)
    - [SearchEntitiesRequest](#ttn.lorawan.v3.SearchEntitiesRequest)
    - [SearchEntitiesRequest.AttributesContainEntry](#ttn.lorawan.v3.SearchEntitiesRequest.AttributesContainEntry)
  
  
  
    - [EndDeviceRegistrySearch](#ttn.lorawan.v3.EndDeviceRegistrySearch)
    - [EntityRegistrySearch](#ttn.lorawan.v3.EntityRegistrySearch)
  

- [lorawan-stack/api/user.proto](#lorawan-stack/api/user.proto)
    - [CreateTemporaryPasswordRequest](#ttn.lorawan.v3.CreateTemporaryPasswordRequest)
    - [CreateUserAPIKeyRequest](#ttn.lorawan.v3.CreateUserAPIKeyRequest)
    - [CreateUserRequest](#ttn.lorawan.v3.CreateUserRequest)
    - [DeleteInvitationRequest](#ttn.lorawan.v3.DeleteInvitationRequest)
    - [GetUserRequest](#ttn.lorawan.v3.GetUserRequest)
    - [Invitation](#ttn.lorawan.v3.Invitation)
    - [Invitations](#ttn.lorawan.v3.Invitations)
    - [ListInvitationsRequest](#ttn.lorawan.v3.ListInvitationsRequest)
    - [ListUserAPIKeysRequest](#ttn.lorawan.v3.ListUserAPIKeysRequest)
    - [ListUserSessionsRequest](#ttn.lorawan.v3.ListUserSessionsRequest)
    - [Picture](#ttn.lorawan.v3.Picture)
    - [Picture.Embedded](#ttn.lorawan.v3.Picture.Embedded)
    - [Picture.SizesEntry](#ttn.lorawan.v3.Picture.SizesEntry)
    - [SendInvitationRequest](#ttn.lorawan.v3.SendInvitationRequest)
    - [UpdateUserAPIKeyRequest](#ttn.lorawan.v3.UpdateUserAPIKeyRequest)
    - [UpdateUserPasswordRequest](#ttn.lorawan.v3.UpdateUserPasswordRequest)
    - [UpdateUserRequest](#ttn.lorawan.v3.UpdateUserRequest)
    - [User](#ttn.lorawan.v3.User)
    - [User.AttributesEntry](#ttn.lorawan.v3.User.AttributesEntry)
    - [UserSession](#ttn.lorawan.v3.UserSession)
    - [UserSessionIdentifiers](#ttn.lorawan.v3.UserSessionIdentifiers)
    - [UserSessions](#ttn.lorawan.v3.UserSessions)
    - [Users](#ttn.lorawan.v3.Users)
  
  
  
  

- [lorawan-stack/api/user_services.proto](#lorawan-stack/api/user_services.proto)
  
  
  
    - [UserAccess](#ttn.lorawan.v3.UserAccess)
    - [UserInvitationRegistry](#ttn.lorawan.v3.UserInvitationRegistry)
    - [UserRegistry](#ttn.lorawan.v3.UserRegistry)
    - [UserSessionRegistry](#ttn.lorawan.v3.UserSessionRegistry)
  

- [Scalar Value Types](#scalar-value-types)



<a name="lorawan-stack/api/_api.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/_api.proto


 

 

 

 



<a name="lorawan-stack/api/application.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/application.proto



<a name="ttn.lorawan.v3.Application"></a>

### Application
Application is the message that defines an Application in the network.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| attributes | [Application.AttributesEntry](#ttn.lorawan.v3.Application.AttributesEntry) | repeated |  |
| contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  |






<a name="ttn.lorawan.v3.Application.AttributesEntry"></a>

### Application.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.Applications"></a>

### Applications



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| applications | [Application](#ttn.lorawan.v3.Application) | repeated |  |






<a name="ttn.lorawan.v3.CreateApplicationAPIKeyRequest"></a>

### CreateApplicationAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| name | [string](#string) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |






<a name="ttn.lorawan.v3.CreateApplicationRequest"></a>

### CreateApplicationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application | [Application](#ttn.lorawan.v3.Application) |  |  |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created application. |






<a name="ttn.lorawan.v3.GetApplicationRequest"></a>

### GetApplicationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.ListApplicationAPIKeysRequest"></a>

### ListApplicationAPIKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListApplicationCollaboratorsRequest"></a>

### ListApplicationCollaboratorsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListApplicationsRequest"></a>

### ListApplicationsRequest
By default we list all applications the caller has rights on.
Set the user or the organization (not both) to instead list the applications
where the user or organization is collaborator on.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.SetApplicationCollaboratorRequest"></a>

### SetApplicationCollaboratorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  |






<a name="ttn.lorawan.v3.UpdateApplicationAPIKeyRequest"></a>

### UpdateApplicationAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  |






<a name="ttn.lorawan.v3.UpdateApplicationRequest"></a>

### UpdateApplicationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application | [Application](#ttn.lorawan.v3.Application) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 

 

 

 



<a name="lorawan-stack/api/application_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/application_services.proto


 

 

 


<a name="ttn.lorawan.v3.ApplicationAccess"></a>

### ApplicationAccess


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListRights | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) | [Rights](#ttn.lorawan.v3.Rights) |  |
| CreateAPIKey | [CreateApplicationAPIKeyRequest](#ttn.lorawan.v3.CreateApplicationAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) |  |
| ListAPIKeys | [ListApplicationAPIKeysRequest](#ttn.lorawan.v3.ListApplicationAPIKeysRequest) | [APIKeys](#ttn.lorawan.v3.APIKeys) |  |
| UpdateAPIKey | [UpdateApplicationAPIKeyRequest](#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) | Update the rights of an existing application API key. To generate an API key, the CreateAPIKey should be used. To delete an API key, update it with zero rights. |
| SetCollaborator | [SetApplicationCollaboratorRequest](#ttn.lorawan.v3.SetApplicationCollaboratorRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) | Setting a collaborator without rights, removes them. |
| ListCollaborators | [ListApplicationCollaboratorsRequest](#ttn.lorawan.v3.ListApplicationCollaboratorsRequest) | [Collaborators](#ttn.lorawan.v3.Collaborators) |  |


<a name="ttn.lorawan.v3.ApplicationRegistry"></a>

### ApplicationRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateApplicationRequest](#ttn.lorawan.v3.CreateApplicationRequest) | [Application](#ttn.lorawan.v3.Application) | Create a new application. This also sets the given organization or user as first collaborator with all possible rights. |
| Get | [GetApplicationRequest](#ttn.lorawan.v3.GetApplicationRequest) | [Application](#ttn.lorawan.v3.Application) | Get the application with the given identifiers, selecting the fields given by the field mask. The method may return more or less fields, depending on the rights of the caller. |
| List | [ListApplicationsRequest](#ttn.lorawan.v3.ListApplicationsRequest) | [Applications](#ttn.lorawan.v3.Applications) | List applications. See request message for details. |
| Update | [UpdateApplicationRequest](#ttn.lorawan.v3.UpdateApplicationRequest) | [Application](#ttn.lorawan.v3.Application) |  |
| Delete | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



<a name="lorawan-stack/api/applicationserver.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/applicationserver.proto



<a name="ttn.lorawan.v3.ApplicationLink"></a>

### ApplicationLink



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| network_server_address | [string](#string) |  | The address of the external Network Server where to link to. The typical format of the address is &#34;host:port&#34;. If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. Leave empty when linking to a cluster Network Server. |
| api_key | [string](#string) |  |  |
| default_formatters | [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters) |  |  |






<a name="ttn.lorawan.v3.ApplicationLinkStats"></a>

### ApplicationLinkStats
Link stats as monitored by the Application Server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| linked_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| network_server_address | [string](#string) |  |  |
| last_up_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the last upstream message has been received from a Network Server. This can be a join-accept, uplink message or downlink message event. |
| up_count | [uint64](#uint64) |  | Number of upstream messages received. |
| last_downlink_forwarded_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Timestamp when the last downlink message has been forwarded to a Network Server. |
| downlink_count | [uint64](#uint64) |  | Number of downlink messages forwarded. |






<a name="ttn.lorawan.v3.GetApplicationLinkRequest"></a>

### GetApplicationLinkRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.SetApplicationLinkRequest"></a>

### SetApplicationLinkRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| link | [ApplicationLink](#ttn.lorawan.v3.ApplicationLink) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 

 

 


<a name="ttn.lorawan.v3.AppAs"></a>

### AppAs
The AppAs service connects an application or integration to an Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Subscribe | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) | [ApplicationUp](#ttn.lorawan.v3.ApplicationUp) stream |  |
| DownlinkQueuePush | [DownlinkQueueRequest](#ttn.lorawan.v3.DownlinkQueueRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| DownlinkQueueReplace | [DownlinkQueueRequest](#ttn.lorawan.v3.DownlinkQueueRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| DownlinkQueueList | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | [ApplicationDownlinks](#ttn.lorawan.v3.ApplicationDownlinks) |  |


<a name="ttn.lorawan.v3.As"></a>

### As
The As service manages the Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetLink | [GetApplicationLinkRequest](#ttn.lorawan.v3.GetApplicationLinkRequest) | [ApplicationLink](#ttn.lorawan.v3.ApplicationLink) |  |
| SetLink | [SetApplicationLinkRequest](#ttn.lorawan.v3.SetApplicationLinkRequest) | [ApplicationLink](#ttn.lorawan.v3.ApplicationLink) | Set a link configuration from the Application Server a Network Server. This call returns immediately after setting the link configuration; it does not wait for a link to establish. To get link statistics or errors, use the `GetLinkStats` call. |
| DeleteLink | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| GetLinkStats | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) | [ApplicationLinkStats](#ttn.lorawan.v3.ApplicationLinkStats) | GetLinkStats returns the link statistics. This call returns a NotFound error code if there is no link for the given application identifiers. This call returns the error code of the link error if linking to a Network Server failed. |


<a name="ttn.lorawan.v3.AsEndDeviceRegistry"></a>

### AsEndDeviceRegistry
The AsEndDeviceRegistry service allows clients to manage their end devices on the Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Get | [GetEndDeviceRequest](#ttn.lorawan.v3.GetEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Get returns the device that matches the given identifiers. If there are multiple matches, an error will be returned. |
| Set | [SetEndDeviceRequest](#ttn.lorawan.v3.SetEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Set creates or updates the device. |
| Delete | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) | Delete deletes the device that matches the given identifiers. If there are multiple matches, an error will be returned. |

 



<a name="lorawan-stack/api/applicationserver_web.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/applicationserver_web.proto



<a name="ttn.lorawan.v3.ApplicationWebhook"></a>

### ApplicationWebhook



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [ApplicationWebhookIdentifiers](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| base_url | [string](#string) |  | Base URL to which the message&#39;s path is appended. |
| headers | [ApplicationWebhook.HeadersEntry](#ttn.lorawan.v3.ApplicationWebhook.HeadersEntry) | repeated | HTTP headers to use. |
| format | [string](#string) |  | The format to use for the body. Supported values depend on the Application Server configuration. |
| uplink_message | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| join_accept | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| downlink_ack | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| downlink_nack | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| downlink_sent | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| downlink_failed | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| downlink_queued | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| location_solved | [ApplicationWebhook.Message](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |






<a name="ttn.lorawan.v3.ApplicationWebhook.HeadersEntry"></a>

### ApplicationWebhook.HeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.ApplicationWebhook.Message"></a>

### ApplicationWebhook.Message



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  | Path to append to the base URL. |






<a name="ttn.lorawan.v3.ApplicationWebhookFormats"></a>

### ApplicationWebhookFormats



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| formats | [ApplicationWebhookFormats.FormatsEntry](#ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry) | repeated | Format and description. |






<a name="ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry"></a>

### ApplicationWebhookFormats.FormatsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.ApplicationWebhookIdentifiers"></a>

### ApplicationWebhookIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| webhook_id | [string](#string) |  |  |






<a name="ttn.lorawan.v3.ApplicationWebhooks"></a>

### ApplicationWebhooks



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| webhooks | [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook) | repeated |  |






<a name="ttn.lorawan.v3.GetApplicationWebhookRequest"></a>

### GetApplicationWebhookRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [ApplicationWebhookIdentifiers](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.ListApplicationWebhooksRequest"></a>

### ListApplicationWebhooksRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.SetApplicationWebhookRequest"></a>

### SetApplicationWebhookRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| webhook | [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 

 

 


<a name="ttn.lorawan.v3.ApplicationWebhookRegistry"></a>

### ApplicationWebhookRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetFormats | [.google.protobuf.Empty](#google.protobuf.Empty) | [ApplicationWebhookFormats](#ttn.lorawan.v3.ApplicationWebhookFormats) |  |
| Get | [GetApplicationWebhookRequest](#ttn.lorawan.v3.GetApplicationWebhookRequest) | [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook) |  |
| List | [ListApplicationWebhooksRequest](#ttn.lorawan.v3.ListApplicationWebhooksRequest) | [ApplicationWebhooks](#ttn.lorawan.v3.ApplicationWebhooks) |  |
| Set | [SetApplicationWebhookRequest](#ttn.lorawan.v3.SetApplicationWebhookRequest) | [ApplicationWebhook](#ttn.lorawan.v3.ApplicationWebhook) |  |
| Delete | [ApplicationWebhookIdentifiers](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



<a name="lorawan-stack/api/client.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/client.proto



<a name="ttn.lorawan.v3.Client"></a>

### Client
An OAuth client on the network.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| attributes | [Client.AttributesEntry](#ttn.lorawan.v3.Client.AttributesEntry) | repeated |  |
| contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  |
| secret | [string](#string) |  | The client secret is only visible to collaborators of the client. |
| redirect_uris | [string](#string) | repeated | The allowed redirect URIs against which authorization requests are checked. If the authorization request does not pass a redirect URI, the first one from this list is taken. |
| state | [State](#ttn.lorawan.v3.State) |  | The reviewing state of the client. This field can only be modified by admins. |
| skip_authorization | [bool](#bool) |  | If set, the authorization page will be skipped. This field can only be modified by admins. |
| endorsed | [bool](#bool) |  | If set, the authorization page will show endorsement. This field can only be modified by admins. |
| grants | [GrantType](#ttn.lorawan.v3.GrantType) | repeated | OAuth flows that can be used for the client to get a token. After a client is created, this field can only be modified by admins. |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated | Rights denotes what rights the client will have access to. Users that previously authorized this client will have to re-authorize the client after rights are added to this list. |






<a name="ttn.lorawan.v3.Client.AttributesEntry"></a>

### Client.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.Clients"></a>

### Clients



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clients | [Client](#ttn.lorawan.v3.Client) | repeated |  |






<a name="ttn.lorawan.v3.CreateClientRequest"></a>

### CreateClientRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client | [Client](#ttn.lorawan.v3.Client) |  |  |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created client. |






<a name="ttn.lorawan.v3.GetClientRequest"></a>

### GetClientRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.ListClientCollaboratorsRequest"></a>

### ListClientCollaboratorsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListClientsRequest"></a>

### ListClientsRequest
By default we list all OAuth clients the caller has rights on.
Set the user or the organization (not both) to instead list the OAuth clients
where the user or organization is collaborator on.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.SetClientCollaboratorRequest"></a>

### SetClientCollaboratorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  |






<a name="ttn.lorawan.v3.UpdateClientRequest"></a>

### UpdateClientRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client | [Client](#ttn.lorawan.v3.Client) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 


<a name="ttn.lorawan.v3.GrantType"></a>

### GrantType
The OAuth2 flows an OAuth client can use to get an access token.

| Name | Number | Description |
| ---- | ------ | ----------- |
| GRANT_AUTHORIZATION_CODE | 0 | Grant type used to exchange an authorization code for an access token. |
| GRANT_PASSWORD | 1 | Grant type used to exchange a user ID and password for an access token. |
| GRANT_REFRESH_TOKEN | 2 | Grant type used to exchange a refresh token for an access token. |


 

 

 



<a name="lorawan-stack/api/client_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/client_services.proto


 

 

 


<a name="ttn.lorawan.v3.ClientAccess"></a>

### ClientAccess


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListRights | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) | [Rights](#ttn.lorawan.v3.Rights) |  |
| SetCollaborator | [SetClientCollaboratorRequest](#ttn.lorawan.v3.SetClientCollaboratorRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) | Set the rights of a collaborator on the OAuth client. Users or organizations are considered to be a collaborator if they have at least one right on the OAuth client. |
| ListCollaborators | [ListClientCollaboratorsRequest](#ttn.lorawan.v3.ListClientCollaboratorsRequest) | [Collaborators](#ttn.lorawan.v3.Collaborators) |  |


<a name="ttn.lorawan.v3.ClientRegistry"></a>

### ClientRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateClientRequest](#ttn.lorawan.v3.CreateClientRequest) | [Client](#ttn.lorawan.v3.Client) | Create a new OAuth client. This also sets the given organization or user as first collaborator with all possible rights. |
| Get | [GetClientRequest](#ttn.lorawan.v3.GetClientRequest) | [Client](#ttn.lorawan.v3.Client) | Get the OAuth client with the given identifiers, selecting the fields given by the field mask. The method may return more or less fields, depending on the rights of the caller. |
| List | [ListClientsRequest](#ttn.lorawan.v3.ListClientsRequest) | [Clients](#ttn.lorawan.v3.Clients) | List OAuth clients. See request message for details. |
| Update | [UpdateClientRequest](#ttn.lorawan.v3.UpdateClientRequest) | [Client](#ttn.lorawan.v3.Client) |  |
| Delete | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



<a name="lorawan-stack/api/cluster.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/cluster.proto



<a name="ttn.lorawan.v3.PeerInfo"></a>

### PeerInfo
PeerInfo


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| grpc_port | [uint32](#uint32) |  | Port on which the gRPC server is exposed. |
| tls | [bool](#bool) |  | Indicates whether the gRPC server uses TLS. |
| roles | [PeerInfo.Role](#ttn.lorawan.v3.PeerInfo.Role) | repeated | Roles of the peer. |
| tags | [PeerInfo.TagsEntry](#ttn.lorawan.v3.PeerInfo.TagsEntry) | repeated | Tags of the peer |






<a name="ttn.lorawan.v3.PeerInfo.TagsEntry"></a>

### PeerInfo.TagsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="ttn.lorawan.v3.PeerInfo.Role"></a>

### PeerInfo.Role


| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 |  |
| ENTITY_REGISTRY | 1 |  |
| ACCESS | 2 |  |
| GATEWAY_SERVER | 3 |  |
| NETWORK_SERVER | 4 |  |
| APPLICATION_SERVER | 5 |  |
| JOIN_SERVER | 6 |  |
| CRYPTO_SERVER | 7 |  |


 

 

 



<a name="lorawan-stack/api/configuration_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/configuration_services.proto



<a name="ttn.lorawan.v3.FrequencyPlanDescription"></a>

### FrequencyPlanDescription



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| base_id | [string](#string) |  | The ID of the frequency that the current frequency plan is based on. |
| name | [string](#string) |  |  |
| base_frequency | [uint32](#uint32) |  | Base frequency in MHz for hardware support (433, 470, 868 or 915) |






<a name="ttn.lorawan.v3.ListFrequencyPlansRequest"></a>

### ListFrequencyPlansRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| base_frequency | [uint32](#uint32) |  | Optional base frequency in MHz for hardware support (433, 470, 868 or 915) |






<a name="ttn.lorawan.v3.ListFrequencyPlansResponse"></a>

### ListFrequencyPlansResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency_plans | [FrequencyPlanDescription](#ttn.lorawan.v3.FrequencyPlanDescription) | repeated |  |





 

 

 


<a name="ttn.lorawan.v3.Configuration"></a>

### Configuration


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListFrequencyPlans | [ListFrequencyPlansRequest](#ttn.lorawan.v3.ListFrequencyPlansRequest) | [ListFrequencyPlansResponse](#ttn.lorawan.v3.ListFrequencyPlansResponse) |  |

 



<a name="lorawan-stack/api/contact_info.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/contact_info.proto



<a name="ttn.lorawan.v3.ContactInfo"></a>

### ContactInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| contact_type | [ContactType](#ttn.lorawan.v3.ContactType) |  |  |
| contact_method | [ContactMethod](#ttn.lorawan.v3.ContactMethod) |  |  |
| value | [string](#string) |  |  |
| public | [bool](#bool) |  |  |
| validated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="ttn.lorawan.v3.ContactInfoValidation"></a>

### ContactInfoValidation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| token | [string](#string) |  |  |
| entity | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) |  |  |
| contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 


<a name="ttn.lorawan.v3.ContactMethod"></a>

### ContactMethod


| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTACT_METHOD_OTHER | 0 |  |
| CONTACT_METHOD_EMAIL | 1 |  |
| CONTACT_METHOD_PHONE | 2 |  |



<a name="ttn.lorawan.v3.ContactType"></a>

### ContactType


| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTACT_TYPE_OTHER | 0 |  |
| CONTACT_TYPE_ABUSE | 1 |  |
| CONTACT_TYPE_BILLING | 2 |  |
| CONTACT_TYPE_TECHNICAL | 3 |  |


 

 


<a name="ttn.lorawan.v3.ContactInfoRegistry"></a>

### ContactInfoRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| RequestValidation | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | [ContactInfoValidation](#ttn.lorawan.v3.ContactInfoValidation) | Request validation for the non-validated contact info for the given entity. |
| Validate | [ContactInfoValidation](#ttn.lorawan.v3.ContactInfoValidation) | [.google.protobuf.Empty](#google.protobuf.Empty) | Validate confirms a contact info validation. |

 



<a name="lorawan-stack/api/end_device.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/end_device.proto



<a name="ttn.lorawan.v3.CreateEndDeviceRequest"></a>

### CreateEndDeviceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| end_device | [EndDevice](#ttn.lorawan.v3.EndDevice) |  |  |






<a name="ttn.lorawan.v3.EndDevice"></a>

### EndDevice
Defines an End Device registration and its state on the network.
The persistence of the EndDevice is divided between the Network Server, Application Server and Join Server.
SDKs are responsible for combining (if desired) the three.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| name | [string](#string) |  | Friendly name of the device. Stored in Entity Registry. |
| description | [string](#string) |  | Description of the device. Stored in Entity Registry. |
| attributes | [EndDevice.AttributesEntry](#ttn.lorawan.v3.EndDevice.AttributesEntry) | repeated | Attributes of the device. Stored in Entity Registry. |
| version_ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | Version Identifiers. Stored in Entity Registry, Network Server and Application Server. |
| service_profile_id | [string](#string) |  | Default service profile. Stored in Entity Registry. |
| network_server_address | [string](#string) |  | The address of the Network Server where this device is supposed to be registered. Stored in Entity Registry and Join Server. The typical format of the address is &#34;host:port&#34;. If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| application_server_address | [string](#string) |  | The address of the Application Server where this device is supposed to be registered. Stored in Entity Registry and Join Server. The typical format of the address is &#34;host:port&#34;. If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| join_server_address | [string](#string) |  | The address of the Join Server where this device is supposed to be registered. Stored in Entity Registry. The typical format of the address is &#34;host:port&#34;. If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| locations | [EndDevice.LocationsEntry](#ttn.lorawan.v3.EndDevice.LocationsEntry) | repeated | Location of the device. Stored in Entity Registry. |
| supports_class_b | [bool](#bool) |  | Whether the device supports class B. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| supports_class_c | [bool](#bool) |  | Whether the device supports class C. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| lorawan_phy_version | [PHYVersion](#ttn.lorawan.v3.PHYVersion) |  | LoRaWAN PHY version. Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| frequency_plan_id | [string](#string) |  | ID of the frequency plan used by this device. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| min_frequency | [uint64](#uint64) |  | Minimum frequency the device is capable of using (Hz). Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| max_frequency | [uint64](#uint64) |  | Maximum frequency the device is capable of using (Hz). Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| supports_join | [bool](#bool) |  | The device supports join (it&#39;s OTAA). Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| resets_join_nonces | [bool](#bool) |  | Whether the device resets the join and dev nonces (not LoRaWAN 1.1 compliant). Stored in Join Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| root_keys | [RootKeys](#ttn.lorawan.v3.RootKeys) |  | Device root keys. Stored in Join Server. |
| net_id | [bytes](#bytes) |  | Home NetID. Stored in Join Server. |
| mac_settings | [MACSettings](#ttn.lorawan.v3.MACSettings) |  | Settings for how the Network Server handles MAC layer for this device. Stored in Network Server. |
| mac_state | [MACState](#ttn.lorawan.v3.MACState) |  | MAC state of the device. Stored in Network Server. |
| session | [Session](#ttn.lorawan.v3.Session) |  | Current session of the device. Stored in Network Server and Application Server. |
| pending_session | [Session](#ttn.lorawan.v3.Session) |  | Pending session. Stored in Network Server and Application Server until RekeyInd is received. |
| last_dev_nonce | [uint32](#uint32) |  | Last DevNonce used. This field is only used for devices using LoRaWAN version 1.1 and later. Stored in Join Server. |
| used_dev_nonces | [uint32](#uint32) | repeated | Used DevNonces sorted in ascending order. This field is only used for devices using LoRaWAN versions preceding 1.1. Stored in Join Server. |
| last_join_nonce | [uint32](#uint32) |  | Last JoinNonce/AppNonce(for devices using LoRaWAN versions preceding 1.1) used. Stored in Join Server. |
| last_rj_count_0 | [uint32](#uint32) |  | Last Rejoin counter value used (type 0/2). Stored in Join Server. |
| last_rj_count_1 | [uint32](#uint32) |  | Last Rejoin counter value used (type 1). Stored in Join Server. |
| last_dev_status_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when last DevStatus MAC command was received. Stored in Network Server. |
| power_state | [PowerState](#ttn.lorawan.v3.PowerState) |  | The power state of the device; whether it is battery-powered or connected to an external power source. Received via the DevStatus MAC command at status_received_at. Stored in Network Server. |
| battery_percentage | [google.protobuf.FloatValue](#google.protobuf.FloatValue) |  | Latest-known battery percentage of the device. Received via the DevStatus MAC command at last_dev_status_received_at or earlier. Stored in Network Server. |
| downlink_margin | [int32](#int32) |  | Demodulation signal-to-noise ratio (dB). Received via the DevStatus MAC command at last_dev_status_received_at. Stored in Network Server. |
| recent_adr_uplinks | [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | repeated | Recent uplink messages with ADR bit set to 1 sorted by time. Stored in Network Server. The field is reset each time an uplink message carrying MACPayload is received with ADR bit set to 0. The number of messages stored is in the range [0,20]; |
| recent_uplinks | [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | repeated | Recent uplink messages sorted by time. Stored in Network Server. The number of messages stored may depend on configuration. |
| recent_downlinks | [DownlinkMessage](#ttn.lorawan.v3.DownlinkMessage) | repeated | Recent downlink messages sorted by time. Stored in Network Server. The number of messages stored may depend on configuration. |
| queued_application_downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated | Queued Application downlink messages. Stored in Application Server, which sets them on the Network Server. |
| formatters | [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters) |  | The payload formatters for this end device. Stored in Application Server. Copied on creation from template identified by version_ids. |
| provisioner_id | [string](#string) |  | ID of the provisioner. Stored in Join Server. |
| provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  | Vendor-specific provisioning data. Stored in Join Server. |






<a name="ttn.lorawan.v3.EndDevice.AttributesEntry"></a>

### EndDevice.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.EndDevice.LocationsEntry"></a>

### EndDevice.LocationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Location](#ttn.lorawan.v3.Location) |  |  |






<a name="ttn.lorawan.v3.EndDeviceBrand"></a>

### EndDeviceBrand



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| url | [string](#string) |  |  |
| logos | [string](#string) | repeated | Logos contains file names of brand logos. |






<a name="ttn.lorawan.v3.EndDeviceModel"></a>

### EndDeviceModel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| brand_id | [string](#string) |  |  |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |






<a name="ttn.lorawan.v3.EndDeviceVersion"></a>

### EndDeviceVersion
Template for creating end devices.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | Version identifiers. |
| lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. |
| lorawan_phy_version | [PHYVersion](#ttn.lorawan.v3.PHYVersion) |  | LoRaWAN PHY version. |
| frequency_plan_id | [string](#string) |  | ID of the frequency plan used by this device. |
| photos | [string](#string) | repeated | Photos contains file names of device photos. |
| supports_class_b | [bool](#bool) |  | Whether the device supports class B. |
| supports_class_c | [bool](#bool) |  | Whether the device supports class C. |
| default_mac_settings | [MACSettings](#ttn.lorawan.v3.MACSettings) |  | Default MAC layer settings of the device. |
| min_frequency | [uint64](#uint64) |  | Minimum frequency the device is capable of using (Hz). |
| max_frequency | [uint64](#uint64) |  | Maximum frequency the device is capable of using (Hz). |
| supports_join | [bool](#bool) |  | The device supports join (it&#39;s OTAA). |
| resets_join_nonces | [bool](#bool) |  | Whether the device resets the join and dev nonces (not LoRaWAN 1.1 compliant). |
| default_formatters | [MessagePayloadFormatters](#ttn.lorawan.v3.MessagePayloadFormatters) |  | Default formatters defining the payload formats for this end device. |






<a name="ttn.lorawan.v3.EndDeviceVersionIdentifiers"></a>

### EndDeviceVersionIdentifiers
Identifies an end device model with version information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| brand_id | [string](#string) |  |  |
| model_id | [string](#string) |  |  |
| hardware_version | [string](#string) |  |  |
| firmware_version | [string](#string) |  |  |






<a name="ttn.lorawan.v3.EndDevices"></a>

### EndDevices



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| end_devices | [EndDevice](#ttn.lorawan.v3.EndDevice) | repeated |  |






<a name="ttn.lorawan.v3.GetEndDeviceRequest"></a>

### GetEndDeviceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.ListEndDevicesRequest"></a>

### ListEndDevicesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.MACParameters"></a>

### MACParameters
MACParameters represent the parameters of the device&#39;s MAC layer (active or desired).
This is used internally by the Network Server and is read only.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_eirp | [float](#float) |  | Maximum EIRP (dBm). |
| uplink_dwell_time | [bool](#bool) |  | Uplink dwell time is set (400ms). |
| downlink_dwell_time | [bool](#bool) |  | Downlink dwell time is set (400ms). |
| adr_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | ADR: data rate index to use. |
| adr_tx_power_index | [uint32](#uint32) |  | ADR: transmission power index to use. |
| adr_nb_trans | [uint32](#uint32) |  | ADR: number of retransmissions. |
| adr_ack_limit | [uint32](#uint32) |  | ADR: number of messages to wait before setting ADRAckReq. |
| adr_ack_delay | [uint32](#uint32) |  | ADR: number of messages to wait after setting ADRAckReq and before changing TxPower or DataRate. |
| rx1_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  | Rx1 delay (Rx2 delay is Rx1 delay &#43; 1 second). |
| rx1_data_rate_offset | [uint32](#uint32) |  | Data rate offset for Rx1. |
| rx2_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Data rate index for Rx2. |
| rx2_frequency | [uint64](#uint64) |  | Frequency for Rx2 (Hz). |
| max_duty_cycle | [AggregatedDutyCycle](#ttn.lorawan.v3.AggregatedDutyCycle) |  | Maximum uplink duty cycle (of all channels). |
| rejoin_time_periodicity | [RejoinTimeExponent](#ttn.lorawan.v3.RejoinTimeExponent) |  | Time within which a rejoin-request must be sent. |
| rejoin_count_periodicity | [RejoinCountExponent](#ttn.lorawan.v3.RejoinCountExponent) |  | Message count within which a rejoin-request must be sent. |
| ping_slot_frequency | [uint64](#uint64) |  | Frequency of the class B ping slot (Hz). |
| ping_slot_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Data rate index of the class B ping slot. |
| beacon_frequency | [uint64](#uint64) |  | Frequency of the class B beacon (Hz). |
| channels | [MACParameters.Channel](#ttn.lorawan.v3.MACParameters.Channel) | repeated | Configured uplink channels and optionally Rx1 frequency. |






<a name="ttn.lorawan.v3.MACParameters.Channel"></a>

### MACParameters.Channel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uplink_frequency | [uint64](#uint64) |  | Uplink frequency of the channel (Hz). |
| downlink_frequency | [uint64](#uint64) |  | Downlink frequency of the channel (Hz). |
| min_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Index of the minimum data rate for uplink. |
| max_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | Index of the maximum data rate for uplink. |
| enable_uplink | [bool](#bool) |  | Channel can be used by device for uplink. |






<a name="ttn.lorawan.v3.MACSettings"></a>

### MACSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| class_b_timeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | Maximum delay for the device to answer a MAC request or a confirmed downlink frame. If unset, the default value from Network Server configuration will be used. |
| ping_slot_periodicity | [MACSettings.PingSlotPeriodValue](#ttn.lorawan.v3.MACSettings.PingSlotPeriodValue) |  | Periodicity of the class B ping slot. If unset, the default value from Network Server configuration will be used. |
| ping_slot_data_rate_index | [MACSettings.DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue) |  | Data rate index of the class B ping slot. If unset, the default value from Network Server configuration will be used. |
| ping_slot_frequency | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  | Frequency of the class B ping slot (Hz). If unset, the default value from Network Server configuration will be used. |
| class_c_timeout | [google.protobuf.Duration](#google.protobuf.Duration) |  | Maximum delay for the device to answer a MAC request or a confirmed downlink frame. If unset, the default value from Network Server configuration will be used. |
| rx1_delay | [MACSettings.RxDelayValue](#ttn.lorawan.v3.MACSettings.RxDelayValue) |  | Class A Rx1 delay. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| rx1_data_rate_offset | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  | Rx1 data rate offset. If unset, the default value from Network Server configuration will be used. |
| rx2_data_rate_index | [MACSettings.DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue) |  | Data rate index for Rx2. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| rx2_frequency | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  | Frequency for Rx2 (Hz). If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| factory_preset_frequencies | [uint64](#uint64) | repeated | List of factory-preset frequencies. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| max_duty_cycle | [MACSettings.AggregatedDutyCycleValue](#ttn.lorawan.v3.MACSettings.AggregatedDutyCycleValue) |  | Maximum uplink duty cycle (of all channels). |
| supports_32_bit_f_cnt | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether the device supports 32-bit frame counters. If unset, the default value from Network Server configuration will be used. |
| use_adr | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether the Network Server should use ADR for the device. If unset, the default value from Network Server configuration will be used. |
| adr_margin | [google.protobuf.FloatValue](#google.protobuf.FloatValue) |  | The ADR margin tells the network server how much margin it should add in ADR requests. A bigger margin is less efficient, but gives a better chance of successful reception. If unset, the default value from Network Server configuration will be used. |
| resets_f_cnt | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Whether the device resets the frame counters (not LoRaWAN compliant). If unset, the default value from Network Server configuration will be used. |
| status_time_periodicity | [google.protobuf.Duration](#google.protobuf.Duration) |  | The interval after which a DevStatusReq MACCommand shall be sent. If unset, the default value from Network Server configuration will be used. |
| status_count_periodicity | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  | Number of uplink messages after which a DevStatusReq MACCommand shall be sent. If unset, the default value from Network Server configuration will be used. |
| desired_rx1_delay | [MACSettings.RxDelayValue](#ttn.lorawan.v3.MACSettings.RxDelayValue) |  | The Rx1 delay Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| desired_rx1_data_rate_offset | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  | The Rx1 data rate offset Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from Network Server configuration will be used. |
| desired_rx2_data_rate_index | [MACSettings.DataRateIndexValue](#ttn.lorawan.v3.MACSettings.DataRateIndexValue) |  | The Rx2 data rate index Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from frequency plan, Network Server configuration or regional parameters specification will be used. |
| desired_rx2_frequency | [google.protobuf.UInt64Value](#google.protobuf.UInt64Value) |  | The Rx2 frequency index Network Server should configure device to use via MAC commands. If unset, the default value from frequency plan, Network Server configuration or regional parameters specification will be used. |






<a name="ttn.lorawan.v3.MACSettings.AggregatedDutyCycleValue"></a>

### MACSettings.AggregatedDutyCycleValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [AggregatedDutyCycle](#ttn.lorawan.v3.AggregatedDutyCycle) |  |  |






<a name="ttn.lorawan.v3.MACSettings.DataRateIndexValue"></a>

### MACSettings.DataRateIndexValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |






<a name="ttn.lorawan.v3.MACSettings.PingSlotPeriodValue"></a>

### MACSettings.PingSlotPeriodValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [PingSlotPeriod](#ttn.lorawan.v3.PingSlotPeriod) |  |  |






<a name="ttn.lorawan.v3.MACSettings.RxDelayValue"></a>

### MACSettings.RxDelayValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  |






<a name="ttn.lorawan.v3.MACState"></a>

### MACState
MACState represents the state of MAC layer of the device.
MACState is reset on each join for OTAA or ResetInd for ABP devices.
This is used internally by the Network Server and is read only.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| current_parameters | [MACParameters](#ttn.lorawan.v3.MACParameters) |  | Current LoRaWAN MAC parameters. |
| desired_parameters | [MACParameters](#ttn.lorawan.v3.MACParameters) |  | Desired LoRaWAN MAC parameters. |
| device_class | [Class](#ttn.lorawan.v3.Class) |  | Currently active LoRaWAN device class - Device class is A by default - If device sets ClassB bit in uplink, this will be set to B - If device sent DeviceModeInd MAC message, this will be set to that value |
| lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. |
| last_confirmed_downlink_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when the last confirmed downlink message or MAC command was scheduled. |
| last_dev_status_f_cnt_up | [uint32](#uint32) |  | Frame counter value of last uplink containing DevStatusAns. |
| ping_slot_periodicity | [PingSlotPeriod](#ttn.lorawan.v3.PingSlotPeriod) |  | Periodicity of the class B ping slot. |
| pending_application_downlink | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  | A confirmed application downlink, for which an acknowledgment is expected to arrive. |
| queued_responses | [MACCommand](#ttn.lorawan.v3.MACCommand) | repeated | Queued MAC responses. Regenerated on each uplink. |
| pending_requests | [MACCommand](#ttn.lorawan.v3.MACCommand) | repeated | Pending MAC requests(i.e. sent requests, for which no response has been received yet). Regenerated on each downlink. |
| queued_join_accept | [MACState.JoinAccept](#ttn.lorawan.v3.MACState.JoinAccept) |  | Queued join-accept. Set each time a (re-)join request accept is received from Join Server and removed each time a downlink is scheduled. |
| pending_join_request | [JoinRequest](#ttn.lorawan.v3.JoinRequest) |  | Pending join request. Set each time a join accept is scheduled and removed each time an uplink is received from the device. |
| rx_windows_available | [bool](#bool) |  | Whether or not Rx windows are expected to be open. Set to true every time an uplink is received. Set to false every time a successful downlink scheduling attempt is made. |






<a name="ttn.lorawan.v3.MACState.JoinAccept"></a>

### MACState.JoinAccept



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload | [bytes](#bytes) |  | Payload of the join-accept received from Join Server. |
| request | [JoinRequest](#ttn.lorawan.v3.JoinRequest) |  | JoinRequest sent to Join Server. |
| keys | [SessionKeys](#ttn.lorawan.v3.SessionKeys) |  | Network session keys associated with the join. |






<a name="ttn.lorawan.v3.Session"></a>

### Session



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dev_addr | [bytes](#bytes) |  | Device Address, issued by the Network Server or chosen by device manufacturer in case of testing range (beginning with 00-03). Known by Network Server, Application Server and Join Server. Owned by Network Server. |
| keys | [SessionKeys](#ttn.lorawan.v3.SessionKeys) |  |  |
| last_f_cnt_up | [uint32](#uint32) |  | Last uplink frame counter value used. Network Server only. Application Server assumes the Network Server checked it. |
| last_n_f_cnt_down | [uint32](#uint32) |  | Last network downlink frame counter value used. Network Server only. |
| last_a_f_cnt_down | [uint32](#uint32) |  | Last application downlink frame counter value used. Application Server only. |
| last_conf_f_cnt_down | [uint32](#uint32) |  | Frame counter of the last confirmed downlink message sent. Network Server only. |
| started_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when the session started. Network Server only. |






<a name="ttn.lorawan.v3.SetEndDeviceRequest"></a>

### SetEndDeviceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| end_device | [EndDevice](#ttn.lorawan.v3.EndDevice) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.UpdateEndDeviceRequest"></a>

### UpdateEndDeviceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| end_device | [EndDevice](#ttn.lorawan.v3.EndDevice) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 


<a name="ttn.lorawan.v3.PowerState"></a>

### PowerState
Power state of the device.

| Name | Number | Description |
| ---- | ------ | ----------- |
| POWER_UNKNOWN | 0 |  |
| POWER_BATTERY | 1 |  |
| POWER_EXTERNAL | 2 |  |


 

 

 



<a name="lorawan-stack/api/end_device_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/end_device_services.proto


 

 

 


<a name="ttn.lorawan.v3.EndDeviceRegistry"></a>

### EndDeviceRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateEndDeviceRequest](#ttn.lorawan.v3.CreateEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Create a new end device within an application. |
| Get | [GetEndDeviceRequest](#ttn.lorawan.v3.GetEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Get the end device with the given identifiers, selecting the fields given by the field mask. |
| List | [ListEndDevicesRequest](#ttn.lorawan.v3.ListEndDevicesRequest) | [EndDevices](#ttn.lorawan.v3.EndDevices) | List applications. See request message for details. |
| Update | [UpdateEndDeviceRequest](#ttn.lorawan.v3.UpdateEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) |  |
| Delete | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



<a name="lorawan-stack/api/enums.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/enums.proto


 


<a name="ttn.lorawan.v3.DownlinkPathConstraint"></a>

### DownlinkPathConstraint


| Name | Number | Description |
| ---- | ------ | ----------- |
| DOWNLINK_PATH_CONSTRAINT_NONE | 0 | Indicates that the gateway can be selected for downlink without constraints by the Network Server. |
| DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER | 1 | Indicates that the gateway can be selected for downlink only if no other or better gateway can be selected. |
| DOWNLINK_PATH_CONSTRAINT_NEVER | 2 | Indicates that this gateway will never be selected for downlink, even if that results in no available downlink path. |



<a name="ttn.lorawan.v3.State"></a>

### State
State enum defines states that an entity can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_REQUESTED | 0 | Denotes that the entity has been requested and is pending review by an admin. |
| STATE_APPROVED | 1 | Denotes that the entity has been reviewed and approved by an admin. |
| STATE_REJECTED | 2 | Denotes that the entity has been reviewed and rejected by an admin. |
| STATE_FLAGGED | 3 | Denotes that the entity has been flagged and is pending review by an admin. |
| STATE_SUSPENDED | 4 | Denotes that the entity has been reviewed and suspended by an admin. |


 

 

 



<a name="lorawan-stack/api/error.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/error.proto



<a name="ttn.lorawan.v3.ErrorDetails"></a>

### ErrorDetails
Error details that are communicated over gRPC (and HTTP) APIs.
The messages (for translation) are stored as &#34;error:&lt;namespace&gt;:&lt;name&gt;&#34;.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | [string](#string) |  | Namespace of the error (typically the package name in the stack). |
| name | [string](#string) |  | Name of the error. |
| message_format | [string](#string) |  | The default (fallback) message format that should be used for the error. This is also used if the client does not have a translation for the error. |
| attributes | [google.protobuf.Struct](#google.protobuf.Struct) |  | Attributes that should be filled into the message format. Any extra attributes can be displayed as error details. |
| correlation_id | [string](#string) |  | The correlation ID of the error can be used to correlate the error to stack traces the network may (or may not) store about recent errors. |
| cause | [ErrorDetails](#ttn.lorawan.v3.ErrorDetails) |  | The error that caused this error. |





 

 

 

 



<a name="lorawan-stack/api/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/events.proto



<a name="ttn.lorawan.v3.Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| identifiers | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | repeated |  |
| data | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| correlation_ids | [string](#string) | repeated |  |
| origin | [string](#string) |  |  |
| context | [Event.ContextEntry](#ttn.lorawan.v3.Event.ContextEntry) | repeated |  |






<a name="ttn.lorawan.v3.Event.ContextEntry"></a>

### Event.ContextEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bytes](#bytes) |  |  |






<a name="ttn.lorawan.v3.StreamEventsRequest"></a>

### StreamEventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifiers | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | repeated |  |
| tail | [uint32](#uint32) |  | If greater than zero, this will return historical events, up to this maximum when the stream starts. If used in combination with &#34;after&#34;, the limit that is reached first, is used. The availability of historical events depends on server support and retention policy. |
| after | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | If not empty, this will return historical events after the given time when the stream starts. If used in combination with &#34;tail&#34;, the limit that is reached first, is used. The availability of historical events depends on server support and retention policy. |





 

 

 


<a name="ttn.lorawan.v3.Events"></a>

### Events
The Events service serves events from the cluster.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Stream | [StreamEventsRequest](#ttn.lorawan.v3.StreamEventsRequest) | [Event](#ttn.lorawan.v3.Event) stream | Stream live events, optionally with a tail of historical events (depending on server support and retention policy). Events may arrive out-of-order. |

 



<a name="lorawan-stack/api/gateway.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/gateway.proto



<a name="ttn.lorawan.v3.CreateGatewayAPIKeyRequest"></a>

### CreateGatewayAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| name | [string](#string) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |






<a name="ttn.lorawan.v3.CreateGatewayRequest"></a>

### CreateGatewayRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway | [Gateway](#ttn.lorawan.v3.Gateway) |  |  |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created gateway. |






<a name="ttn.lorawan.v3.Gateway"></a>

### Gateway
Gateway is the message that defines a gateway on the network.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| attributes | [Gateway.AttributesEntry](#ttn.lorawan.v3.Gateway.AttributesEntry) | repeated |  |
| contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  |
| version_ids | [GatewayVersionIdentifiers](#ttn.lorawan.v3.GatewayVersionIdentifiers) |  |  |
| gateway_server_address | [string](#string) |  | The address of the Gateway Server to connect to. The typical format of the address is &#34;host:port&#34;. If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| auto_update | [bool](#bool) |  |  |
| update_channel | [string](#string) |  |  |
| frequency_plan_id | [string](#string) |  |  |
| antennas | [GatewayAntenna](#ttn.lorawan.v3.GatewayAntenna) | repeated |  |
| status_public | [bool](#bool) |  | The status of this gateway may be publicly displayed. |
| location_public | [bool](#bool) |  | The location of this gateway may be publicly displayed. |
| schedule_downlink_late | [bool](#bool) |  | Enable server-side buffering of downlink messages. This is recommended for gateways using the Semtech UDP Packet Forwarder v2.x or older, as it does not feature a just-in-time queue. If enabled, the Gateway Server schedules the downlink message late to the gateway so that it does not overwrite previously scheduled downlink messages that have not been transmitted yet. |
| enforce_duty_cycle | [bool](#bool) |  | Enforcing gateway duty cycle is recommended for all gateways to respect spectrum regulations. Disable enforcing the duty cycle only in controlled research and development environments. |
| downlink_path_constraint | [DownlinkPathConstraint](#ttn.lorawan.v3.DownlinkPathConstraint) |  |  |






<a name="ttn.lorawan.v3.Gateway.AttributesEntry"></a>

### Gateway.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.GatewayAntenna"></a>

### GatewayAntenna
GatewayAntenna is the message that defines a gateway antenna.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gain | [float](#float) |  | gain is the antenna gain relative to this gateway, in dBi. |
| location | [Location](#ttn.lorawan.v3.Location) |  | location is the antenna&#39;s location. |
| attributes | [GatewayAntenna.AttributesEntry](#ttn.lorawan.v3.GatewayAntenna.AttributesEntry) | repeated |  |






<a name="ttn.lorawan.v3.GatewayAntenna.AttributesEntry"></a>

### GatewayAntenna.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.GatewayBrand"></a>

### GatewayBrand



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| url | [string](#string) |  |  |
| logos | [string](#string) | repeated | Logos contains file names of brand logos. |






<a name="ttn.lorawan.v3.GatewayConnectionStats"></a>

### GatewayConnectionStats
Connection stats as monitored by the Gateway Server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connected_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| protocol | [string](#string) |  | Protocol used to connect (for example, udp, mqtt, grpc) |
| last_status_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| last_status | [GatewayStatus](#ttn.lorawan.v3.GatewayStatus) |  |  |
| last_uplink_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| uplink_count | [uint64](#uint64) |  |  |
| last_downlink_received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| downlink_count | [uint64](#uint64) |  |  |






<a name="ttn.lorawan.v3.GatewayModel"></a>

### GatewayModel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| brand_id | [string](#string) |  |  |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |






<a name="ttn.lorawan.v3.GatewayRadio"></a>

### GatewayRadio



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enable | [bool](#bool) |  |  |
| chip_type | [string](#string) |  |  |
| frequency | [uint64](#uint64) |  |  |
| rssi_offset | [float](#float) |  |  |
| tx_configuration | [GatewayRadio.TxConfiguration](#ttn.lorawan.v3.GatewayRadio.TxConfiguration) |  |  |






<a name="ttn.lorawan.v3.GatewayRadio.TxConfiguration"></a>

### GatewayRadio.TxConfiguration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| min_frequency | [uint64](#uint64) |  |  |
| max_frequency | [uint64](#uint64) |  |  |
| notch_frequency | [uint64](#uint64) |  |  |






<a name="ttn.lorawan.v3.GatewayStatus"></a>

### GatewayStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Current time of the gateway |
| boot_time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Boot time of the gateway - can be left out to save bandwidth; old value will be kept |
| versions | [GatewayStatus.VersionsEntry](#ttn.lorawan.v3.GatewayStatus.VersionsEntry) | repeated | Versions of gateway subsystems - each field can be left out to save bandwidth; old value will be kept - map keys are written in snake_case - for example: firmware: &#34;2.0.4&#34; forwarder: &#34;v2-3.3.1&#34; fpga: &#34;48&#34; dsp: &#34;27&#34; hal: &#34;v2-3.5.0&#34; |
| antenna_locations | [Location](#ttn.lorawan.v3.Location) | repeated | Location of each gateway&#39;s antenna - if left out, server uses registry-set location as fallback |
| ip | [string](#string) | repeated | IP addresses of this gateway. Repeated addresses can be used to communicate addresses of multiple interfaces (LAN, Public IP, ...). |
| metrics | [GatewayStatus.MetricsEntry](#ttn.lorawan.v3.GatewayStatus.MetricsEntry) | repeated | Metrics - can be used for forwarding gateway metrics such as temperatures or performance metrics - map keys are written in snake_case |
| advanced | [google.protobuf.Struct](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case |






<a name="ttn.lorawan.v3.GatewayStatus.MetricsEntry"></a>

### GatewayStatus.MetricsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [float](#float) |  |  |






<a name="ttn.lorawan.v3.GatewayStatus.VersionsEntry"></a>

### GatewayStatus.VersionsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.GatewayVersion"></a>

### GatewayVersion
Template for creating gateways.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [GatewayVersionIdentifiers](#ttn.lorawan.v3.GatewayVersionIdentifiers) |  | Version identifiers. |
| photos | [string](#string) | repeated | Photos contains file names of gateway photos. |
| radios | [GatewayRadio](#ttn.lorawan.v3.GatewayRadio) | repeated |  |
| clock_source | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.GatewayVersionIdentifiers"></a>

### GatewayVersionIdentifiers
Identifies an end device model with version information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| brand_id | [string](#string) |  |  |
| model_id | [string](#string) |  |  |
| hardware_version | [string](#string) |  |  |
| firmware_version | [string](#string) |  |  |






<a name="ttn.lorawan.v3.Gateways"></a>

### Gateways



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateways | [Gateway](#ttn.lorawan.v3.Gateway) | repeated |  |






<a name="ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest"></a>

### GetGatewayIdentifiersForEUIRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| eui | [bytes](#bytes) |  |  |






<a name="ttn.lorawan.v3.GetGatewayRequest"></a>

### GetGatewayRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.ListGatewayAPIKeysRequest"></a>

### ListGatewayAPIKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListGatewayCollaboratorsRequest"></a>

### ListGatewayCollaboratorsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListGatewaysRequest"></a>

### ListGatewaysRequest
By default we list all gateways the caller has rights on.
Set the user or the organization (not both) to instead list the gateways
where the user or organization is collaborator on.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.SetGatewayCollaboratorRequest"></a>

### SetGatewayCollaboratorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  |






<a name="ttn.lorawan.v3.UpdateGatewayAPIKeyRequest"></a>

### UpdateGatewayAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  |






<a name="ttn.lorawan.v3.UpdateGatewayRequest"></a>

### UpdateGatewayRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway | [Gateway](#ttn.lorawan.v3.Gateway) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 

 

 

 



<a name="lorawan-stack/api/gateway_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/gateway_services.proto



<a name="ttn.lorawan.v3.PullGatewayConfigurationRequest"></a>

### PullGatewayConfigurationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 

 

 


<a name="ttn.lorawan.v3.GatewayAccess"></a>

### GatewayAccess


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListRights | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) | [Rights](#ttn.lorawan.v3.Rights) |  |
| CreateAPIKey | [CreateGatewayAPIKeyRequest](#ttn.lorawan.v3.CreateGatewayAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) |  |
| ListAPIKeys | [ListGatewayAPIKeysRequest](#ttn.lorawan.v3.ListGatewayAPIKeysRequest) | [APIKeys](#ttn.lorawan.v3.APIKeys) |  |
| UpdateAPIKey | [UpdateGatewayAPIKeyRequest](#ttn.lorawan.v3.UpdateGatewayAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) | Update the rights of an existing gateway API key. To generate an API key, the CreateAPIKey should be used. To delete an API key, update it with zero rights. |
| SetCollaborator | [SetGatewayCollaboratorRequest](#ttn.lorawan.v3.SetGatewayCollaboratorRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) | Set the rights of a collaborator on the gateway. Users or organizations are considered to be a collaborator if they have at least one right on the gateway. |
| ListCollaborators | [ListGatewayCollaboratorsRequest](#ttn.lorawan.v3.ListGatewayCollaboratorsRequest) | [Collaborators](#ttn.lorawan.v3.Collaborators) |  |


<a name="ttn.lorawan.v3.GatewayConfigurator"></a>

### GatewayConfigurator


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| PullConfiguration | [PullGatewayConfigurationRequest](#ttn.lorawan.v3.PullGatewayConfigurationRequest) | [Gateway](#ttn.lorawan.v3.Gateway) stream |  |


<a name="ttn.lorawan.v3.GatewayRegistry"></a>

### GatewayRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateGatewayRequest](#ttn.lorawan.v3.CreateGatewayRequest) | [Gateway](#ttn.lorawan.v3.Gateway) | Create a new gateway. This also sets the given organization or user as first collaborator with all possible rights. |
| Get | [GetGatewayRequest](#ttn.lorawan.v3.GetGatewayRequest) | [Gateway](#ttn.lorawan.v3.Gateway) | Get the gateway with the given identifiers, selecting the fields given by the field mask. The method may return more or less fields, depending on the rights of the caller. |
| GetIdentifiersForEUI | [GetGatewayIdentifiersForEUIRequest](#ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest) | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |
| List | [ListGatewaysRequest](#ttn.lorawan.v3.ListGatewaysRequest) | [Gateways](#ttn.lorawan.v3.Gateways) | List gateways. See request message for details. |
| Update | [UpdateGatewayRequest](#ttn.lorawan.v3.UpdateGatewayRequest) | [Gateway](#ttn.lorawan.v3.Gateway) |  |
| Delete | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



<a name="lorawan-stack/api/gatewayserver.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/gatewayserver.proto



<a name="ttn.lorawan.v3.GatewayDown"></a>

### GatewayDown
GatewayDown contains downlink messages for the gateway.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| downlink_message | [DownlinkMessage](#ttn.lorawan.v3.DownlinkMessage) |  | DownlinkMessage for the gateway. |






<a name="ttn.lorawan.v3.GatewayUp"></a>

### GatewayUp
GatewayUp may contain zero or more uplink messages and/or a status message for the gateway.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uplink_messages | [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | repeated | UplinkMessages received by the gateway. |
| gateway_status | [GatewayStatus](#ttn.lorawan.v3.GatewayStatus) |  |  |
| tx_acknowledgment | [TxAcknowledgment](#ttn.lorawan.v3.TxAcknowledgment) |  |  |






<a name="ttn.lorawan.v3.ScheduleDownlinkResponse"></a>

### ScheduleDownlinkResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| delay | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |





 

 

 


<a name="ttn.lorawan.v3.Gs"></a>

### Gs


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetGatewayConnectionStats | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) | [GatewayConnectionStats](#ttn.lorawan.v3.GatewayConnectionStats) | Get statistics about the current gateway connection to the Gateway Server. This is not persisted between reconnects. |


<a name="ttn.lorawan.v3.GtwGs"></a>

### GtwGs
The GtwGs service connects a gateway to a Gateway Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| LinkGateway | [GatewayUp](#ttn.lorawan.v3.GatewayUp) stream | [GatewayDown](#ttn.lorawan.v3.GatewayDown) stream | Link the gateway to the Gateway Server. |
| GetConcentratorConfig | [.google.protobuf.Empty](#google.protobuf.Empty) | [ConcentratorConfig](#ttn.lorawan.v3.ConcentratorConfig) | GetConcentratorConfig associated to the gateway. |


<a name="ttn.lorawan.v3.NsGs"></a>

### NsGs
The NsGs service connects a Network Server to a Gateway Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ScheduleDownlink | [DownlinkMessage](#ttn.lorawan.v3.DownlinkMessage) | [ScheduleDownlinkResponse](#ttn.lorawan.v3.ScheduleDownlinkResponse) | ScheduleDownlink instructs the Gateway Server to schedule a downlink message. The Gateway Server may refuse if there are any conflicts in the schedule or if a duty cycle prevents the gateway from transmitting. |

 



<a name="lorawan-stack/api/identifiers.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/identifiers.proto



<a name="ttn.lorawan.v3.ApplicationIdentifiers"></a>

### ApplicationIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_id | [string](#string) |  |  |






<a name="ttn.lorawan.v3.ClientIdentifiers"></a>

### ClientIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_id | [string](#string) |  |  |






<a name="ttn.lorawan.v3.CombinedIdentifiers"></a>

### CombinedIdentifiers
Combine the identifiers of multiple entities.
The main purpose of this message is its use in events.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entity_identifiers | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) | repeated |  |






<a name="ttn.lorawan.v3.EndDeviceIdentifiers"></a>

### EndDeviceIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| device_id | [string](#string) |  |  |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| dev_eui | [bytes](#bytes) |  | The LoRaWAN DevEUI. |
| join_eui | [bytes](#bytes) |  | The LoRaWAN JoinEUI (or AppEUI for LoRaWAN 1.0 end devices). |
| dev_addr | [bytes](#bytes) |  | The LoRaWAN DevAddr. |






<a name="ttn.lorawan.v3.EntityIdentifiers"></a>

### EntityIdentifiers
EntityIdentifiers contains one of the possible entity identifiers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |






<a name="ttn.lorawan.v3.GatewayIdentifiers"></a>

### GatewayIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_id | [string](#string) |  |  |
| eui | [bytes](#bytes) |  | Secondary identifier, which can only be used in specific requests. |






<a name="ttn.lorawan.v3.OrganizationIdentifiers"></a>

### OrganizationIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_id | [string](#string) |  | This ID shares namespace with user IDs. |






<a name="ttn.lorawan.v3.OrganizationOrUserIdentifiers"></a>

### OrganizationOrUserIdentifiers
OrganizationOrUserIdentifiers contains either organization or user identifiers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |






<a name="ttn.lorawan.v3.UserIdentifiers"></a>

### UserIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | This ID shares namespace with organization IDs. |
| email | [string](#string) |  | Secondary identifier, which can only be used in specific requests. |





 

 

 

 



<a name="lorawan-stack/api/identityserver.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/identityserver.proto



<a name="ttn.lorawan.v3.AuthInfoResponse"></a>

### AuthInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| api_key | [AuthInfoResponse.APIKeyAccess](#ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess) |  |  |
| oauth_access_token | [OAuthAccessToken](#ttn.lorawan.v3.OAuthAccessToken) |  |  |
| universal_rights | [Rights](#ttn.lorawan.v3.Rights) |  |  |
| is_admin | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess"></a>

### AuthInfoResponse.APIKeyAccess



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  |
| entity_ids | [EntityIdentifiers](#ttn.lorawan.v3.EntityIdentifiers) |  |  |





 

 

 


<a name="ttn.lorawan.v3.EntityAccess"></a>

### EntityAccess


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| AuthInfo | [.google.protobuf.Empty](#google.protobuf.Empty) | [AuthInfoResponse](#ttn.lorawan.v3.AuthInfoResponse) | AuthInfo returns information about the authentication that is used on the request. |

 



<a name="lorawan-stack/api/join.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/join.proto



<a name="ttn.lorawan.v3.JoinRequest"></a>

### JoinRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw_payload | [bytes](#bytes) |  |  |
| payload | [Message](#ttn.lorawan.v3.Message) |  |  |
| dev_addr | [bytes](#bytes) |  |  |
| selected_mac_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  |  |
| net_id | [bytes](#bytes) |  |  |
| downlink_settings | [DLSettings](#ttn.lorawan.v3.DLSettings) |  |  |
| rx_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  |
| cf_list | [CFList](#ttn.lorawan.v3.CFList) |  | Optional CFList. |
| correlation_ids | [string](#string) | repeated |  |






<a name="ttn.lorawan.v3.JoinResponse"></a>

### JoinResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw_payload | [bytes](#bytes) |  |  |
| session_keys | [SessionKeys](#ttn.lorawan.v3.SessionKeys) |  |  |
| lifetime | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| correlation_ids | [string](#string) | repeated |  |





 

 

 

 



<a name="lorawan-stack/api/joinserver.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/joinserver.proto



<a name="ttn.lorawan.v3.AppSKeyResponse"></a>

### AppSKeyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| app_s_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Session Key. |






<a name="ttn.lorawan.v3.CryptoServicePayloadRequest"></a>

### CryptoServicePayloadRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  |  |
| payload | [bytes](#bytes) |  |  |
| provisioner_id | [string](#string) |  |  |
| provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |






<a name="ttn.lorawan.v3.CryptoServicePayloadResponse"></a>

### CryptoServicePayloadResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload | [bytes](#bytes) |  |  |






<a name="ttn.lorawan.v3.DeriveSessionKeysRequest"></a>

### DeriveSessionKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| lorawan_version | [MACVersion](#ttn.lorawan.v3.MACVersion) |  |  |
| join_nonce | [bytes](#bytes) |  |  |
| dev_nonce | [bytes](#bytes) |  |  |
| net_id | [bytes](#bytes) |  |  |
| provisioner_id | [string](#string) |  |  |
| provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |






<a name="ttn.lorawan.v3.GetRootKeysRequest"></a>

### GetRootKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| provisioner_id | [string](#string) |  |  |
| provisioning_data | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |






<a name="ttn.lorawan.v3.JoinAcceptMICRequest"></a>

### JoinAcceptMICRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload_request | [CryptoServicePayloadRequest](#ttn.lorawan.v3.CryptoServicePayloadRequest) |  |  |
| join_request_type | [RejoinType](#ttn.lorawan.v3.RejoinType) |  |  |
| dev_nonce | [bytes](#bytes) |  |  |






<a name="ttn.lorawan.v3.NwkSKeysResponse"></a>

### NwkSKeysResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| f_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Forwarding Network Session Integrity Key (or Network Session Key in 1.0 compatibility mode). |
| s_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Serving Network Session Integrity Key. |
| nwk_s_enc_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Session Encryption Key. |






<a name="ttn.lorawan.v3.ProvisionEndDevicesRequest"></a>

### ProvisionEndDevicesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| provisioner_id | [string](#string) |  | ID of the provisioner service as configured in the Join Server. |
| provisioning_data | [bytes](#bytes) |  | Vendor-specific provisioning data. |
| list | [ProvisionEndDevicesRequest.IdentifiersList](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList) |  | List of device identifiers that will be provisioned. The device identifiers must contain device_id and dev_eui. If set, the application_ids must equal the provision request&#39;s application_ids. The number of entries in data must match the number of given identifiers. |
| range | [ProvisionEndDevicesRequest.IdentifiersRange](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange) |  | Provision devices in a range. The device_id will be generated by the provisioner from the vendor-specific data. The dev_eui will be issued from the given start_dev_eui. |
| from_data | [ProvisionEndDevicesRequest.IdentifiersFromData](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData) |  | Provision devices with identifiers from the given data. The device_id and dev_eui will be generated by the provisioner from the vendor-specific data. |






<a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData"></a>

### ProvisionEndDevicesRequest.IdentifiersFromData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| join_eui | [bytes](#bytes) |  |  |






<a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList"></a>

### ProvisionEndDevicesRequest.IdentifiersList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| join_eui | [bytes](#bytes) |  |  |
| end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | repeated |  |






<a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange"></a>

### ProvisionEndDevicesRequest.IdentifiersRange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| join_eui | [bytes](#bytes) |  |  |
| start_dev_eui | [bytes](#bytes) |  | DevEUI to start issuing from. |






<a name="ttn.lorawan.v3.SessionKeyRequest"></a>

### SessionKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys. |
| dev_eui | [bytes](#bytes) |  | LoRaWAN DevEUI. |





 

 

 


<a name="ttn.lorawan.v3.ApplicationCryptoService"></a>

### ApplicationCryptoService
Service for application layer cryptographic operations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| DeriveAppSKey | [DeriveSessionKeysRequest](#ttn.lorawan.v3.DeriveSessionKeysRequest) | [AppSKeyResponse](#ttn.lorawan.v3.AppSKeyResponse) |  |
| GetAppKey | [GetRootKeysRequest](#ttn.lorawan.v3.GetRootKeysRequest) | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) | Get the AppKey. Crypto Servers may return status code UNIMPLEMENTED when root keys are not exposed. |


<a name="ttn.lorawan.v3.AsJs"></a>

### AsJs
The AsJs service connects an Application Server to a Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetAppSKey | [SessionKeyRequest](#ttn.lorawan.v3.SessionKeyRequest) | [AppSKeyResponse](#ttn.lorawan.v3.AppSKeyResponse) |  |


<a name="ttn.lorawan.v3.JsEndDeviceRegistry"></a>

### JsEndDeviceRegistry
The JsEndDeviceRegistry service allows clients to manage their end devices on the Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Get | [GetEndDeviceRequest](#ttn.lorawan.v3.GetEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Get returns the device that matches the given identifiers. If there are multiple matches, an error will be returned. |
| Set | [SetEndDeviceRequest](#ttn.lorawan.v3.SetEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Set creates or updates the device. |
| Provision | [ProvisionEndDevicesRequest](#ttn.lorawan.v3.ProvisionEndDevicesRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) stream | Provision returns end devices that are provisioned using the given vendor-specific data. The devices are not set in the registry. |
| Delete | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) | Delete deletes the device that matches the given identifiers. If there are multiple matches, an error will be returned. |


<a name="ttn.lorawan.v3.NetworkCryptoService"></a>

### NetworkCryptoService
Service for network layer cryptographic operations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| JoinRequestMIC | [CryptoServicePayloadRequest](#ttn.lorawan.v3.CryptoServicePayloadRequest) | [CryptoServicePayloadResponse](#ttn.lorawan.v3.CryptoServicePayloadResponse) |  |
| JoinAcceptMIC | [JoinAcceptMICRequest](#ttn.lorawan.v3.JoinAcceptMICRequest) | [CryptoServicePayloadResponse](#ttn.lorawan.v3.CryptoServicePayloadResponse) |  |
| EncryptJoinAccept | [CryptoServicePayloadRequest](#ttn.lorawan.v3.CryptoServicePayloadRequest) | [CryptoServicePayloadResponse](#ttn.lorawan.v3.CryptoServicePayloadResponse) |  |
| EncryptRejoinAccept | [CryptoServicePayloadRequest](#ttn.lorawan.v3.CryptoServicePayloadRequest) | [CryptoServicePayloadResponse](#ttn.lorawan.v3.CryptoServicePayloadResponse) |  |
| DeriveNwkSKeys | [DeriveSessionKeysRequest](#ttn.lorawan.v3.DeriveSessionKeysRequest) | [NwkSKeysResponse](#ttn.lorawan.v3.NwkSKeysResponse) |  |
| GetNwkKey | [GetRootKeysRequest](#ttn.lorawan.v3.GetRootKeysRequest) | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) | Get the NwkKey. Crypto Servers may return status code UNIMPLEMENTED when root keys are not exposed. |


<a name="ttn.lorawan.v3.NsJs"></a>

### NsJs
The NsJs service connects a Network Server to a Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| HandleJoin | [JoinRequest](#ttn.lorawan.v3.JoinRequest) | [JoinResponse](#ttn.lorawan.v3.JoinResponse) |  |
| GetNwkSKeys | [SessionKeyRequest](#ttn.lorawan.v3.SessionKeyRequest) | [NwkSKeysResponse](#ttn.lorawan.v3.NwkSKeysResponse) |  |

 



<a name="lorawan-stack/api/keys.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/keys.proto



<a name="ttn.lorawan.v3.KeyEnvelope"></a>

### KeyEnvelope



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [bytes](#bytes) |  | The (encrypted) key. |
| kek_label | [string](#string) |  | The label of the RFC 3394 key-encryption-key (KEK) that was used to encrypt the key. |






<a name="ttn.lorawan.v3.RootKeys"></a>

### RootKeys
Root keys for a LoRaWAN device.
These are stored on the Join Server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| root_key_id | [string](#string) |  | Join Server issued identifier for the root keys. |
| app_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Key. |
| nwk_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Key. |






<a name="ttn.lorawan.v3.SessionKeys"></a>

### SessionKeys
Session keys for a LoRaWAN session.
Only the components for which the keys were meant, will have the key-encryption-key (KEK) to decrypt the individual keys.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys. This ID can be used to request the keys from the Join Server in case the are lost. |
| f_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Forwarding Network Session Integrity Key (or Network Session Key in 1.0 compatibility mode). This key is stored by the (forwarding) Network Server. |
| s_nwk_s_int_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Serving Network Session Integrity Key. This key is stored by the (serving) Network Server. |
| nwk_s_enc_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Session Encryption Key. This key is stored by the (serving) Network Server. |
| app_s_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Session Key. This key is stored by the Application Server. |





 

 

 

 



<a name="lorawan-stack/api/lorawan.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/lorawan.proto



<a name="ttn.lorawan.v3.CFList"></a>

### CFList



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [CFListType](#ttn.lorawan.v3.CFListType) |  |  |
| freq | [uint32](#uint32) | repeated | Frequencies to be broadcasted, in hecto-Hz. These values are broadcasted as 24 bits unsigned integers. This field should not contain default values. |
| ch_masks | [bool](#bool) | repeated | ChMasks controlling the channels to be used. Length of this field must be equal to the amount of uplink channels defined by the selected frequency plan. |






<a name="ttn.lorawan.v3.DLSettings"></a>

### DLSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rx1_dr_offset | [uint32](#uint32) |  |  |
| rx2_dr | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |
| opt_neg | [bool](#bool) |  | OptNeg is set if Network Server implements LoRaWAN 1.1 or greater. |






<a name="ttn.lorawan.v3.DataRate"></a>

### DataRate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lora | [LoRaDataRate](#ttn.lorawan.v3.LoRaDataRate) |  |  |
| fsk | [FSKDataRate](#ttn.lorawan.v3.FSKDataRate) |  |  |






<a name="ttn.lorawan.v3.DownlinkPath"></a>

### DownlinkPath



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uplink_token | [bytes](#bytes) |  |  |
| fixed | [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers) |  |  |






<a name="ttn.lorawan.v3.FCtrl"></a>

### FCtrl



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| adr | [bool](#bool) |  |  |
| adr_ack_req | [bool](#bool) |  | Only on uplink. |
| ack | [bool](#bool) |  |  |
| f_pending | [bool](#bool) |  | Only on downlink. |
| class_b | [bool](#bool) |  | Only on uplink. |






<a name="ttn.lorawan.v3.FHDR"></a>

### FHDR



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dev_addr | [bytes](#bytes) |  |  |
| f_ctrl | [FCtrl](#ttn.lorawan.v3.FCtrl) |  |  |
| f_cnt | [uint32](#uint32) |  |  |
| f_opts | [bytes](#bytes) |  |  |






<a name="ttn.lorawan.v3.FSKDataRate"></a>

### FSKDataRate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bit_rate | [uint32](#uint32) |  | Bit rate (bps). |






<a name="ttn.lorawan.v3.GatewayAntennaIdentifiers"></a>

### GatewayAntennaIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| antenna_index | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.JoinAcceptPayload"></a>

### JoinAcceptPayload



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| encrypted | [bytes](#bytes) |  |  |
| join_nonce | [bytes](#bytes) |  |  |
| net_id | [bytes](#bytes) |  |  |
| dev_addr | [bytes](#bytes) |  |  |
| dl_settings | [DLSettings](#ttn.lorawan.v3.DLSettings) |  |  |
| rx_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  |
| cf_list | [CFList](#ttn.lorawan.v3.CFList) |  |  |






<a name="ttn.lorawan.v3.JoinRequestPayload"></a>

### JoinRequestPayload



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| join_eui | [bytes](#bytes) |  |  |
| dev_eui | [bytes](#bytes) |  |  |
| dev_nonce | [bytes](#bytes) |  |  |






<a name="ttn.lorawan.v3.LoRaDataRate"></a>

### LoRaDataRate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bandwidth | [uint32](#uint32) |  | Bandwidth (Hz). |
| spreading_factor | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.MACCommand"></a>

### MACCommand



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cid | [MACCommandIdentifier](#ttn.lorawan.v3.MACCommandIdentifier) |  |  |
| raw_payload | [bytes](#bytes) |  |  |
| reset_ind | [MACCommand.ResetInd](#ttn.lorawan.v3.MACCommand.ResetInd) |  |  |
| reset_conf | [MACCommand.ResetConf](#ttn.lorawan.v3.MACCommand.ResetConf) |  |  |
| link_check_ans | [MACCommand.LinkCheckAns](#ttn.lorawan.v3.MACCommand.LinkCheckAns) |  |  |
| link_adr_req | [MACCommand.LinkADRReq](#ttn.lorawan.v3.MACCommand.LinkADRReq) |  |  |
| link_adr_ans | [MACCommand.LinkADRAns](#ttn.lorawan.v3.MACCommand.LinkADRAns) |  |  |
| duty_cycle_req | [MACCommand.DutyCycleReq](#ttn.lorawan.v3.MACCommand.DutyCycleReq) |  |  |
| rx_param_setup_req | [MACCommand.RxParamSetupReq](#ttn.lorawan.v3.MACCommand.RxParamSetupReq) |  |  |
| rx_param_setup_ans | [MACCommand.RxParamSetupAns](#ttn.lorawan.v3.MACCommand.RxParamSetupAns) |  |  |
| dev_status_ans | [MACCommand.DevStatusAns](#ttn.lorawan.v3.MACCommand.DevStatusAns) |  |  |
| new_channel_req | [MACCommand.NewChannelReq](#ttn.lorawan.v3.MACCommand.NewChannelReq) |  |  |
| new_channel_ans | [MACCommand.NewChannelAns](#ttn.lorawan.v3.MACCommand.NewChannelAns) |  |  |
| dl_channel_req | [MACCommand.DLChannelReq](#ttn.lorawan.v3.MACCommand.DLChannelReq) |  |  |
| dl_channel_ans | [MACCommand.DLChannelAns](#ttn.lorawan.v3.MACCommand.DLChannelAns) |  |  |
| rx_timing_setup_req | [MACCommand.RxTimingSetupReq](#ttn.lorawan.v3.MACCommand.RxTimingSetupReq) |  |  |
| tx_param_setup_req | [MACCommand.TxParamSetupReq](#ttn.lorawan.v3.MACCommand.TxParamSetupReq) |  |  |
| rekey_ind | [MACCommand.RekeyInd](#ttn.lorawan.v3.MACCommand.RekeyInd) |  |  |
| rekey_conf | [MACCommand.RekeyConf](#ttn.lorawan.v3.MACCommand.RekeyConf) |  |  |
| adr_param_setup_req | [MACCommand.ADRParamSetupReq](#ttn.lorawan.v3.MACCommand.ADRParamSetupReq) |  |  |
| device_time_ans | [MACCommand.DeviceTimeAns](#ttn.lorawan.v3.MACCommand.DeviceTimeAns) |  |  |
| force_rejoin_req | [MACCommand.ForceRejoinReq](#ttn.lorawan.v3.MACCommand.ForceRejoinReq) |  |  |
| rejoin_param_setup_req | [MACCommand.RejoinParamSetupReq](#ttn.lorawan.v3.MACCommand.RejoinParamSetupReq) |  |  |
| rejoin_param_setup_ans | [MACCommand.RejoinParamSetupAns](#ttn.lorawan.v3.MACCommand.RejoinParamSetupAns) |  |  |
| ping_slot_info_req | [MACCommand.PingSlotInfoReq](#ttn.lorawan.v3.MACCommand.PingSlotInfoReq) |  |  |
| ping_slot_channel_req | [MACCommand.PingSlotChannelReq](#ttn.lorawan.v3.MACCommand.PingSlotChannelReq) |  |  |
| ping_slot_channel_ans | [MACCommand.PingSlotChannelAns](#ttn.lorawan.v3.MACCommand.PingSlotChannelAns) |  |  |
| beacon_timing_ans | [MACCommand.BeaconTimingAns](#ttn.lorawan.v3.MACCommand.BeaconTimingAns) |  |  |
| beacon_freq_req | [MACCommand.BeaconFreqReq](#ttn.lorawan.v3.MACCommand.BeaconFreqReq) |  |  |
| beacon_freq_ans | [MACCommand.BeaconFreqAns](#ttn.lorawan.v3.MACCommand.BeaconFreqAns) |  |  |
| device_mode_ind | [MACCommand.DeviceModeInd](#ttn.lorawan.v3.MACCommand.DeviceModeInd) |  |  |
| device_mode_conf | [MACCommand.DeviceModeConf](#ttn.lorawan.v3.MACCommand.DeviceModeConf) |  |  |






<a name="ttn.lorawan.v3.MACCommand.ADRParamSetupReq"></a>

### MACCommand.ADRParamSetupReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| adr_ack_limit_exponent | [ADRAckLimitExponent](#ttn.lorawan.v3.ADRAckLimitExponent) |  | Exponent e that configures the ADR_ACK_LIMIT = 2^e messages. |
| adr_ack_delay_exponent | [ADRAckDelayExponent](#ttn.lorawan.v3.ADRAckDelayExponent) |  | Exponent e that configures the ADR_ACK_DELAY = 2^e messages. |






<a name="ttn.lorawan.v3.MACCommand.BeaconFreqAns"></a>

### MACCommand.BeaconFreqAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency_ack | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACCommand.BeaconFreqReq"></a>

### MACCommand.BeaconFreqReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency | [uint64](#uint64) |  | Frequency of the Class B beacons (Hz). |






<a name="ttn.lorawan.v3.MACCommand.BeaconTimingAns"></a>

### MACCommand.BeaconTimingAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| delay | [uint32](#uint32) |  | (uint16) See LoRaWAN specification. |
| channel_index | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.MACCommand.DLChannelAns"></a>

### MACCommand.DLChannelAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel_index_ack | [bool](#bool) |  |  |
| frequency_ack | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACCommand.DLChannelReq"></a>

### MACCommand.DLChannelReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel_index | [uint32](#uint32) |  |  |
| frequency | [uint64](#uint64) |  | Downlink channel frequency (Hz). |






<a name="ttn.lorawan.v3.MACCommand.DevStatusAns"></a>

### MACCommand.DevStatusAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| battery | [uint32](#uint32) |  | Device battery status. 0 indicates that the device is connected to an external power source. 1..254 indicates a battery level. 255 indicates that the device was not able to measure the battery level. |
| margin | [int32](#int32) |  | SNR of the last downlink (dB; [-32, &#43;31]). |






<a name="ttn.lorawan.v3.MACCommand.DeviceModeConf"></a>

### MACCommand.DeviceModeConf



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| class | [Class](#ttn.lorawan.v3.Class) |  |  |






<a name="ttn.lorawan.v3.MACCommand.DeviceModeInd"></a>

### MACCommand.DeviceModeInd



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| class | [Class](#ttn.lorawan.v3.Class) |  |  |






<a name="ttn.lorawan.v3.MACCommand.DeviceTimeAns"></a>

### MACCommand.DeviceTimeAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="ttn.lorawan.v3.MACCommand.DutyCycleReq"></a>

### MACCommand.DutyCycleReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_duty_cycle | [AggregatedDutyCycle](#ttn.lorawan.v3.AggregatedDutyCycle) |  |  |






<a name="ttn.lorawan.v3.MACCommand.ForceRejoinReq"></a>

### MACCommand.ForceRejoinReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rejoin_type | [RejoinType](#ttn.lorawan.v3.RejoinType) |  |  |
| data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |
| max_retries | [uint32](#uint32) |  |  |
| period_exponent | [RejoinPeriodExponent](#ttn.lorawan.v3.RejoinPeriodExponent) |  | Exponent e that configures the rejoin period = 32 * 2^e &#43; rand(0,32) seconds. |






<a name="ttn.lorawan.v3.MACCommand.LinkADRAns"></a>

### MACCommand.LinkADRAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel_mask_ack | [bool](#bool) |  |  |
| data_rate_index_ack | [bool](#bool) |  |  |
| tx_power_index_ack | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACCommand.LinkADRReq"></a>

### MACCommand.LinkADRReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |
| tx_power_index | [uint32](#uint32) |  |  |
| channel_mask | [bool](#bool) | repeated |  |
| channel_mask_control | [uint32](#uint32) |  |  |
| nb_trans | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.MACCommand.LinkCheckAns"></a>

### MACCommand.LinkCheckAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| margin | [uint32](#uint32) |  | Indicates the link margin in dB of the received LinkCheckReq, relative to the demodulation floor. |
| gateway_count | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.MACCommand.NewChannelAns"></a>

### MACCommand.NewChannelAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency_ack | [bool](#bool) |  |  |
| data_rate_ack | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACCommand.NewChannelReq"></a>

### MACCommand.NewChannelReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel_index | [uint32](#uint32) |  |  |
| frequency | [uint64](#uint64) |  | Channel frequency (Hz). |
| min_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |
| max_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |






<a name="ttn.lorawan.v3.MACCommand.PingSlotChannelAns"></a>

### MACCommand.PingSlotChannelAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency_ack | [bool](#bool) |  |  |
| data_rate_index_ack | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACCommand.PingSlotChannelReq"></a>

### MACCommand.PingSlotChannelReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency | [uint64](#uint64) |  | Ping slot channel frequency (Hz). |
| data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |






<a name="ttn.lorawan.v3.MACCommand.PingSlotInfoReq"></a>

### MACCommand.PingSlotInfoReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| period | [PingSlotPeriod](#ttn.lorawan.v3.PingSlotPeriod) |  |  |






<a name="ttn.lorawan.v3.MACCommand.RejoinParamSetupAns"></a>

### MACCommand.RejoinParamSetupAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_time_exponent_ack | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACCommand.RejoinParamSetupReq"></a>

### MACCommand.RejoinParamSetupReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_count_exponent | [RejoinCountExponent](#ttn.lorawan.v3.RejoinCountExponent) |  | Exponent e that configures the rejoin counter = 2^(e&#43;4) messages. |
| max_time_exponent | [RejoinTimeExponent](#ttn.lorawan.v3.RejoinTimeExponent) |  | Exponent e that configures the rejoin timer = 2^(e&#43;10) seconds. |






<a name="ttn.lorawan.v3.MACCommand.RekeyConf"></a>

### MACCommand.RekeyConf



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  |






<a name="ttn.lorawan.v3.MACCommand.RekeyInd"></a>

### MACCommand.RekeyInd



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  |






<a name="ttn.lorawan.v3.MACCommand.ResetConf"></a>

### MACCommand.ResetConf



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  |






<a name="ttn.lorawan.v3.MACCommand.ResetInd"></a>

### MACCommand.ResetInd



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| minor_version | [Minor](#ttn.lorawan.v3.Minor) |  |  |






<a name="ttn.lorawan.v3.MACCommand.RxParamSetupAns"></a>

### MACCommand.RxParamSetupAns



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rx2_data_rate_index_ack | [bool](#bool) |  |  |
| rx1_data_rate_offset_ack | [bool](#bool) |  |  |
| rx2_frequency_ack | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACCommand.RxParamSetupReq"></a>

### MACCommand.RxParamSetupReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rx2_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  |  |
| rx1_data_rate_offset | [uint32](#uint32) |  |  |
| rx2_frequency | [uint64](#uint64) |  | Rx2 frequency (Hz). |






<a name="ttn.lorawan.v3.MACCommand.RxTimingSetupReq"></a>

### MACCommand.RxTimingSetupReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  |  |






<a name="ttn.lorawan.v3.MACCommand.TxParamSetupReq"></a>

### MACCommand.TxParamSetupReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_eirp_index | [DeviceEIRP](#ttn.lorawan.v3.DeviceEIRP) |  | Indicates the maximum EIRP value in dBm, indexed by the following vector: [ 8 10 12 13 14 16 18 20 21 24 26 27 29 30 33 36 ] |
| uplink_dwell_time | [bool](#bool) |  |  |
| downlink_dwell_time | [bool](#bool) |  |  |






<a name="ttn.lorawan.v3.MACPayload"></a>

### MACPayload



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| f_hdr | [FHDR](#ttn.lorawan.v3.FHDR) |  |  |
| f_port | [uint32](#uint32) |  |  |
| frm_payload | [bytes](#bytes) |  |  |
| decoded_payload | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |






<a name="ttn.lorawan.v3.MHDR"></a>

### MHDR



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| m_type | [MType](#ttn.lorawan.v3.MType) |  |  |
| major | [Major](#ttn.lorawan.v3.Major) |  |  |






<a name="ttn.lorawan.v3.Message"></a>

### Message



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| m_hdr | [MHDR](#ttn.lorawan.v3.MHDR) |  |  |
| mic | [bytes](#bytes) |  |  |
| mac_payload | [MACPayload](#ttn.lorawan.v3.MACPayload) |  |  |
| join_request_payload | [JoinRequestPayload](#ttn.lorawan.v3.JoinRequestPayload) |  |  |
| join_accept_payload | [JoinAcceptPayload](#ttn.lorawan.v3.JoinAcceptPayload) |  |  |
| rejoin_request_payload | [RejoinRequestPayload](#ttn.lorawan.v3.RejoinRequestPayload) |  |  |






<a name="ttn.lorawan.v3.RejoinRequestPayload"></a>

### RejoinRequestPayload



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rejoin_type | [RejoinType](#ttn.lorawan.v3.RejoinType) |  |  |
| net_id | [bytes](#bytes) |  |  |
| join_eui | [bytes](#bytes) |  |  |
| dev_eui | [bytes](#bytes) |  |  |
| rejoin_cnt | [uint32](#uint32) |  | Contains RJCount0 or RJCount1 depending on rejoin_type. |






<a name="ttn.lorawan.v3.TxRequest"></a>

### TxRequest
TxRequest is a request for transmission.
If sent to a roaming partner, this request is used to generate the DLMetadata Object (see Backend Interfaces 1.0, Table 22).
If the gateway has a scheduler, this request is sent to the gateway, in the order of gateway_ids.
Otherwise, the Gateway Server attempts to schedule the request and creates the TxSettings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| class | [Class](#ttn.lorawan.v3.Class) |  |  |
| downlink_paths | [DownlinkPath](#ttn.lorawan.v3.DownlinkPath) | repeated | Downlink paths used to select a gateway for downlink. In class A, the downlink paths are required to only contain uplink tokens. In class B and C, the downlink paths may contain uplink tokens and fixed gateways antenna identifiers. |
| rx1_delay | [RxDelay](#ttn.lorawan.v3.RxDelay) |  | Rx1 delay (Rx2 delay is Rx1 delay &#43; 1 second). |
| rx1_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | LoRaWAN data rate index for Rx1. |
| rx1_frequency | [uint64](#uint64) |  | Frequency (Hz) for Rx1. |
| rx2_data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | LoRaWAN data rate index for Rx2. |
| rx2_frequency | [uint64](#uint64) |  | Frequency (Hz) for Rx2. |
| priority | [TxSchedulePriority](#ttn.lorawan.v3.TxSchedulePriority) |  | Priority for scheduling. Requests with a higher priority are allocated more channel time than messages with a lower priority, in duty-cycle limited regions. A priority of HIGH or higher sets the HiPriorityFlag in the DLMetadata Object. |
| absolute_time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time when the downlink message should be transmitted. This value is only valid for class C downlink; class A downlink uses uplink tokens and class B downlink is scheduled on ping slots. This requires the gateway to have GPS time sychronization. If the absolute time is not set, the first available time will be used that does not conflict or violate regional limitations. |
| advanced | [google.protobuf.Struct](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case |






<a name="ttn.lorawan.v3.TxSettings"></a>

### TxSettings
TxSettings contains the settings for a transmission.
This message is used on both uplink and downlink.
On downlink, this is a scheduled transmission.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data_rate | [DataRate](#ttn.lorawan.v3.DataRate) |  | Data rate. |
| data_rate_index | [DataRateIndex](#ttn.lorawan.v3.DataRateIndex) |  | LoRaWAN data rate index. |
| coding_rate | [string](#string) |  | LoRa coding rate. |
| frequency | [uint64](#uint64) |  | Frequency (Hz). |
| enable_crc | [bool](#bool) |  | Send a CRC in the packet; only on uplink; on downlink, CRC should not be enabled. |
| timestamp | [uint32](#uint32) |  | Timestamp of the gateway concentrator when the uplink message was received, or when the downlink message should be transmitted (microseconds). On downlink, set timestamp to 0 and time to null to use immediate scheduling. |
| time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time of the gateway when the uplink message was received, or when the downlink message should be transmitted. For downlink, this requires the gateway to have GPS time synchronization. |
| downlink | [TxSettings.Downlink](#ttn.lorawan.v3.TxSettings.Downlink) |  | Transmission settings for downlink. |






<a name="ttn.lorawan.v3.TxSettings.Downlink"></a>

### TxSettings.Downlink
Transmission settings for downlink.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| antenna_index | [uint32](#uint32) |  | Index of the antenna on which the uplink was received and/or downlink must be sent. |
| tx_power | [float](#float) |  | Transmission power (dBm). Only on downlink. |
| invert_polarization | [bool](#bool) |  | Invert LoRa polarization; false for LoRaWAN uplink, true for downlink. |






<a name="ttn.lorawan.v3.UplinkToken"></a>

### UplinkToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers) |  |  |
| timestamp | [uint32](#uint32) |  |  |





 


<a name="ttn.lorawan.v3.ADRAckDelayExponent"></a>

### ADRAckDelayExponent


| Name | Number | Description |
| ---- | ------ | ----------- |
| ADR_ACK_DELAY_1 | 0 |  |
| ADR_ACK_DELAY_2 | 1 |  |
| ADR_ACK_DELAY_4 | 2 |  |
| ADR_ACK_DELAY_8 | 3 |  |
| ADR_ACK_DELAY_16 | 4 |  |
| ADR_ACK_DELAY_32 | 5 |  |
| ADR_ACK_DELAY_64 | 6 |  |
| ADR_ACK_DELAY_128 | 7 |  |
| ADR_ACK_DELAY_256 | 8 |  |
| ADR_ACK_DELAY_512 | 9 |  |
| ADR_ACK_DELAY_1024 | 10 |  |
| ADR_ACK_DELAY_2048 | 11 |  |
| ADR_ACK_DELAY_4096 | 12 |  |
| ADR_ACK_DELAY_8192 | 13 |  |
| ADR_ACK_DELAY_16384 | 14 |  |
| ADR_ACK_DELAY_32768 | 15 |  |



<a name="ttn.lorawan.v3.ADRAckLimitExponent"></a>

### ADRAckLimitExponent


| Name | Number | Description |
| ---- | ------ | ----------- |
| ADR_ACK_LIMIT_1 | 0 |  |
| ADR_ACK_LIMIT_2 | 1 |  |
| ADR_ACK_LIMIT_4 | 2 |  |
| ADR_ACK_LIMIT_8 | 3 |  |
| ADR_ACK_LIMIT_16 | 4 |  |
| ADR_ACK_LIMIT_32 | 5 |  |
| ADR_ACK_LIMIT_64 | 6 |  |
| ADR_ACK_LIMIT_128 | 7 |  |
| ADR_ACK_LIMIT_256 | 8 |  |
| ADR_ACK_LIMIT_512 | 9 |  |
| ADR_ACK_LIMIT_1024 | 10 |  |
| ADR_ACK_LIMIT_2048 | 11 |  |
| ADR_ACK_LIMIT_4096 | 12 |  |
| ADR_ACK_LIMIT_8192 | 13 |  |
| ADR_ACK_LIMIT_16384 | 14 |  |
| ADR_ACK_LIMIT_32768 | 15 |  |



<a name="ttn.lorawan.v3.AggregatedDutyCycle"></a>

### AggregatedDutyCycle


| Name | Number | Description |
| ---- | ------ | ----------- |
| DUTY_CYCLE_1 | 0 | 100%. |
| DUTY_CYCLE_2 | 1 | 50%. |
| DUTY_CYCLE_4 | 2 | 25%. |
| DUTY_CYCLE_8 | 3 | 12.5%. |
| DUTY_CYCLE_16 | 4 | 6.25%. |
| DUTY_CYCLE_32 | 5 | 3.125%. |
| DUTY_CYCLE_64 | 6 | 1.5625%. |
| DUTY_CYCLE_128 | 7 | Roughly 0.781%. |
| DUTY_CYCLE_256 | 8 | Roughly 0.390%. |
| DUTY_CYCLE_512 | 9 | Roughly 0.195%. |
| DUTY_CYCLE_1024 | 10 | Roughly 0.098%. |
| DUTY_CYCLE_2048 | 11 | Roughly 0.049%. |
| DUTY_CYCLE_4096 | 12 | Roughly 0.024%. |
| DUTY_CYCLE_8192 | 13 | Roughly 0.012%. |
| DUTY_CYCLE_16384 | 14 | Roughly 0.006%. |
| DUTY_CYCLE_32768 | 15 | Roughly 0.003%. |



<a name="ttn.lorawan.v3.CFListType"></a>

### CFListType


| Name | Number | Description |
| ---- | ------ | ----------- |
| FREQUENCIES | 0 |  |
| CHANNEL_MASKS | 1 |  |



<a name="ttn.lorawan.v3.Class"></a>

### Class


| Name | Number | Description |
| ---- | ------ | ----------- |
| CLASS_A | 0 |  |
| CLASS_B | 1 |  |
| CLASS_C | 2 |  |



<a name="ttn.lorawan.v3.DataRateIndex"></a>

### DataRateIndex


| Name | Number | Description |
| ---- | ------ | ----------- |
| DATA_RATE_0 | 0 |  |
| DATA_RATE_1 | 1 |  |
| DATA_RATE_2 | 2 |  |
| DATA_RATE_3 | 3 |  |
| DATA_RATE_4 | 4 |  |
| DATA_RATE_5 | 5 |  |
| DATA_RATE_6 | 6 |  |
| DATA_RATE_7 | 7 |  |
| DATA_RATE_8 | 8 |  |
| DATA_RATE_9 | 9 |  |
| DATA_RATE_10 | 10 |  |
| DATA_RATE_11 | 11 |  |
| DATA_RATE_12 | 12 |  |
| DATA_RATE_13 | 13 |  |
| DATA_RATE_14 | 14 |  |
| DATA_RATE_15 | 15 |  |



<a name="ttn.lorawan.v3.DeviceEIRP"></a>

### DeviceEIRP


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEVICE_EIRP_8 | 0 | 8 dBm. |
| DEVICE_EIRP_10 | 1 | 10 dBm. |
| DEVICE_EIRP_12 | 2 | 12 dBm. |
| DEVICE_EIRP_13 | 3 | 13 dBm. |
| DEVICE_EIRP_14 | 4 | 14 dBm. |
| DEVICE_EIRP_16 | 5 | 16 dBm. |
| DEVICE_EIRP_18 | 6 | 18 dBm. |
| DEVICE_EIRP_20 | 7 | 20 dBm. |
| DEVICE_EIRP_21 | 8 | 21 dBm. |
| DEVICE_EIRP_24 | 9 | 24 dBm. |
| DEVICE_EIRP_26 | 10 | 26 dBm. |
| DEVICE_EIRP_27 | 11 | 27 dBm. |
| DEVICE_EIRP_29 | 12 | 29 dBm. |
| DEVICE_EIRP_30 | 13 | 30 dBm. |
| DEVICE_EIRP_33 | 14 | 33 dBm. |
| DEVICE_EIRP_36 | 15 | 36 dBm. |



<a name="ttn.lorawan.v3.MACCommandIdentifier"></a>

### MACCommandIdentifier


| Name | Number | Description |
| ---- | ------ | ----------- |
| CID_RFU_0 | 0 |  |
| CID_RESET | 1 |  |
| CID_LINK_CHECK | 2 |  |
| CID_LINK_ADR | 3 |  |
| CID_DUTY_CYCLE | 4 |  |
| CID_RX_PARAM_SETUP | 5 |  |
| CID_DEV_STATUS | 6 |  |
| CID_NEW_CHANNEL | 7 |  |
| CID_RX_TIMING_SETUP | 8 |  |
| CID_TX_PARAM_SETUP | 9 |  |
| CID_DL_CHANNEL | 10 |  |
| CID_REKEY | 11 |  |
| CID_ADR_PARAM_SETUP | 12 |  |
| CID_DEVICE_TIME | 13 |  |
| CID_FORCE_REJOIN | 14 |  |
| CID_REJOIN_PARAM_SETUP | 15 |  |
| CID_PING_SLOT_INFO | 16 |  |
| CID_PING_SLOT_CHANNEL | 17 |  |
| CID_BEACON_TIMING | 18 | Deprecated |
| CID_BEACON_FREQ | 19 |  |
| CID_DEVICE_MODE | 32 |  |



<a name="ttn.lorawan.v3.MACVersion"></a>

### MACVersion


| Name | Number | Description |
| ---- | ------ | ----------- |
| MAC_UNKNOWN | 0 |  |
| MAC_V1_0 | 1 |  |
| MAC_V1_0_1 | 2 |  |
| MAC_V1_0_2 | 3 |  |
| MAC_V1_1 | 4 |  |
| MAC_V1_0_3 | 5 |  |



<a name="ttn.lorawan.v3.MType"></a>

### MType


| Name | Number | Description |
| ---- | ------ | ----------- |
| JOIN_REQUEST | 0 |  |
| JOIN_ACCEPT | 1 |  |
| UNCONFIRMED_UP | 2 |  |
| UNCONFIRMED_DOWN | 3 |  |
| CONFIRMED_UP | 4 |  |
| CONFIRMED_DOWN | 5 |  |
| REJOIN_REQUEST | 6 |  |
| PROPRIETARY | 7 |  |



<a name="ttn.lorawan.v3.Major"></a>

### Major


| Name | Number | Description |
| ---- | ------ | ----------- |
| LORAWAN_R1 | 0 |  |



<a name="ttn.lorawan.v3.Minor"></a>

### Minor


| Name | Number | Description |
| ---- | ------ | ----------- |
| MINOR_RFU_0 | 0 |  |
| MINOR_1 | 1 |  |
| MINOR_RFU_2 | 2 |  |
| MINOR_RFU_3 | 3 |  |
| MINOR_RFU_4 | 4 |  |
| MINOR_RFU_5 | 5 |  |
| MINOR_RFU_6 | 6 |  |
| MINOR_RFU_7 | 7 |  |
| MINOR_RFU_8 | 8 |  |
| MINOR_RFU_9 | 9 |  |
| MINOR_RFU_10 | 10 |  |
| MINOR_RFU_11 | 11 |  |
| MINOR_RFU_12 | 12 |  |
| MINOR_RFU_13 | 13 |  |
| MINOR_RFU_14 | 14 |  |
| MINOR_RFU_15 | 15 |  |



<a name="ttn.lorawan.v3.PHYVersion"></a>

### PHYVersion


| Name | Number | Description |
| ---- | ------ | ----------- |
| PHY_UNKNOWN | 0 |  |
| PHY_V1_0 | 1 |  |
| PHY_V1_0_1 | 2 |  |
| PHY_V1_0_2_REV_A | 3 |  |
| PHY_V1_0_2_REV_B | 4 |  |
| PHY_V1_1_REV_A | 5 |  |
| PHY_V1_1_REV_B | 6 |  |
| PHY_V1_0_3_REV_A | 7 |  |



<a name="ttn.lorawan.v3.PingSlotPeriod"></a>

### PingSlotPeriod


| Name | Number | Description |
| ---- | ------ | ----------- |
| PING_EVERY_1S | 0 | Every second. |
| PING_EVERY_2S | 1 | Every 2 seconds. |
| PING_EVERY_4S | 2 | Every 4 seconds. |
| PING_EVERY_8S | 3 | Every 8 seconds. |
| PING_EVERY_16S | 4 | Every 16 seconds. |
| PING_EVERY_32S | 5 | Every 32 seconds. |
| PING_EVERY_64S | 6 | Every 64 seconds. |
| PING_EVERY_128S | 7 | Every 128 seconds. |



<a name="ttn.lorawan.v3.RejoinCountExponent"></a>

### RejoinCountExponent


| Name | Number | Description |
| ---- | ------ | ----------- |
| REJOIN_COUNT_16 | 0 |  |
| REJOIN_COUNT_32 | 1 |  |
| REJOIN_COUNT_64 | 2 |  |
| REJOIN_COUNT_128 | 3 |  |
| REJOIN_COUNT_256 | 4 |  |
| REJOIN_COUNT_512 | 5 |  |
| REJOIN_COUNT_1024 | 6 |  |
| REJOIN_COUNT_2048 | 7 |  |
| REJOIN_COUNT_4096 | 8 |  |
| REJOIN_COUNT_8192 | 9 |  |
| REJOIN_COUNT_16384 | 10 |  |
| REJOIN_COUNT_32768 | 11 |  |
| REJOIN_COUNT_65536 | 12 |  |
| REJOIN_COUNT_131072 | 13 |  |
| REJOIN_COUNT_262144 | 14 |  |
| REJOIN_COUNT_524288 | 15 |  |



<a name="ttn.lorawan.v3.RejoinPeriodExponent"></a>

### RejoinPeriodExponent


| Name | Number | Description |
| ---- | ------ | ----------- |
| REJOIN_PERIOD_0 | 0 | Every 32 to 64 seconds. |
| REJOIN_PERIOD_1 | 1 | Every 64 to 96 seconds. |
| REJOIN_PERIOD_2 | 2 | Every 128 to 160 seconds. |
| REJOIN_PERIOD_3 | 3 | Every 256 to 288 seconds. |
| REJOIN_PERIOD_4 | 4 | Every 512 to 544 seconds. |
| REJOIN_PERIOD_5 | 5 | Every 1024 to 1056 seconds. |
| REJOIN_PERIOD_6 | 6 | Every 2048 to 2080 seconds. |
| REJOIN_PERIOD_7 | 7 | Every 4096 to 4128 seconds. |



<a name="ttn.lorawan.v3.RejoinTimeExponent"></a>

### RejoinTimeExponent


| Name | Number | Description |
| ---- | ------ | ----------- |
| REJOIN_TIME_0 | 0 | Every ~17.1 minutes. |
| REJOIN_TIME_1 | 1 | Every ~34.1 minutes. |
| REJOIN_TIME_2 | 2 | Every ~1.1 hours. |
| REJOIN_TIME_3 | 3 | Every ~2.3 hours. |
| REJOIN_TIME_4 | 4 | Every ~4.6 hours. |
| REJOIN_TIME_5 | 5 | Every ~9.1 hours. |
| REJOIN_TIME_6 | 6 | Every ~18.2 hours. |
| REJOIN_TIME_7 | 7 | Every ~1.5 days. |
| REJOIN_TIME_8 | 8 | Every ~3.0 days. |
| REJOIN_TIME_9 | 9 | Every ~6.1 days. |
| REJOIN_TIME_10 | 10 | Every ~12.1 days. |
| REJOIN_TIME_11 | 11 | Every ~3.5 weeks. |
| REJOIN_TIME_12 | 12 | Every ~1.6 months. |
| REJOIN_TIME_13 | 13 | Every ~3.2 months. |
| REJOIN_TIME_14 | 14 | Every ~6.4 months. |
| REJOIN_TIME_15 | 15 | Every ~1.1 year. |



<a name="ttn.lorawan.v3.RejoinType"></a>

### RejoinType


| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTEXT | 0 | Resets DevAddr, Session Keys, Frame Counters, Radio Parameters. |
| SESSION | 1 | Equivalent to the initial JoinRequest. |
| KEYS | 2 | Resets DevAddr, Session Keys, Frame Counters, while keeping the Radio Parameters. |



<a name="ttn.lorawan.v3.RxDelay"></a>

### RxDelay


| Name | Number | Description |
| ---- | ------ | ----------- |
| RX_DELAY_0 | 0 | 1 second. |
| RX_DELAY_1 | 1 | 1 second. |
| RX_DELAY_2 | 2 | 2 seconds. |
| RX_DELAY_3 | 3 | 3 seconds. |
| RX_DELAY_4 | 4 | 4 seconds. |
| RX_DELAY_5 | 5 | 5 seconds. |
| RX_DELAY_6 | 6 | 6 seconds. |
| RX_DELAY_7 | 7 | 7 seconds. |
| RX_DELAY_8 | 8 | 8 seconds. |
| RX_DELAY_9 | 9 | 9 seconds. |
| RX_DELAY_10 | 10 | 10 seconds. |
| RX_DELAY_11 | 11 | 11 seconds. |
| RX_DELAY_12 | 12 | 12 seconds. |
| RX_DELAY_13 | 13 | 13 seconds. |
| RX_DELAY_14 | 14 | 14 seconds. |
| RX_DELAY_15 | 15 | 15 seconds. |



<a name="ttn.lorawan.v3.TxSchedulePriority"></a>

### TxSchedulePriority


| Name | Number | Description |
| ---- | ------ | ----------- |
| LOWEST | 0 |  |
| LOW | 1 |  |
| BELOW_NORMAL | 2 |  |
| NORMAL | 3 |  |
| ABOVE_NORMAL | 4 |  |
| HIGH | 5 |  |
| HIGHEST | 6 |  |


 

 

 



<a name="lorawan-stack/api/message_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/message_services.proto



<a name="ttn.lorawan.v3.ProcessDownlinkMessageRequest"></a>

### ProcessDownlinkMessageRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| end_device_version_ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  |  |
| message | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| parameter | [string](#string) |  |  |






<a name="ttn.lorawan.v3.ProcessUplinkMessageRequest"></a>

### ProcessUplinkMessageRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| end_device_version_ids | [EndDeviceVersionIdentifiers](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  |  |
| message | [ApplicationUplink](#ttn.lorawan.v3.ApplicationUplink) |  |  |
| parameter | [string](#string) |  |  |





 

 

 


<a name="ttn.lorawan.v3.DownlinkMessageProcessor"></a>

### DownlinkMessageProcessor
The DownlinkMessageProcessor service processes downlink messages.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Process | [ProcessDownlinkMessageRequest](#ttn.lorawan.v3.ProcessDownlinkMessageRequest) | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |


<a name="ttn.lorawan.v3.UplinkMessageProcessor"></a>

### UplinkMessageProcessor
The UplinkMessageProcessor service processes uplink messages.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Process | [ProcessUplinkMessageRequest](#ttn.lorawan.v3.ProcessUplinkMessageRequest) | [ApplicationUplink](#ttn.lorawan.v3.ApplicationUplink) |  |

 



<a name="lorawan-stack/api/messages.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/messages.proto



<a name="ttn.lorawan.v3.ApplicationDownlink"></a>

### ApplicationDownlink



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys used by this downlink. |
| f_port | [uint32](#uint32) |  |  |
| f_cnt | [uint32](#uint32) |  |  |
| frm_payload | [bytes](#bytes) |  |  |
| decoded_payload | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |
| confirmed | [bool](#bool) |  |  |
| class_b_c | [ApplicationDownlink.ClassBC](#ttn.lorawan.v3.ApplicationDownlink.ClassBC) |  | Optional gateway and timing information for class B and C. If set, this downlink message will only be transmitted as class B or C downlink. If not set, this downlink message may be transmitted in class A, B and C. |
| priority | [TxSchedulePriority](#ttn.lorawan.v3.TxSchedulePriority) |  | Priority for scheduling the downlink message. |
| correlation_ids | [string](#string) | repeated |  |






<a name="ttn.lorawan.v3.ApplicationDownlink.ClassBC"></a>

### ApplicationDownlink.ClassBC



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateways | [GatewayAntennaIdentifiers](#ttn.lorawan.v3.GatewayAntennaIdentifiers) | repeated | Possible gateway identifiers and antenna index to use for this downlink message. The Network Server selects one of these gateways for downlink, based on connectivity, signal quality, channel utilization and an available slot. If none of the gateways can be selected, the downlink message fails. If empty, a gateway and antenna is selected automatically from the gateways seen in recent uplinks. |
| absolute_time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Absolute time when the downlink message should be transmitted. This requires the gateway to have GPS time synchronization. If the time is in the past or if there is a scheduling conflict, the downlink message fails. If null, the time is selected based on slot availability. This is recommended in class B mode. |






<a name="ttn.lorawan.v3.ApplicationDownlinkFailed"></a>

### ApplicationDownlinkFailed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| downlink | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| error | [ErrorDetails](#ttn.lorawan.v3.ErrorDetails) |  |  |






<a name="ttn.lorawan.v3.ApplicationDownlinks"></a>

### ApplicationDownlinks



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  |






<a name="ttn.lorawan.v3.ApplicationInvalidatedDownlinks"></a>

### ApplicationInvalidatedDownlinks



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  |
| last_f_cnt_down | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.ApplicationJoinAccept"></a>

### ApplicationJoinAccept



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys negotiated in this join. |
| app_s_key | [KeyEnvelope](#ttn.lorawan.v3.KeyEnvelope) |  | Encrypted Application Session Key (if Join Server sent it to Network Server). |
| invalidated_downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated | Downlink messages in the queue that got invalidated because of the session change. |
| pending_session | [bool](#bool) |  | Indicates whether the security context refers to the pending session, i.e. when this join-accept is an answer to a rejoin-request. |






<a name="ttn.lorawan.v3.ApplicationLocation"></a>

### ApplicationLocation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service | [string](#string) |  |  |
| location | [Location](#ttn.lorawan.v3.Location) |  |  |
| attributes | [ApplicationLocation.AttributesEntry](#ttn.lorawan.v3.ApplicationLocation.AttributesEntry) | repeated |  |






<a name="ttn.lorawan.v3.ApplicationLocation.AttributesEntry"></a>

### ApplicationLocation.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.ApplicationUp"></a>

### ApplicationUp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| correlation_ids | [string](#string) | repeated |  |
| uplink_message | [ApplicationUplink](#ttn.lorawan.v3.ApplicationUplink) |  |  |
| join_accept | [ApplicationJoinAccept](#ttn.lorawan.v3.ApplicationJoinAccept) |  |  |
| downlink_ack | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| downlink_nack | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| downlink_sent | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| downlink_failed | [ApplicationDownlinkFailed](#ttn.lorawan.v3.ApplicationDownlinkFailed) |  |  |
| downlink_queued | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| downlink_queue_invalidated | [ApplicationInvalidatedDownlinks](#ttn.lorawan.v3.ApplicationInvalidatedDownlinks) |  |  |
| location_solved | [ApplicationLocation](#ttn.lorawan.v3.ApplicationLocation) |  |  |






<a name="ttn.lorawan.v3.ApplicationUplink"></a>

### ApplicationUplink



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_key_id | [bytes](#bytes) |  | Join Server issued identifier for the session keys used by this uplink. |
| f_port | [uint32](#uint32) |  |  |
| f_cnt | [uint32](#uint32) |  |  |
| frm_payload | [bytes](#bytes) |  |  |
| decoded_payload | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |
| rx_metadata | [RxMetadata](#ttn.lorawan.v3.RxMetadata) | repeated |  |
| settings | [TxSettings](#ttn.lorawan.v3.TxSettings) |  |  |






<a name="ttn.lorawan.v3.DownlinkMessage"></a>

### DownlinkMessage
Downlink message from the network to the end device


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw_payload | [bytes](#bytes) |  |  |
| payload | [Message](#ttn.lorawan.v3.Message) |  |  |
| end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| request | [TxRequest](#ttn.lorawan.v3.TxRequest) |  |  |
| scheduled | [TxSettings](#ttn.lorawan.v3.TxSettings) |  |  |
| correlation_ids | [string](#string) | repeated |  |






<a name="ttn.lorawan.v3.DownlinkQueueRequest"></a>

### DownlinkQueueRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| end_device_ids | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| downlinks | [ApplicationDownlink](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  |






<a name="ttn.lorawan.v3.MessagePayloadFormatters"></a>

### MessagePayloadFormatters



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| up_formatter | [PayloadFormatter](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter for uplink messages, must be set together with its parameter. |
| up_formatter_parameter | [string](#string) |  | Parameter for the up_formatter, must be set together. |
| down_formatter | [PayloadFormatter](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter for downlink messages, must be set together with its parameter. |
| down_formatter_parameter | [string](#string) |  | Parameter for the down_formatter, must be set together. |






<a name="ttn.lorawan.v3.TxAcknowledgment"></a>

### TxAcknowledgment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| correlation_ids | [string](#string) | repeated |  |
| result | [TxAcknowledgment.Result](#ttn.lorawan.v3.TxAcknowledgment.Result) |  |  |






<a name="ttn.lorawan.v3.UplinkMessage"></a>

### UplinkMessage
Uplink message from the end device to the network


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw_payload | [bytes](#bytes) |  |  |
| payload | [Message](#ttn.lorawan.v3.Message) |  |  |
| settings | [TxSettings](#ttn.lorawan.v3.TxSettings) |  |  |
| rx_metadata | [RxMetadata](#ttn.lorawan.v3.RxMetadata) | repeated |  |
| received_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Server time when a component received the message. The Gateway Server, Network Server and Application Server may set this value to their local server time of reception. |
| correlation_ids | [string](#string) | repeated |  |
| gateway_channel_index | [uint32](#uint32) |  | Index of the gateway channel that received the message. Set by Gateway Server. |
| device_channel_index | [uint32](#uint32) |  | Index of the device channel that received the message. Set by Network Server. |





 


<a name="ttn.lorawan.v3.PayloadFormatter"></a>

### PayloadFormatter


| Name | Number | Description |
| ---- | ------ | ----------- |
| FORMATTER_NONE | 0 | No payload formatter to work with raw payload only. |
| FORMATTER_REPOSITORY | 1 | Use payload formatter for the end device type from a repository. |
| FORMATTER_GRPC_SERVICE | 2 | gRPC service payload formatter. The parameter is the host:port of the service. |
| FORMATTER_JAVASCRIPT | 3 | Custom payload formatter that executes Javascript code. The parameter is a JavaScript filename. |
| FORMATTER_CAYENNELPP | 4 | CayenneLPP payload formatter.

More payload formatters can be added. |



<a name="ttn.lorawan.v3.TxAcknowledgment.Result"></a>

### TxAcknowledgment.Result


| Name | Number | Description |
| ---- | ------ | ----------- |
| SUCCESS | 0 |  |
| UNKNOWN_ERROR | 1 |  |
| TOO_LATE | 2 |  |
| TOO_EARLY | 3 |  |
| COLLISION_PACKET | 4 |  |
| COLLISION_BEACON | 5 |  |
| TX_FREQ | 6 |  |
| TX_POWER | 7 |  |
| GPS_UNLOCKED | 8 |  |


 

 

 



<a name="lorawan-stack/api/metadata.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/metadata.proto



<a name="ttn.lorawan.v3.Location"></a>

### Location



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| latitude | [double](#double) |  | The North–South position (degrees; -90 to &#43;90), where 0 is the equator, North pole is positive, South pole is negative. |
| longitude | [double](#double) |  | The East-West position (degrees; -180 to &#43;180), where 0 is the Prime Meridian (Greenwich), East is positive , West is negative. |
| altitude | [int32](#int32) |  | The altitude (meters), where 0 is the mean sea level. |
| accuracy | [int32](#int32) |  | The accuracy of the location (meters). |
| source | [LocationSource](#ttn.lorawan.v3.LocationSource) |  | Source of the location information. |






<a name="ttn.lorawan.v3.RxMetadata"></a>

### RxMetadata
Contains metadata for a received message. Each antenna that receives
a message corresponds to one RxMetadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| gateway_ids | [GatewayIdentifiers](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| antenna_index | [uint32](#uint32) |  |  |
| time | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| timestamp | [uint32](#uint32) |  | Gateway concentrator timestamp when the Rx finished (microseconds). |
| fine_timestamp | [uint64](#uint64) |  | Gateway&#39;s internal fine timestamp when the Rx finished (nanoseconds). |
| encrypted_fine_timestamp | [bytes](#bytes) |  | Encrypted gateway&#39;s internal fine timestamp when the Rx finished (nanoseconds). |
| encrypted_fine_timestamp_key_id | [string](#string) |  |  |
| rssi | [float](#float) |  | Received signal strength (dBm). |
| channel_rssi | [float](#float) |  | Received channel power (dBm). |
| rssi_standard_deviation | [float](#float) |  | Standard deviation of the RSSI. |
| snr | [float](#float) |  | Signal-to-noise ratio (dB). |
| frequency_offset | [int64](#int64) |  | Frequency offset (Hz). |
| location | [Location](#ttn.lorawan.v3.Location) |  | Antenna location; injected by the Gateway Server. |
| downlink_path_constraint | [DownlinkPathConstraint](#ttn.lorawan.v3.DownlinkPathConstraint) |  | Gateway downlink path constraint; injected by the Gateway Server. |
| uplink_token | [bytes](#bytes) |  | Uplink token to be included in the Tx request in class A downlink; injected by gateway, Gateway Server or fNS. |
| advanced | [google.protobuf.Struct](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case |





 


<a name="ttn.lorawan.v3.LocationSource"></a>

### LocationSource


| Name | Number | Description |
| ---- | ------ | ----------- |
| SOURCE_UNKNOWN | 0 | The source of the location is not known or not set. |
| SOURCE_GPS | 1 | The location is determined by GPS. |
| SOURCE_REGISTRY | 3 | The location is set in and updated from a registry. |
| SOURCE_IP_GEOLOCATION | 4 | The location is estimated with IP geolocation. |
| SOURCE_WIFI_RSSI_GEOLOCATION | 5 | The location is estimated with WiFi RSSI geolocation. |
| SOURCE_BT_RSSI_GEOLOCATION | 6 | The location is estimated with BT/BLE RSSI geolocation. |
| SOURCE_LORA_RSSI_GEOLOCATION | 7 | The location is estimated with LoRa RSSI geolocation. |
| SOURCE_LORA_TDOA_GEOLOCATION | 8 | The location is estimated with LoRa TDOA geolocation. |
| SOURCE_COMBINED_GEOLOCATION | 9 | The location is estimated by a combination of geolocation sources.

More estimation methods can be added. |


 

 

 



<a name="lorawan-stack/api/networkserver.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/networkserver.proto


 

 

 


<a name="ttn.lorawan.v3.AsNs"></a>

### AsNs
The AsNs service connects an Application Server to a Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| LinkApplication | [.google.protobuf.Empty](#google.protobuf.Empty) stream | [ApplicationUp](#ttn.lorawan.v3.ApplicationUp) stream |  |
| DownlinkQueueReplace | [DownlinkQueueRequest](#ttn.lorawan.v3.DownlinkQueueRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| DownlinkQueuePush | [DownlinkQueueRequest](#ttn.lorawan.v3.DownlinkQueueRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| DownlinkQueueList | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | [ApplicationDownlinks](#ttn.lorawan.v3.ApplicationDownlinks) |  |


<a name="ttn.lorawan.v3.GsNs"></a>

### GsNs
The GsNs service connects a Gateway Server to a Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| HandleUplink | [UplinkMessage](#ttn.lorawan.v3.UplinkMessage) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |


<a name="ttn.lorawan.v3.NsEndDeviceRegistry"></a>

### NsEndDeviceRegistry
The NsEndDeviceRegistry service allows clients to manage their end devices on the Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Get | [GetEndDeviceRequest](#ttn.lorawan.v3.GetEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Get returns the device that matches the given identifiers. If there are multiple matches, an error will be returned. |
| Set | [SetEndDeviceRequest](#ttn.lorawan.v3.SetEndDeviceRequest) | [EndDevice](#ttn.lorawan.v3.EndDevice) | Set creates or updates the device. |
| Delete | [EndDeviceIdentifiers](#ttn.lorawan.v3.EndDeviceIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) | Delete deletes the device that matches the given identifiers. If there are multiple matches, an error will be returned. |

 



<a name="lorawan-stack/api/oauth.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/oauth.proto



<a name="ttn.lorawan.v3.ListOAuthAccessTokensRequest"></a>

### ListOAuthAccessTokensRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest"></a>

### ListOAuthClientAuthorizationsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.OAuthAccessToken"></a>

### OAuthAccessToken



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| id | [string](#string) |  |  |
| access_token | [string](#string) |  |  |
| refresh_token | [string](#string) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="ttn.lorawan.v3.OAuthAccessTokenIdentifiers"></a>

### OAuthAccessTokenIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| id | [string](#string) |  |  |






<a name="ttn.lorawan.v3.OAuthAccessTokens"></a>

### OAuthAccessTokens



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokens | [OAuthAccessToken](#ttn.lorawan.v3.OAuthAccessToken) | repeated |  |






<a name="ttn.lorawan.v3.OAuthAuthorizationCode"></a>

### OAuthAuthorizationCode



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |
| code | [string](#string) |  |  |
| redirect_uri | [string](#string) |  |  |
| state | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="ttn.lorawan.v3.OAuthClientAuthorization"></a>

### OAuthClientAuthorization



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers"></a>

### OAuthClientAuthorizationIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| client_ids | [ClientIdentifiers](#ttn.lorawan.v3.ClientIdentifiers) |  |  |






<a name="ttn.lorawan.v3.OAuthClientAuthorizations"></a>

### OAuthClientAuthorizations



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authorizations | [OAuthClientAuthorization](#ttn.lorawan.v3.OAuthClientAuthorization) | repeated |  |





 

 

 

 



<a name="lorawan-stack/api/oauth_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/oauth_services.proto


 

 

 


<a name="ttn.lorawan.v3.OAuthAuthorizationRegistry"></a>

### OAuthAuthorizationRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| List | [ListOAuthClientAuthorizationsRequest](#ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest) | [OAuthClientAuthorizations](#ttn.lorawan.v3.OAuthClientAuthorizations) |  |
| ListTokens | [ListOAuthAccessTokensRequest](#ttn.lorawan.v3.ListOAuthAccessTokensRequest) | [OAuthAccessTokens](#ttn.lorawan.v3.OAuthAccessTokens) |  |
| Delete | [OAuthClientAuthorizationIdentifiers](#ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| DeleteToken | [OAuthAccessTokenIdentifiers](#ttn.lorawan.v3.OAuthAccessTokenIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



<a name="lorawan-stack/api/organization.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/organization.proto



<a name="ttn.lorawan.v3.CreateOrganizationAPIKeyRequest"></a>

### CreateOrganizationAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| name | [string](#string) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |






<a name="ttn.lorawan.v3.CreateOrganizationRequest"></a>

### CreateOrganizationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization | [Organization](#ttn.lorawan.v3.Organization) |  |  |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created application. NOTE: It is currently not possible to have organizations collaborating on other organizations. |






<a name="ttn.lorawan.v3.GetOrganizationRequest"></a>

### GetOrganizationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.ListOrganizationAPIKeysRequest"></a>

### ListOrganizationAPIKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListOrganizationCollaboratorsRequest"></a>

### ListOrganizationCollaboratorsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListOrganizationsRequest"></a>

### ListOrganizationsRequest
By default we list all organizations the caller has rights on.
Set the user to instead list the organizations
where the user or organization is collaborator on.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| collaborator | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | NOTE: It is currently not possible to have organizations collaborating on other organizations. |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.Organization"></a>

### Organization



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| attributes | [Organization.AttributesEntry](#ttn.lorawan.v3.Organization.AttributesEntry) | repeated |  |
| contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  |






<a name="ttn.lorawan.v3.Organization.AttributesEntry"></a>

### Organization.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.Organizations"></a>

### Organizations



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organizations | [Organization](#ttn.lorawan.v3.Organization) | repeated |  |






<a name="ttn.lorawan.v3.SetOrganizationCollaboratorRequest"></a>

### SetOrganizationCollaboratorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| collaborator | [Collaborator](#ttn.lorawan.v3.Collaborator) |  |  |






<a name="ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest"></a>

### UpdateOrganizationAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization_ids | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  |






<a name="ttn.lorawan.v3.UpdateOrganizationRequest"></a>

### UpdateOrganizationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| organization | [Organization](#ttn.lorawan.v3.Organization) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |





 

 

 

 



<a name="lorawan-stack/api/organization_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/organization_services.proto


 

 

 


<a name="ttn.lorawan.v3.OrganizationAccess"></a>

### OrganizationAccess


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListRights | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) | [Rights](#ttn.lorawan.v3.Rights) |  |
| CreateAPIKey | [CreateOrganizationAPIKeyRequest](#ttn.lorawan.v3.CreateOrganizationAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) |  |
| ListAPIKeys | [ListOrganizationAPIKeysRequest](#ttn.lorawan.v3.ListOrganizationAPIKeysRequest) | [APIKeys](#ttn.lorawan.v3.APIKeys) |  |
| UpdateAPIKey | [UpdateOrganizationAPIKeyRequest](#ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) | Update the rights of an existing organization API key. To generate an API key, the CreateAPIKey should be used. To delete an API key, update it with zero rights. |
| SetCollaborator | [SetOrganizationCollaboratorRequest](#ttn.lorawan.v3.SetOrganizationCollaboratorRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) | Set the rights of a collaborator (member) on the organization. Users are considered to be a collaborator if they have at least one right on the organization. Note that only users can collaborate (be member of) an organization. |
| ListCollaborators | [ListOrganizationCollaboratorsRequest](#ttn.lorawan.v3.ListOrganizationCollaboratorsRequest) | [Collaborators](#ttn.lorawan.v3.Collaborators) |  |


<a name="ttn.lorawan.v3.OrganizationRegistry"></a>

### OrganizationRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateOrganizationRequest](#ttn.lorawan.v3.CreateOrganizationRequest) | [Organization](#ttn.lorawan.v3.Organization) | Create a new organization. This also sets the given user as first collaborator with all possible rights. |
| Get | [GetOrganizationRequest](#ttn.lorawan.v3.GetOrganizationRequest) | [Organization](#ttn.lorawan.v3.Organization) | Get the organization with the given identifiers, selecting the fields given by the field mask. The method may return more or less fields, depending on the rights of the caller. |
| List | [ListOrganizationsRequest](#ttn.lorawan.v3.ListOrganizationsRequest) | [Organizations](#ttn.lorawan.v3.Organizations) | List organizations. See request message for details. |
| Update | [UpdateOrganizationRequest](#ttn.lorawan.v3.UpdateOrganizationRequest) | [Organization](#ttn.lorawan.v3.Organization) |  |
| Delete | [OrganizationIdentifiers](#ttn.lorawan.v3.OrganizationIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



<a name="lorawan-stack/api/regional.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/regional.proto



<a name="ttn.lorawan.v3.ConcentratorConfig"></a>

### ConcentratorConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channels | [ConcentratorConfig.Channel](#ttn.lorawan.v3.ConcentratorConfig.Channel) | repeated |  |
| lora_standard_channel | [ConcentratorConfig.LoRaStandardChannel](#ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel) |  |  |
| fsk_channel | [ConcentratorConfig.FSKChannel](#ttn.lorawan.v3.ConcentratorConfig.FSKChannel) |  |  |
| lbt | [ConcentratorConfig.LBTConfiguration](#ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration) |  |  |
| ping_slot | [ConcentratorConfig.Channel](#ttn.lorawan.v3.ConcentratorConfig.Channel) |  |  |
| radios | [GatewayRadio](#ttn.lorawan.v3.GatewayRadio) | repeated |  |
| clock_source | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.ConcentratorConfig.Channel"></a>

### ConcentratorConfig.Channel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency | [uint64](#uint64) |  | Frequency (Hz). |
| radio | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.ConcentratorConfig.FSKChannel"></a>

### ConcentratorConfig.FSKChannel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency | [uint64](#uint64) |  | Frequency (Hz). |
| radio | [uint32](#uint32) |  |  |






<a name="ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration"></a>

### ConcentratorConfig.LBTConfiguration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rssi_target | [float](#float) |  | Received signal strength (dBm). |
| rssi_offset | [float](#float) |  | Received signal strength offset (dBm). |
| scan_time | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel"></a>

### ConcentratorConfig.LoRaStandardChannel



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency | [uint64](#uint64) |  | Frequency (Hz). |
| radio | [uint32](#uint32) |  |  |
| bandwidth | [uint32](#uint32) |  | Bandwidth (Hz). |
| spreading_factor | [uint32](#uint32) |  |  |





 

 

 

 



<a name="lorawan-stack/api/rights.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/rights.proto



<a name="ttn.lorawan.v3.APIKey"></a>

### APIKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | Immutable and unique public identifier for the API key. Generated by the Access Server. |
| key | [string](#string) |  | Immutable and unique secret value of the API key. Generated by the Access Server. |
| name | [string](#string) |  | User-defined (friendly) name for the API key. |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated | Rights that are granted to this API key. |






<a name="ttn.lorawan.v3.APIKeys"></a>

### APIKeys



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| api_keys | [APIKey](#ttn.lorawan.v3.APIKey) | repeated |  |






<a name="ttn.lorawan.v3.Collaborator"></a>

### Collaborator



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [OrganizationOrUserIdentifiers](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |






<a name="ttn.lorawan.v3.Collaborators"></a>

### Collaborators



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| collaborators | [Collaborator](#ttn.lorawan.v3.Collaborator) | repeated |  |






<a name="ttn.lorawan.v3.Rights"></a>

### Rights



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |





 


<a name="ttn.lorawan.v3.Right"></a>

### Right
Right is the enum that defines all the different rights to do something in the network.

| Name | Number | Description |
| ---- | ------ | ----------- |
| right_invalid | 0 |  |
| RIGHT_USER_INFO | 1 | The right to view user information. |
| RIGHT_USER_SETTINGS_BASIC | 2 | The right to edit basic user settings. |
| RIGHT_USER_SETTINGS_API_KEYS | 3 | The right to view and edit user API keys. |
| RIGHT_USER_DELETE | 4 | The right to delete user account. |
| RIGHT_USER_AUTHORIZED_CLIENTS | 5 | The right to view and edit authorized OAuth clients of the user. |
| RIGHT_USER_APPLICATIONS_LIST | 6 | The right to list applications the user is a collaborator of. |
| RIGHT_USER_APPLICATIONS_CREATE | 7 | The right to create an application under the user account. |
| RIGHT_USER_GATEWAYS_LIST | 8 | The right to list gateways the user is a collaborator of. |
| RIGHT_USER_GATEWAYS_CREATE | 9 | The right to create a gateway under the account of the user. |
| RIGHT_USER_CLIENTS_LIST | 10 | The right to list OAuth clients the user is a collaborator of. |
| RIGHT_USER_CLIENTS_CREATE | 11 | The right to create an OAuth client under the account of the user. |
| RIGHT_USER_ORGANIZATIONS_LIST | 12 | The right to list organizations the user is a member of. |
| RIGHT_USER_ORGANIZATIONS_CREATE | 13 | The right to create an organization under the user account. |
| RIGHT_USER_ALL | 14 | The pseudo-right for all (current and future) user rights. |
| RIGHT_APPLICATION_INFO | 15 | The right to view application information. |
| RIGHT_APPLICATION_SETTINGS_BASIC | 16 | The right to edit basic application settings. |
| RIGHT_APPLICATION_SETTINGS_API_KEYS | 17 | The right to view and edit application API keys. |
| RIGHT_APPLICATION_SETTINGS_COLLABORATORS | 18 | The right to view and edit application collaborators. |
| RIGHT_APPLICATION_DELETE | 19 | The right to delete application. |
| RIGHT_APPLICATION_DEVICES_READ | 20 | The right to view devices in application. |
| RIGHT_APPLICATION_DEVICES_WRITE | 21 | The right to create devices in application. |
| RIGHT_APPLICATION_DEVICES_READ_KEYS | 22 | The right to view device keys in application. Note that keys may not be stored in a way that supports viewing them. |
| RIGHT_APPLICATION_DEVICES_WRITE_KEYS | 23 | The right to edit device keys in application. |
| RIGHT_APPLICATION_TRAFFIC_READ | 24 | The right to read application traffic (uplink and downlink). |
| RIGHT_APPLICATION_TRAFFIC_UP_WRITE | 25 | The right to write uplink application traffic. |
| RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE | 26 | The right to write downlink application traffic. |
| RIGHT_APPLICATION_LINK | 27 | The right to link as Application to a Network Server for traffic exchange, i.e. read uplink and write downlink (API keys only). This right is typically only given to an Application Server. |
| RIGHT_APPLICATION_ALL | 28 | The pseudo-right for all (current and future) application rights. |
| RIGHT_CLIENT_ALL | 29 | The pseudo-right for all (current and future) OAuth client rights. |
| RIGHT_GATEWAY_INFO | 30 | The right to view gateway information. |
| RIGHT_GATEWAY_SETTINGS_BASIC | 31 | The right to edit basic gateway settings. |
| RIGHT_GATEWAY_SETTINGS_API_KEYS | 32 | The right to view and edit gateway API keys. |
| RIGHT_GATEWAY_SETTINGS_COLLABORATORS | 33 | The right to view and edit gateway collaborators. |
| RIGHT_GATEWAY_DELETE | 34 | The right to delete gateway. |
| RIGHT_GATEWAY_TRAFFIC_READ | 35 | The right to read gateway traffic. |
| RIGHT_GATEWAY_TRAFFIC_DOWN_WRITE | 36 | The right to write downlink gateway traffic. |
| RIGHT_GATEWAY_LINK | 37 | The right to link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink (API keys only) |
| RIGHT_GATEWAY_STATUS_READ | 38 | The right to view gateway status. |
| RIGHT_GATEWAY_LOCATION_READ | 39 | The right to view view gateway location. |
| RIGHT_GATEWAY_ALL | 40 | The pseudo-right for all (current and future) gateway rights. |
| RIGHT_ORGANIZATION_INFO | 41 | The right to view organization information. |
| RIGHT_ORGANIZATION_SETTINGS_BASIC | 42 | The right to edit basic organization settings. |
| RIGHT_ORGANIZATION_SETTINGS_API_KEYS | 43 | The right to view and edit organization API keys. |
| RIGHT_ORGANIZATION_SETTINGS_MEMBERS | 44 | The right to view and edit organization members. |
| RIGHT_ORGANIZATION_DELETE | 45 | The right to delete organization. |
| RIGHT_ORGANIZATION_APPLICATIONS_LIST | 46 | The right to list the applications the organization is a collaborator of. |
| RIGHT_ORGANIZATION_APPLICATIONS_CREATE | 47 | The right to create an application under the organization. |
| RIGHT_ORGANIZATION_GATEWAYS_LIST | 48 | The right to list the gateways the organization is a collaborator of. |
| RIGHT_ORGANIZATION_GATEWAYS_CREATE | 49 | The right to create a gateway under the organization. |
| RIGHT_ORGANIZATION_CLIENTS_LIST | 50 | The right to list the OAuth clients the organization is a collaborator of. |
| RIGHT_ORGANIZATION_CLIENTS_CREATE | 51 | The right to create an OAuth client under the organization. |
| RIGHT_ORGANIZATION_ADD_AS_COLLABORATOR | 52 | The right to add the organization as a collaborator on an existing entity. |
| RIGHT_ORGANIZATION_ALL | 53 | The pseudo-right for all (current and future) organization rights. |
| RIGHT_SEND_INVITES | 54 | The right to send invites to new users. Note that this is not prefixed with &#34;USER_&#34;; it is not a right on the user entity. |
| RIGHT_ALL | 55 | The pseudo-right for all (current and future) possible rights. |


 

 

 



<a name="lorawan-stack/api/search_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/search_services.proto



<a name="ttn.lorawan.v3.SearchEndDevicesRequest"></a>

### SearchEndDevicesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_ids | [ApplicationIdentifiers](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| id_contains | [string](#string) |  | Find end devices where the ID contains this substring. |
| name_contains | [string](#string) |  | Find end devices where the name contains this substring. |
| description_contains | [string](#string) |  | Find end devices where the description contains this substring. |
| attributes_contain | [SearchEndDevicesRequest.AttributesContainEntry](#ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry) | repeated | Find end devices where the given attributes contain these substrings. |
| dev_eui_contains | [string](#string) |  | Find end devices where the (hexadecimal) DevEUI contains this substring. |
| join_eui_contains | [string](#string) |  | Find end devices where the (hexadecimal) JoinEUI contains this substring. |
| dev_addr_contains | [string](#string) |  | Find end devices where the (hexadecimal) DevAddr contains this substring. |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry"></a>

### SearchEndDevicesRequest.AttributesContainEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.SearchEntitiesRequest"></a>

### SearchEntitiesRequest
This message is used for finding entities in the EntityRegistrySearch service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id_contains | [string](#string) |  | Find entities where the ID contains this substring. |
| name_contains | [string](#string) |  | Find entities where the name contains this substring. |
| description_contains | [string](#string) |  | Find entities where the description contains this substring. |
| attributes_contain | [SearchEntitiesRequest.AttributesContainEntry](#ttn.lorawan.v3.SearchEntitiesRequest.AttributesContainEntry) | repeated | Find entities where the given attributes contain these substrings. |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.SearchEntitiesRequest.AttributesContainEntry"></a>

### SearchEntitiesRequest.AttributesContainEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="ttn.lorawan.v3.EndDeviceRegistrySearch"></a>

### EndDeviceRegistrySearch
The EndDeviceRegistrySearch service indexes devices in the EndDeviceRegistry
and enables searching for them.
This service is not implemented on all deployments.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SearchEndDevices | [SearchEndDevicesRequest](#ttn.lorawan.v3.SearchEndDevicesRequest) | [EndDevices](#ttn.lorawan.v3.EndDevices) |  |


<a name="ttn.lorawan.v3.EntityRegistrySearch"></a>

### EntityRegistrySearch
The EntityRegistrySearch service indexes entities in the various registries
and enables searching for them.
This service is not implemented on all deployments.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SearchApplications | [SearchEntitiesRequest](#ttn.lorawan.v3.SearchEntitiesRequest) | [Applications](#ttn.lorawan.v3.Applications) |  |
| SearchClients | [SearchEntitiesRequest](#ttn.lorawan.v3.SearchEntitiesRequest) | [Clients](#ttn.lorawan.v3.Clients) |  |
| SearchGateways | [SearchEntitiesRequest](#ttn.lorawan.v3.SearchEntitiesRequest) | [Gateways](#ttn.lorawan.v3.Gateways) |  |
| SearchOrganizations | [SearchEntitiesRequest](#ttn.lorawan.v3.SearchEntitiesRequest) | [Organizations](#ttn.lorawan.v3.Organizations) |  |
| SearchUsers | [SearchEntitiesRequest](#ttn.lorawan.v3.SearchEntitiesRequest) | [Users](#ttn.lorawan.v3.Users) |  |

 



<a name="lorawan-stack/api/user.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/user.proto



<a name="ttn.lorawan.v3.CreateTemporaryPasswordRequest"></a>

### CreateTemporaryPasswordRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |






<a name="ttn.lorawan.v3.CreateUserAPIKeyRequest"></a>

### CreateUserAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| name | [string](#string) |  |  |
| rights | [Right](#ttn.lorawan.v3.Right) | repeated |  |






<a name="ttn.lorawan.v3.CreateUserRequest"></a>

### CreateUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#ttn.lorawan.v3.User) |  |  |
| invitation_token | [string](#string) |  |  |






<a name="ttn.lorawan.v3.DeleteInvitationRequest"></a>

### DeleteInvitationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| email | [string](#string) |  |  |






<a name="ttn.lorawan.v3.GetUserRequest"></a>

### GetUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.Invitation"></a>

### Invitation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| email | [string](#string) |  |  |
| token | [string](#string) |  |  |
| expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| accepted_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| accepted_by | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |






<a name="ttn.lorawan.v3.Invitations"></a>

### Invitations



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| invitations | [Invitation](#ttn.lorawan.v3.Invitation) | repeated |  |






<a name="ttn.lorawan.v3.ListInvitationsRequest"></a>

### ListInvitationsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListUserAPIKeysRequest"></a>

### ListUserAPIKeysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.ListUserSessionsRequest"></a>

### ListUserSessionsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| order | [string](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| limit | [uint32](#uint32) |  | Limit the number of results per page. |
| page | [uint32](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |






<a name="ttn.lorawan.v3.Picture"></a>

### Picture



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| embedded | [Picture.Embedded](#ttn.lorawan.v3.Picture.Embedded) |  | Embedded picture, always maximum 128px in size. Omitted if there are external URLs available (in sizes). |
| sizes | [Picture.SizesEntry](#ttn.lorawan.v3.Picture.SizesEntry) | repeated | URLs of the picture for different sizes, if available on a CDN. |






<a name="ttn.lorawan.v3.Picture.Embedded"></a>

### Picture.Embedded



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mime_type | [string](#string) |  | MIME type of the picture. |
| data | [bytes](#bytes) |  | Picture data. A data URI can be constructed as follows: `data:&lt;mime_type&gt;;base64,&lt;data&gt;`. |






<a name="ttn.lorawan.v3.Picture.SizesEntry"></a>

### Picture.SizesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [uint32](#uint32) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.SendInvitationRequest"></a>

### SendInvitationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| email | [string](#string) |  |  |






<a name="ttn.lorawan.v3.UpdateUserAPIKeyRequest"></a>

### UpdateUserAPIKeyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| api_key | [APIKey](#ttn.lorawan.v3.APIKey) |  |  |






<a name="ttn.lorawan.v3.UpdateUserPasswordRequest"></a>

### UpdateUserPasswordRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| new | [string](#string) |  |  |
| old | [string](#string) |  |  |






<a name="ttn.lorawan.v3.UpdateUserRequest"></a>

### UpdateUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#ttn.lorawan.v3.User) |  |  |
| field_mask | [google.protobuf.FieldMask](#google.protobuf.FieldMask) |  |  |






<a name="ttn.lorawan.v3.User"></a>

### User
User is the message that defines an user on the network.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| attributes | [User.AttributesEntry](#ttn.lorawan.v3.User.AttributesEntry) | repeated |  |
| contact_info | [ContactInfo](#ttn.lorawan.v3.ContactInfo) | repeated |  |
| primary_email_address | [string](#string) |  | Primary email address that can be used for logging in. This address is not public, use contact_info for that. |
| primary_email_address_validated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| password | [string](#string) |  | Only used on create; never returned on API calls. |
| password_updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| require_password_update | [bool](#bool) |  |  |
| state | [State](#ttn.lorawan.v3.State) |  | The reviewing state of the user. This field can only be modified by admins. |
| admin | [bool](#bool) |  | This user is an admin. This field can only be modified by other admins. |
| temporary_password | [string](#string) |  | The temporary password can only be used to update a user&#39;s password; never returned on API calls. |
| temporary_password_created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| temporary_password_expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| profile_picture | [Picture](#ttn.lorawan.v3.Picture) |  |  |






<a name="ttn.lorawan.v3.User.AttributesEntry"></a>

### User.AttributesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="ttn.lorawan.v3.UserSession"></a>

### UserSession



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| session_id | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| expires_at | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="ttn.lorawan.v3.UserSessionIdentifiers"></a>

### UserSessionIdentifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_ids | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| session_id | [string](#string) |  |  |






<a name="ttn.lorawan.v3.UserSessions"></a>

### UserSessions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sessions | [UserSession](#ttn.lorawan.v3.UserSession) | repeated |  |






<a name="ttn.lorawan.v3.Users"></a>

### Users



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [User](#ttn.lorawan.v3.User) | repeated |  |





 

 

 

 



<a name="lorawan-stack/api/user_services.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## lorawan-stack/api/user_services.proto


 

 

 


<a name="ttn.lorawan.v3.UserAccess"></a>

### UserAccess


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListRights | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) | [Rights](#ttn.lorawan.v3.Rights) |  |
| CreateAPIKey | [CreateUserAPIKeyRequest](#ttn.lorawan.v3.CreateUserAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) |  |
| ListAPIKeys | [ListUserAPIKeysRequest](#ttn.lorawan.v3.ListUserAPIKeysRequest) | [APIKeys](#ttn.lorawan.v3.APIKeys) |  |
| UpdateAPIKey | [UpdateUserAPIKeyRequest](#ttn.lorawan.v3.UpdateUserAPIKeyRequest) | [APIKey](#ttn.lorawan.v3.APIKey) | Update the rights of an existing user API key. To generate an API key, the CreateAPIKey should be used. To delete an API key, update it with zero rights. |


<a name="ttn.lorawan.v3.UserInvitationRegistry"></a>

### UserInvitationRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Send | [SendInvitationRequest](#ttn.lorawan.v3.SendInvitationRequest) | [Invitation](#ttn.lorawan.v3.Invitation) |  |
| List | [ListInvitationsRequest](#ttn.lorawan.v3.ListInvitationsRequest) | [Invitations](#ttn.lorawan.v3.Invitations) |  |
| Delete | [DeleteInvitationRequest](#ttn.lorawan.v3.DeleteInvitationRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |


<a name="ttn.lorawan.v3.UserRegistry"></a>

### UserRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateUserRequest](#ttn.lorawan.v3.CreateUserRequest) | [User](#ttn.lorawan.v3.User) | Register a new user. This method may be restricted by network settings. |
| Get | [GetUserRequest](#ttn.lorawan.v3.GetUserRequest) | [User](#ttn.lorawan.v3.User) | Get the user with the given identifiers, selecting the fields given by the field mask. The method may return more or less fields, depending on the rights of the caller. |
| Update | [UpdateUserRequest](#ttn.lorawan.v3.UpdateUserRequest) | [User](#ttn.lorawan.v3.User) |  |
| CreateTemporaryPassword | [CreateTemporaryPasswordRequest](#ttn.lorawan.v3.CreateTemporaryPasswordRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) | Create a temporary password that can be used for updating a forgotten password. The generated password is sent to the user&#39;s email address. |
| UpdatePassword | [UpdateUserPasswordRequest](#ttn.lorawan.v3.UpdateUserPasswordRequest) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |
| Delete | [UserIdentifiers](#ttn.lorawan.v3.UserIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |


<a name="ttn.lorawan.v3.UserSessionRegistry"></a>

### UserSessionRegistry


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| List | [ListUserSessionsRequest](#ttn.lorawan.v3.ListUserSessionsRequest) | [UserSessions](#ttn.lorawan.v3.UserSessions) |  |
| Delete | [UserSessionIdentifiers](#ttn.lorawan.v3.UserSessionIdentifiers) | [.google.protobuf.Empty](#google.protobuf.Empty) |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |
| <a name="double" /> double |  | double | double | float |
| <a name="float" /> float |  | float | float | float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long |
| <a name="bool" /> bool |  | bool | boolean | boolean |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str |

