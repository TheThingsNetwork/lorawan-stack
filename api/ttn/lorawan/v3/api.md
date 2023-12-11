<a name="top"></a>

# API Documentation

## <a name="toc">Table of Contents</a>

- [File `ttn/lorawan/v3/_api.proto`](#ttn/lorawan/v3/_api.proto)
- [File `ttn/lorawan/v3/application.proto`](#ttn/lorawan/v3/application.proto)
  - [Message `Application`](#ttn.lorawan.v3.Application)
  - [Message `Application.AttributesEntry`](#ttn.lorawan.v3.Application.AttributesEntry)
  - [Message `Applications`](#ttn.lorawan.v3.Applications)
  - [Message `CreateApplicationAPIKeyRequest`](#ttn.lorawan.v3.CreateApplicationAPIKeyRequest)
  - [Message `CreateApplicationRequest`](#ttn.lorawan.v3.CreateApplicationRequest)
  - [Message `DeleteApplicationAPIKeyRequest`](#ttn.lorawan.v3.DeleteApplicationAPIKeyRequest)
  - [Message `DeleteApplicationCollaboratorRequest`](#ttn.lorawan.v3.DeleteApplicationCollaboratorRequest)
  - [Message `GetApplicationAPIKeyRequest`](#ttn.lorawan.v3.GetApplicationAPIKeyRequest)
  - [Message `GetApplicationCollaboratorRequest`](#ttn.lorawan.v3.GetApplicationCollaboratorRequest)
  - [Message `GetApplicationRequest`](#ttn.lorawan.v3.GetApplicationRequest)
  - [Message `IssueDevEUIResponse`](#ttn.lorawan.v3.IssueDevEUIResponse)
  - [Message `ListApplicationAPIKeysRequest`](#ttn.lorawan.v3.ListApplicationAPIKeysRequest)
  - [Message `ListApplicationCollaboratorsRequest`](#ttn.lorawan.v3.ListApplicationCollaboratorsRequest)
  - [Message `ListApplicationsRequest`](#ttn.lorawan.v3.ListApplicationsRequest)
  - [Message `SetApplicationCollaboratorRequest`](#ttn.lorawan.v3.SetApplicationCollaboratorRequest)
  - [Message `UpdateApplicationAPIKeyRequest`](#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest)
  - [Message `UpdateApplicationRequest`](#ttn.lorawan.v3.UpdateApplicationRequest)
- [File `ttn/lorawan/v3/application_services.proto`](#ttn/lorawan/v3/application_services.proto)
  - [Service `ApplicationAccess`](#ttn.lorawan.v3.ApplicationAccess)
  - [Service `ApplicationRegistry`](#ttn.lorawan.v3.ApplicationRegistry)
- [File `ttn/lorawan/v3/applicationserver.proto`](#ttn/lorawan/v3/applicationserver.proto)
  - [Message `ApplicationLink`](#ttn.lorawan.v3.ApplicationLink)
  - [Message `ApplicationLinkStats`](#ttn.lorawan.v3.ApplicationLinkStats)
  - [Message `AsConfiguration`](#ttn.lorawan.v3.AsConfiguration)
  - [Message `AsConfiguration.PubSub`](#ttn.lorawan.v3.AsConfiguration.PubSub)
  - [Message `AsConfiguration.PubSub.Providers`](#ttn.lorawan.v3.AsConfiguration.PubSub.Providers)
  - [Message `AsConfiguration.Webhooks`](#ttn.lorawan.v3.AsConfiguration.Webhooks)
  - [Message `DecodeDownlinkRequest`](#ttn.lorawan.v3.DecodeDownlinkRequest)
  - [Message `DecodeDownlinkResponse`](#ttn.lorawan.v3.DecodeDownlinkResponse)
  - [Message `DecodeUplinkRequest`](#ttn.lorawan.v3.DecodeUplinkRequest)
  - [Message `DecodeUplinkResponse`](#ttn.lorawan.v3.DecodeUplinkResponse)
  - [Message `EncodeDownlinkRequest`](#ttn.lorawan.v3.EncodeDownlinkRequest)
  - [Message `EncodeDownlinkResponse`](#ttn.lorawan.v3.EncodeDownlinkResponse)
  - [Message `GetApplicationLinkRequest`](#ttn.lorawan.v3.GetApplicationLinkRequest)
  - [Message `GetAsConfigurationRequest`](#ttn.lorawan.v3.GetAsConfigurationRequest)
  - [Message `GetAsConfigurationResponse`](#ttn.lorawan.v3.GetAsConfigurationResponse)
  - [Message `NsAsHandleUplinkRequest`](#ttn.lorawan.v3.NsAsHandleUplinkRequest)
  - [Message `SetApplicationLinkRequest`](#ttn.lorawan.v3.SetApplicationLinkRequest)
  - [Enum `AsConfiguration.PubSub.Providers.Status`](#ttn.lorawan.v3.AsConfiguration.PubSub.Providers.Status)
  - [Service `AppAs`](#ttn.lorawan.v3.AppAs)
  - [Service `As`](#ttn.lorawan.v3.As)
  - [Service `AsEndDeviceBatchRegistry`](#ttn.lorawan.v3.AsEndDeviceBatchRegistry)
  - [Service `AsEndDeviceRegistry`](#ttn.lorawan.v3.AsEndDeviceRegistry)
  - [Service `NsAs`](#ttn.lorawan.v3.NsAs)
- [File `ttn/lorawan/v3/applicationserver_integrations_alcsync.proto`](#ttn/lorawan/v3/applicationserver_integrations_alcsync.proto)
  - [Message `ALCSyncCommand`](#ttn.lorawan.v3.ALCSyncCommand)
  - [Message `ALCSyncCommand.AppTimeAns`](#ttn.lorawan.v3.ALCSyncCommand.AppTimeAns)
  - [Message `ALCSyncCommand.AppTimeReq`](#ttn.lorawan.v3.ALCSyncCommand.AppTimeReq)
  - [Enum `ALCSyncCommandIdentifier`](#ttn.lorawan.v3.ALCSyncCommandIdentifier)
- [File `ttn/lorawan/v3/applicationserver_integrations_storage.proto`](#ttn/lorawan/v3/applicationserver_integrations_storage.proto)
  - [Message `ContinuationTokenPayload`](#ttn.lorawan.v3.ContinuationTokenPayload)
  - [Message `GetStoredApplicationUpCountRequest`](#ttn.lorawan.v3.GetStoredApplicationUpCountRequest)
  - [Message `GetStoredApplicationUpCountResponse`](#ttn.lorawan.v3.GetStoredApplicationUpCountResponse)
  - [Message `GetStoredApplicationUpCountResponse.CountEntry`](#ttn.lorawan.v3.GetStoredApplicationUpCountResponse.CountEntry)
  - [Message `GetStoredApplicationUpRequest`](#ttn.lorawan.v3.GetStoredApplicationUpRequest)
  - [Service `ApplicationUpStorage`](#ttn.lorawan.v3.ApplicationUpStorage)
- [File `ttn/lorawan/v3/applicationserver_packages.proto`](#ttn/lorawan/v3/applicationserver_packages.proto)
  - [Message `ApplicationPackage`](#ttn.lorawan.v3.ApplicationPackage)
  - [Message `ApplicationPackageAssociation`](#ttn.lorawan.v3.ApplicationPackageAssociation)
  - [Message `ApplicationPackageAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageAssociationIdentifiers)
  - [Message `ApplicationPackageAssociations`](#ttn.lorawan.v3.ApplicationPackageAssociations)
  - [Message `ApplicationPackageDefaultAssociation`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociation)
  - [Message `ApplicationPackageDefaultAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociationIdentifiers)
  - [Message `ApplicationPackageDefaultAssociations`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociations)
  - [Message `ApplicationPackages`](#ttn.lorawan.v3.ApplicationPackages)
  - [Message `GetApplicationPackageAssociationRequest`](#ttn.lorawan.v3.GetApplicationPackageAssociationRequest)
  - [Message `GetApplicationPackageDefaultAssociationRequest`](#ttn.lorawan.v3.GetApplicationPackageDefaultAssociationRequest)
  - [Message `ListApplicationPackageAssociationRequest`](#ttn.lorawan.v3.ListApplicationPackageAssociationRequest)
  - [Message `ListApplicationPackageDefaultAssociationRequest`](#ttn.lorawan.v3.ListApplicationPackageDefaultAssociationRequest)
  - [Message `SetApplicationPackageAssociationRequest`](#ttn.lorawan.v3.SetApplicationPackageAssociationRequest)
  - [Message `SetApplicationPackageDefaultAssociationRequest`](#ttn.lorawan.v3.SetApplicationPackageDefaultAssociationRequest)
  - [Service `ApplicationPackageRegistry`](#ttn.lorawan.v3.ApplicationPackageRegistry)
- [File `ttn/lorawan/v3/applicationserver_pubsub.proto`](#ttn/lorawan/v3/applicationserver_pubsub.proto)
  - [Message `ApplicationPubSub`](#ttn.lorawan.v3.ApplicationPubSub)
  - [Message `ApplicationPubSub.AWSIoTProvider`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider)
  - [Message `ApplicationPubSub.AWSIoTProvider.AccessKey`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.AccessKey)
  - [Message `ApplicationPubSub.AWSIoTProvider.AssumeRole`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.AssumeRole)
  - [Message `ApplicationPubSub.AWSIoTProvider.DefaultIntegration`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.DefaultIntegration)
  - [Message `ApplicationPubSub.MQTTProvider`](#ttn.lorawan.v3.ApplicationPubSub.MQTTProvider)
  - [Message `ApplicationPubSub.MQTTProvider.HeadersEntry`](#ttn.lorawan.v3.ApplicationPubSub.MQTTProvider.HeadersEntry)
  - [Message `ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message)
  - [Message `ApplicationPubSub.NATSProvider`](#ttn.lorawan.v3.ApplicationPubSub.NATSProvider)
  - [Message `ApplicationPubSubFormats`](#ttn.lorawan.v3.ApplicationPubSubFormats)
  - [Message `ApplicationPubSubFormats.FormatsEntry`](#ttn.lorawan.v3.ApplicationPubSubFormats.FormatsEntry)
  - [Message `ApplicationPubSubIdentifiers`](#ttn.lorawan.v3.ApplicationPubSubIdentifiers)
  - [Message `ApplicationPubSubs`](#ttn.lorawan.v3.ApplicationPubSubs)
  - [Message `GetApplicationPubSubRequest`](#ttn.lorawan.v3.GetApplicationPubSubRequest)
  - [Message `ListApplicationPubSubsRequest`](#ttn.lorawan.v3.ListApplicationPubSubsRequest)
  - [Message `SetApplicationPubSubRequest`](#ttn.lorawan.v3.SetApplicationPubSubRequest)
  - [Enum `ApplicationPubSub.MQTTProvider.QoS`](#ttn.lorawan.v3.ApplicationPubSub.MQTTProvider.QoS)
  - [Service `ApplicationPubSubRegistry`](#ttn.lorawan.v3.ApplicationPubSubRegistry)
- [File `ttn/lorawan/v3/applicationserver_web.proto`](#ttn/lorawan/v3/applicationserver_web.proto)
  - [Message `ApplicationWebhook`](#ttn.lorawan.v3.ApplicationWebhook)
  - [Message `ApplicationWebhook.HeadersEntry`](#ttn.lorawan.v3.ApplicationWebhook.HeadersEntry)
  - [Message `ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message)
  - [Message `ApplicationWebhook.TemplateFieldsEntry`](#ttn.lorawan.v3.ApplicationWebhook.TemplateFieldsEntry)
  - [Message `ApplicationWebhookFormats`](#ttn.lorawan.v3.ApplicationWebhookFormats)
  - [Message `ApplicationWebhookFormats.FormatsEntry`](#ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry)
  - [Message `ApplicationWebhookHealth`](#ttn.lorawan.v3.ApplicationWebhookHealth)
  - [Message `ApplicationWebhookHealth.WebhookHealthStatusHealthy`](#ttn.lorawan.v3.ApplicationWebhookHealth.WebhookHealthStatusHealthy)
  - [Message `ApplicationWebhookHealth.WebhookHealthStatusUnhealthy`](#ttn.lorawan.v3.ApplicationWebhookHealth.WebhookHealthStatusUnhealthy)
  - [Message `ApplicationWebhookIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookIdentifiers)
  - [Message `ApplicationWebhookTemplate`](#ttn.lorawan.v3.ApplicationWebhookTemplate)
  - [Message `ApplicationWebhookTemplate.HeadersEntry`](#ttn.lorawan.v3.ApplicationWebhookTemplate.HeadersEntry)
  - [Message `ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message)
  - [Message `ApplicationWebhookTemplateField`](#ttn.lorawan.v3.ApplicationWebhookTemplateField)
  - [Message `ApplicationWebhookTemplateIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookTemplateIdentifiers)
  - [Message `ApplicationWebhookTemplates`](#ttn.lorawan.v3.ApplicationWebhookTemplates)
  - [Message `ApplicationWebhooks`](#ttn.lorawan.v3.ApplicationWebhooks)
  - [Message `GetApplicationWebhookRequest`](#ttn.lorawan.v3.GetApplicationWebhookRequest)
  - [Message `GetApplicationWebhookTemplateRequest`](#ttn.lorawan.v3.GetApplicationWebhookTemplateRequest)
  - [Message `ListApplicationWebhookTemplatesRequest`](#ttn.lorawan.v3.ListApplicationWebhookTemplatesRequest)
  - [Message `ListApplicationWebhooksRequest`](#ttn.lorawan.v3.ListApplicationWebhooksRequest)
  - [Message `SetApplicationWebhookRequest`](#ttn.lorawan.v3.SetApplicationWebhookRequest)
  - [Service `ApplicationWebhookRegistry`](#ttn.lorawan.v3.ApplicationWebhookRegistry)
- [File `ttn/lorawan/v3/client.proto`](#ttn/lorawan/v3/client.proto)
  - [Message `Client`](#ttn.lorawan.v3.Client)
  - [Message `Client.AttributesEntry`](#ttn.lorawan.v3.Client.AttributesEntry)
  - [Message `Clients`](#ttn.lorawan.v3.Clients)
  - [Message `CreateClientRequest`](#ttn.lorawan.v3.CreateClientRequest)
  - [Message `DeleteClientCollaboratorRequest`](#ttn.lorawan.v3.DeleteClientCollaboratorRequest)
  - [Message `GetClientCollaboratorRequest`](#ttn.lorawan.v3.GetClientCollaboratorRequest)
  - [Message `GetClientRequest`](#ttn.lorawan.v3.GetClientRequest)
  - [Message `ListClientCollaboratorsRequest`](#ttn.lorawan.v3.ListClientCollaboratorsRequest)
  - [Message `ListClientsRequest`](#ttn.lorawan.v3.ListClientsRequest)
  - [Message `SetClientCollaboratorRequest`](#ttn.lorawan.v3.SetClientCollaboratorRequest)
  - [Message `UpdateClientRequest`](#ttn.lorawan.v3.UpdateClientRequest)
  - [Enum `GrantType`](#ttn.lorawan.v3.GrantType)
- [File `ttn/lorawan/v3/client_services.proto`](#ttn/lorawan/v3/client_services.proto)
  - [Service `ClientAccess`](#ttn.lorawan.v3.ClientAccess)
  - [Service `ClientRegistry`](#ttn.lorawan.v3.ClientRegistry)
- [File `ttn/lorawan/v3/cluster.proto`](#ttn/lorawan/v3/cluster.proto)
  - [Message `PeerInfo`](#ttn.lorawan.v3.PeerInfo)
  - [Message `PeerInfo.TagsEntry`](#ttn.lorawan.v3.PeerInfo.TagsEntry)
- [File `ttn/lorawan/v3/configuration_services.proto`](#ttn/lorawan/v3/configuration_services.proto)
  - [Message `BandDescription`](#ttn.lorawan.v3.BandDescription)
  - [Message `BandDescription.BandDataRate`](#ttn.lorawan.v3.BandDescription.BandDataRate)
  - [Message `BandDescription.Beacon`](#ttn.lorawan.v3.BandDescription.Beacon)
  - [Message `BandDescription.Channel`](#ttn.lorawan.v3.BandDescription.Channel)
  - [Message `BandDescription.DataRatesEntry`](#ttn.lorawan.v3.BandDescription.DataRatesEntry)
  - [Message `BandDescription.DwellTime`](#ttn.lorawan.v3.BandDescription.DwellTime)
  - [Message `BandDescription.RelayParameters`](#ttn.lorawan.v3.BandDescription.RelayParameters)
  - [Message `BandDescription.RelayParameters.RelayWORChannel`](#ttn.lorawan.v3.BandDescription.RelayParameters.RelayWORChannel)
  - [Message `BandDescription.Rx2Parameters`](#ttn.lorawan.v3.BandDescription.Rx2Parameters)
  - [Message `BandDescription.SubBandParameters`](#ttn.lorawan.v3.BandDescription.SubBandParameters)
  - [Message `FrequencyPlanDescription`](#ttn.lorawan.v3.FrequencyPlanDescription)
  - [Message `GetPhyVersionsRequest`](#ttn.lorawan.v3.GetPhyVersionsRequest)
  - [Message `GetPhyVersionsResponse`](#ttn.lorawan.v3.GetPhyVersionsResponse)
  - [Message `GetPhyVersionsResponse.VersionInfo`](#ttn.lorawan.v3.GetPhyVersionsResponse.VersionInfo)
  - [Message `ListBandsRequest`](#ttn.lorawan.v3.ListBandsRequest)
  - [Message `ListBandsResponse`](#ttn.lorawan.v3.ListBandsResponse)
  - [Message `ListBandsResponse.DescriptionsEntry`](#ttn.lorawan.v3.ListBandsResponse.DescriptionsEntry)
  - [Message `ListBandsResponse.VersionedBandDescription`](#ttn.lorawan.v3.ListBandsResponse.VersionedBandDescription)
  - [Message `ListBandsResponse.VersionedBandDescription.BandEntry`](#ttn.lorawan.v3.ListBandsResponse.VersionedBandDescription.BandEntry)
  - [Message `ListFrequencyPlansRequest`](#ttn.lorawan.v3.ListFrequencyPlansRequest)
  - [Message `ListFrequencyPlansResponse`](#ttn.lorawan.v3.ListFrequencyPlansResponse)
  - [Service `Configuration`](#ttn.lorawan.v3.Configuration)
- [File `ttn/lorawan/v3/contact_info.proto`](#ttn/lorawan/v3/contact_info.proto)
  - [Message `ContactInfo`](#ttn.lorawan.v3.ContactInfo)
  - [Message `ContactInfoValidation`](#ttn.lorawan.v3.ContactInfoValidation)
  - [Enum `ContactMethod`](#ttn.lorawan.v3.ContactMethod)
  - [Enum `ContactType`](#ttn.lorawan.v3.ContactType)
  - [Service `ContactInfoRegistry`](#ttn.lorawan.v3.ContactInfoRegistry)
- [File `ttn/lorawan/v3/deviceclaimingserver.proto`](#ttn/lorawan/v3/deviceclaimingserver.proto)
  - [Message `AuthorizeApplicationRequest`](#ttn.lorawan.v3.AuthorizeApplicationRequest)
  - [Message `AuthorizeGatewayRequest`](#ttn.lorawan.v3.AuthorizeGatewayRequest)
  - [Message `BatchUnclaimEndDevicesRequest`](#ttn.lorawan.v3.BatchUnclaimEndDevicesRequest)
  - [Message `BatchUnclaimEndDevicesResponse`](#ttn.lorawan.v3.BatchUnclaimEndDevicesResponse)
  - [Message `BatchUnclaimEndDevicesResponse.FailedEntry`](#ttn.lorawan.v3.BatchUnclaimEndDevicesResponse.FailedEntry)
  - [Message `CUPSRedirection`](#ttn.lorawan.v3.CUPSRedirection)
  - [Message `CUPSRedirection.ClientTLS`](#ttn.lorawan.v3.CUPSRedirection.ClientTLS)
  - [Message `ClaimEndDeviceRequest`](#ttn.lorawan.v3.ClaimEndDeviceRequest)
  - [Message `ClaimEndDeviceRequest.AuthenticatedIdentifiers`](#ttn.lorawan.v3.ClaimEndDeviceRequest.AuthenticatedIdentifiers)
  - [Message `ClaimGatewayRequest`](#ttn.lorawan.v3.ClaimGatewayRequest)
  - [Message `ClaimGatewayRequest.AuthenticatedIdentifiers`](#ttn.lorawan.v3.ClaimGatewayRequest.AuthenticatedIdentifiers)
  - [Message `GetClaimStatusResponse`](#ttn.lorawan.v3.GetClaimStatusResponse)
  - [Message `GetClaimStatusResponse.VendorSpecific`](#ttn.lorawan.v3.GetClaimStatusResponse.VendorSpecific)
  - [Message `GetInfoByGatewayEUIRequest`](#ttn.lorawan.v3.GetInfoByGatewayEUIRequest)
  - [Message `GetInfoByGatewayEUIResponse`](#ttn.lorawan.v3.GetInfoByGatewayEUIResponse)
  - [Message `GetInfoByJoinEUIRequest`](#ttn.lorawan.v3.GetInfoByJoinEUIRequest)
  - [Message `GetInfoByJoinEUIResponse`](#ttn.lorawan.v3.GetInfoByJoinEUIResponse)
  - [Message `GetInfoByJoinEUIsRequest`](#ttn.lorawan.v3.GetInfoByJoinEUIsRequest)
  - [Message `GetInfoByJoinEUIsResponse`](#ttn.lorawan.v3.GetInfoByJoinEUIsResponse)
  - [Service `EndDeviceBatchClaimingServer`](#ttn.lorawan.v3.EndDeviceBatchClaimingServer)
  - [Service `EndDeviceClaimingServer`](#ttn.lorawan.v3.EndDeviceClaimingServer)
  - [Service `GatewayClaimingServer`](#ttn.lorawan.v3.GatewayClaimingServer)
- [File `ttn/lorawan/v3/devicerepository.proto`](#ttn/lorawan/v3/devicerepository.proto)
  - [Message `DecodedMessagePayload`](#ttn.lorawan.v3.DecodedMessagePayload)
  - [Message `EncodedMessagePayload`](#ttn.lorawan.v3.EncodedMessagePayload)
  - [Message `EndDeviceBrand`](#ttn.lorawan.v3.EndDeviceBrand)
  - [Message `EndDeviceModel`](#ttn.lorawan.v3.EndDeviceModel)
  - [Message `EndDeviceModel.Battery`](#ttn.lorawan.v3.EndDeviceModel.Battery)
  - [Message `EndDeviceModel.Compliances`](#ttn.lorawan.v3.EndDeviceModel.Compliances)
  - [Message `EndDeviceModel.Compliances.Compliance`](#ttn.lorawan.v3.EndDeviceModel.Compliances.Compliance)
  - [Message `EndDeviceModel.Dimensions`](#ttn.lorawan.v3.EndDeviceModel.Dimensions)
  - [Message `EndDeviceModel.FirmwareVersion`](#ttn.lorawan.v3.EndDeviceModel.FirmwareVersion)
  - [Message `EndDeviceModel.FirmwareVersion.Profile`](#ttn.lorawan.v3.EndDeviceModel.FirmwareVersion.Profile)
  - [Message `EndDeviceModel.FirmwareVersion.ProfilesEntry`](#ttn.lorawan.v3.EndDeviceModel.FirmwareVersion.ProfilesEntry)
  - [Message `EndDeviceModel.HardwareVersion`](#ttn.lorawan.v3.EndDeviceModel.HardwareVersion)
  - [Message `EndDeviceModel.OperatingConditions`](#ttn.lorawan.v3.EndDeviceModel.OperatingConditions)
  - [Message `EndDeviceModel.OperatingConditions.Limits`](#ttn.lorawan.v3.EndDeviceModel.OperatingConditions.Limits)
  - [Message `EndDeviceModel.Photos`](#ttn.lorawan.v3.EndDeviceModel.Photos)
  - [Message `EndDeviceModel.Reseller`](#ttn.lorawan.v3.EndDeviceModel.Reseller)
  - [Message `EndDeviceModel.Videos`](#ttn.lorawan.v3.EndDeviceModel.Videos)
  - [Message `GetEndDeviceBrandRequest`](#ttn.lorawan.v3.GetEndDeviceBrandRequest)
  - [Message `GetEndDeviceModelRequest`](#ttn.lorawan.v3.GetEndDeviceModelRequest)
  - [Message `GetPayloadFormatterRequest`](#ttn.lorawan.v3.GetPayloadFormatterRequest)
  - [Message `GetTemplateRequest`](#ttn.lorawan.v3.GetTemplateRequest)
  - [Message `GetTemplateRequest.EndDeviceProfileIdentifiers`](#ttn.lorawan.v3.GetTemplateRequest.EndDeviceProfileIdentifiers)
  - [Message `ListEndDeviceBrandsRequest`](#ttn.lorawan.v3.ListEndDeviceBrandsRequest)
  - [Message `ListEndDeviceBrandsResponse`](#ttn.lorawan.v3.ListEndDeviceBrandsResponse)
  - [Message `ListEndDeviceModelsRequest`](#ttn.lorawan.v3.ListEndDeviceModelsRequest)
  - [Message `ListEndDeviceModelsResponse`](#ttn.lorawan.v3.ListEndDeviceModelsResponse)
  - [Message `MessagePayloadDecoder`](#ttn.lorawan.v3.MessagePayloadDecoder)
  - [Message `MessagePayloadDecoder.Example`](#ttn.lorawan.v3.MessagePayloadDecoder.Example)
  - [Message `MessagePayloadEncoder`](#ttn.lorawan.v3.MessagePayloadEncoder)
  - [Message `MessagePayloadEncoder.Example`](#ttn.lorawan.v3.MessagePayloadEncoder.Example)
  - [Enum `KeyProvisioning`](#ttn.lorawan.v3.KeyProvisioning)
  - [Enum `KeySecurity`](#ttn.lorawan.v3.KeySecurity)
  - [Service `DeviceRepository`](#ttn.lorawan.v3.DeviceRepository)
- [File `ttn/lorawan/v3/email_messages.proto`](#ttn/lorawan/v3/email_messages.proto)
  - [Message `CreateClientEmailMessage`](#ttn.lorawan.v3.CreateClientEmailMessage)
- [File `ttn/lorawan/v3/end_device.proto`](#ttn/lorawan/v3/end_device.proto)
  - [Message `ADRSettings`](#ttn.lorawan.v3.ADRSettings)
  - [Message `ADRSettings.DisabledMode`](#ttn.lorawan.v3.ADRSettings.DisabledMode)
  - [Message `ADRSettings.DynamicMode`](#ttn.lorawan.v3.ADRSettings.DynamicMode)
  - [Message `ADRSettings.DynamicMode.ChannelSteeringSettings`](#ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings)
  - [Message `ADRSettings.DynamicMode.ChannelSteeringSettings.DisabledMode`](#ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings.DisabledMode)
  - [Message `ADRSettings.DynamicMode.ChannelSteeringSettings.LoRaNarrowMode`](#ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings.LoRaNarrowMode)
  - [Message `ADRSettings.StaticMode`](#ttn.lorawan.v3.ADRSettings.StaticMode)
  - [Message `BatchDeleteEndDevicesRequest`](#ttn.lorawan.v3.BatchDeleteEndDevicesRequest)
  - [Message `BatchGetEndDevicesRequest`](#ttn.lorawan.v3.BatchGetEndDevicesRequest)
  - [Message `BatchUpdateEndDeviceLastSeenRequest`](#ttn.lorawan.v3.BatchUpdateEndDeviceLastSeenRequest)
  - [Message `BatchUpdateEndDeviceLastSeenRequest.EndDeviceLastSeenUpdate`](#ttn.lorawan.v3.BatchUpdateEndDeviceLastSeenRequest.EndDeviceLastSeenUpdate)
  - [Message `BoolValue`](#ttn.lorawan.v3.BoolValue)
  - [Message `ConvertEndDeviceTemplateRequest`](#ttn.lorawan.v3.ConvertEndDeviceTemplateRequest)
  - [Message `CreateEndDeviceRequest`](#ttn.lorawan.v3.CreateEndDeviceRequest)
  - [Message `DevAddrPrefix`](#ttn.lorawan.v3.DevAddrPrefix)
  - [Message `EndDevice`](#ttn.lorawan.v3.EndDevice)
  - [Message `EndDevice.AttributesEntry`](#ttn.lorawan.v3.EndDevice.AttributesEntry)
  - [Message `EndDevice.LocationsEntry`](#ttn.lorawan.v3.EndDevice.LocationsEntry)
  - [Message `EndDeviceAuthenticationCode`](#ttn.lorawan.v3.EndDeviceAuthenticationCode)
  - [Message `EndDeviceTemplate`](#ttn.lorawan.v3.EndDeviceTemplate)
  - [Message `EndDeviceTemplateFormat`](#ttn.lorawan.v3.EndDeviceTemplateFormat)
  - [Message `EndDeviceTemplateFormats`](#ttn.lorawan.v3.EndDeviceTemplateFormats)
  - [Message `EndDeviceTemplateFormats.FormatsEntry`](#ttn.lorawan.v3.EndDeviceTemplateFormats.FormatsEntry)
  - [Message `EndDeviceVersion`](#ttn.lorawan.v3.EndDeviceVersion)
  - [Message `EndDevices`](#ttn.lorawan.v3.EndDevices)
  - [Message `GetEndDeviceIdentifiersForEUIsRequest`](#ttn.lorawan.v3.GetEndDeviceIdentifiersForEUIsRequest)
  - [Message `GetEndDeviceRequest`](#ttn.lorawan.v3.GetEndDeviceRequest)
  - [Message `ListEndDevicesRequest`](#ttn.lorawan.v3.ListEndDevicesRequest)
  - [Message `MACParameters`](#ttn.lorawan.v3.MACParameters)
  - [Message `MACParameters.Channel`](#ttn.lorawan.v3.MACParameters.Channel)
  - [Message `MACSettings`](#ttn.lorawan.v3.MACSettings)
  - [Message `MACState`](#ttn.lorawan.v3.MACState)
  - [Message `MACState.DataRateRange`](#ttn.lorawan.v3.MACState.DataRateRange)
  - [Message `MACState.DataRateRanges`](#ttn.lorawan.v3.MACState.DataRateRanges)
  - [Message `MACState.DownlinkMessage`](#ttn.lorawan.v3.MACState.DownlinkMessage)
  - [Message `MACState.DownlinkMessage.Message`](#ttn.lorawan.v3.MACState.DownlinkMessage.Message)
  - [Message `MACState.DownlinkMessage.Message.MACPayload`](#ttn.lorawan.v3.MACState.DownlinkMessage.Message.MACPayload)
  - [Message `MACState.DownlinkMessage.Message.MHDR`](#ttn.lorawan.v3.MACState.DownlinkMessage.Message.MHDR)
  - [Message `MACState.JoinAccept`](#ttn.lorawan.v3.MACState.JoinAccept)
  - [Message `MACState.JoinRequest`](#ttn.lorawan.v3.MACState.JoinRequest)
  - [Message `MACState.RejectedDataRateRangesEntry`](#ttn.lorawan.v3.MACState.RejectedDataRateRangesEntry)
  - [Message `MACState.UplinkMessage`](#ttn.lorawan.v3.MACState.UplinkMessage)
  - [Message `MACState.UplinkMessage.RxMetadata`](#ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata)
  - [Message `MACState.UplinkMessage.RxMetadata.PacketBrokerMetadata`](#ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata.PacketBrokerMetadata)
  - [Message `MACState.UplinkMessage.RxMetadata.RelayMetadata`](#ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata.RelayMetadata)
  - [Message `MACState.UplinkMessage.TxSettings`](#ttn.lorawan.v3.MACState.UplinkMessage.TxSettings)
  - [Message `RelayParameters`](#ttn.lorawan.v3.RelayParameters)
  - [Message `RelayUplinkForwardingRule`](#ttn.lorawan.v3.RelayUplinkForwardingRule)
  - [Message `ResetAndGetEndDeviceRequest`](#ttn.lorawan.v3.ResetAndGetEndDeviceRequest)
  - [Message `ServedRelayParameters`](#ttn.lorawan.v3.ServedRelayParameters)
  - [Message `ServingRelayForwardingLimits`](#ttn.lorawan.v3.ServingRelayForwardingLimits)
  - [Message `ServingRelayParameters`](#ttn.lorawan.v3.ServingRelayParameters)
  - [Message `Session`](#ttn.lorawan.v3.Session)
  - [Message `SetEndDeviceRequest`](#ttn.lorawan.v3.SetEndDeviceRequest)
  - [Message `UpdateEndDeviceRequest`](#ttn.lorawan.v3.UpdateEndDeviceRequest)
  - [Enum `PowerState`](#ttn.lorawan.v3.PowerState)
- [File `ttn/lorawan/v3/end_device_services.proto`](#ttn/lorawan/v3/end_device_services.proto)
  - [Service `EndDeviceBatchRegistry`](#ttn.lorawan.v3.EndDeviceBatchRegistry)
  - [Service `EndDeviceRegistry`](#ttn.lorawan.v3.EndDeviceRegistry)
  - [Service `EndDeviceTemplateConverter`](#ttn.lorawan.v3.EndDeviceTemplateConverter)
- [File `ttn/lorawan/v3/enums.proto`](#ttn/lorawan/v3/enums.proto)
  - [Enum `ClusterRole`](#ttn.lorawan.v3.ClusterRole)
  - [Enum `DownlinkPathConstraint`](#ttn.lorawan.v3.DownlinkPathConstraint)
  - [Enum `State`](#ttn.lorawan.v3.State)
- [File `ttn/lorawan/v3/error.proto`](#ttn/lorawan/v3/error.proto)
  - [Message `ErrorDetails`](#ttn.lorawan.v3.ErrorDetails)
- [File `ttn/lorawan/v3/events.proto`](#ttn/lorawan/v3/events.proto)
  - [Message `Event`](#ttn.lorawan.v3.Event)
  - [Message `Event.Authentication`](#ttn.lorawan.v3.Event.Authentication)
  - [Message `Event.ContextEntry`](#ttn.lorawan.v3.Event.ContextEntry)
  - [Message `FindRelatedEventsRequest`](#ttn.lorawan.v3.FindRelatedEventsRequest)
  - [Message `FindRelatedEventsResponse`](#ttn.lorawan.v3.FindRelatedEventsResponse)
  - [Message `StreamEventsRequest`](#ttn.lorawan.v3.StreamEventsRequest)
  - [Service `Events`](#ttn.lorawan.v3.Events)
- [File `ttn/lorawan/v3/gateway.proto`](#ttn/lorawan/v3/gateway.proto)
  - [Message `CreateGatewayAPIKeyRequest`](#ttn.lorawan.v3.CreateGatewayAPIKeyRequest)
  - [Message `CreateGatewayRequest`](#ttn.lorawan.v3.CreateGatewayRequest)
  - [Message `DeleteGatewayAPIKeyRequest`](#ttn.lorawan.v3.DeleteGatewayAPIKeyRequest)
  - [Message `DeleteGatewayCollaboratorRequest`](#ttn.lorawan.v3.DeleteGatewayCollaboratorRequest)
  - [Message `Gateway`](#ttn.lorawan.v3.Gateway)
  - [Message `Gateway.AttributesEntry`](#ttn.lorawan.v3.Gateway.AttributesEntry)
  - [Message `Gateway.LRFHSS`](#ttn.lorawan.v3.Gateway.LRFHSS)
  - [Message `GatewayAntenna`](#ttn.lorawan.v3.GatewayAntenna)
  - [Message `GatewayAntenna.AttributesEntry`](#ttn.lorawan.v3.GatewayAntenna.AttributesEntry)
  - [Message `GatewayBrand`](#ttn.lorawan.v3.GatewayBrand)
  - [Message `GatewayClaimAuthenticationCode`](#ttn.lorawan.v3.GatewayClaimAuthenticationCode)
  - [Message `GatewayConnectionStats`](#ttn.lorawan.v3.GatewayConnectionStats)
  - [Message `GatewayConnectionStats.RoundTripTimes`](#ttn.lorawan.v3.GatewayConnectionStats.RoundTripTimes)
  - [Message `GatewayConnectionStats.SubBand`](#ttn.lorawan.v3.GatewayConnectionStats.SubBand)
  - [Message `GatewayModel`](#ttn.lorawan.v3.GatewayModel)
  - [Message `GatewayRadio`](#ttn.lorawan.v3.GatewayRadio)
  - [Message `GatewayRadio.TxConfiguration`](#ttn.lorawan.v3.GatewayRadio.TxConfiguration)
  - [Message `GatewayRemoteAddress`](#ttn.lorawan.v3.GatewayRemoteAddress)
  - [Message `GatewayStatus`](#ttn.lorawan.v3.GatewayStatus)
  - [Message `GatewayStatus.MetricsEntry`](#ttn.lorawan.v3.GatewayStatus.MetricsEntry)
  - [Message `GatewayStatus.VersionsEntry`](#ttn.lorawan.v3.GatewayStatus.VersionsEntry)
  - [Message `GatewayVersionIdentifiers`](#ttn.lorawan.v3.GatewayVersionIdentifiers)
  - [Message `Gateways`](#ttn.lorawan.v3.Gateways)
  - [Message `GetGatewayAPIKeyRequest`](#ttn.lorawan.v3.GetGatewayAPIKeyRequest)
  - [Message `GetGatewayCollaboratorRequest`](#ttn.lorawan.v3.GetGatewayCollaboratorRequest)
  - [Message `GetGatewayIdentifiersForEUIRequest`](#ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest)
  - [Message `GetGatewayRequest`](#ttn.lorawan.v3.GetGatewayRequest)
  - [Message `ListGatewayAPIKeysRequest`](#ttn.lorawan.v3.ListGatewayAPIKeysRequest)
  - [Message `ListGatewayCollaboratorsRequest`](#ttn.lorawan.v3.ListGatewayCollaboratorsRequest)
  - [Message `ListGatewaysRequest`](#ttn.lorawan.v3.ListGatewaysRequest)
  - [Message `SetGatewayCollaboratorRequest`](#ttn.lorawan.v3.SetGatewayCollaboratorRequest)
  - [Message `UpdateGatewayAPIKeyRequest`](#ttn.lorawan.v3.UpdateGatewayAPIKeyRequest)
  - [Message `UpdateGatewayRequest`](#ttn.lorawan.v3.UpdateGatewayRequest)
  - [Enum `GatewayAntennaPlacement`](#ttn.lorawan.v3.GatewayAntennaPlacement)
- [File `ttn/lorawan/v3/gateway_configuration.proto`](#ttn/lorawan/v3/gateway_configuration.proto)
  - [Message `GetGatewayConfigurationRequest`](#ttn.lorawan.v3.GetGatewayConfigurationRequest)
  - [Message `GetGatewayConfigurationResponse`](#ttn.lorawan.v3.GetGatewayConfigurationResponse)
  - [Service `GatewayConfigurationService`](#ttn.lorawan.v3.GatewayConfigurationService)
- [File `ttn/lorawan/v3/gateway_services.proto`](#ttn/lorawan/v3/gateway_services.proto)
  - [Message `AssertGatewayRightsRequest`](#ttn.lorawan.v3.AssertGatewayRightsRequest)
  - [Message `BatchDeleteGatewaysRequest`](#ttn.lorawan.v3.BatchDeleteGatewaysRequest)
  - [Message `PullGatewayConfigurationRequest`](#ttn.lorawan.v3.PullGatewayConfigurationRequest)
  - [Service `GatewayAccess`](#ttn.lorawan.v3.GatewayAccess)
  - [Service `GatewayBatchAccess`](#ttn.lorawan.v3.GatewayBatchAccess)
  - [Service `GatewayBatchRegistry`](#ttn.lorawan.v3.GatewayBatchRegistry)
  - [Service `GatewayConfigurator`](#ttn.lorawan.v3.GatewayConfigurator)
  - [Service `GatewayRegistry`](#ttn.lorawan.v3.GatewayRegistry)
- [File `ttn/lorawan/v3/gatewayserver.proto`](#ttn/lorawan/v3/gatewayserver.proto)
  - [Message `BatchGetGatewayConnectionStatsRequest`](#ttn.lorawan.v3.BatchGetGatewayConnectionStatsRequest)
  - [Message `BatchGetGatewayConnectionStatsResponse`](#ttn.lorawan.v3.BatchGetGatewayConnectionStatsResponse)
  - [Message `BatchGetGatewayConnectionStatsResponse.EntriesEntry`](#ttn.lorawan.v3.BatchGetGatewayConnectionStatsResponse.EntriesEntry)
  - [Message `GatewayDown`](#ttn.lorawan.v3.GatewayDown)
  - [Message `GatewayUp`](#ttn.lorawan.v3.GatewayUp)
  - [Message `ScheduleDownlinkErrorDetails`](#ttn.lorawan.v3.ScheduleDownlinkErrorDetails)
  - [Message `ScheduleDownlinkResponse`](#ttn.lorawan.v3.ScheduleDownlinkResponse)
  - [Service `Gs`](#ttn.lorawan.v3.Gs)
  - [Service `GtwGs`](#ttn.lorawan.v3.GtwGs)
  - [Service `NsGs`](#ttn.lorawan.v3.NsGs)
- [File `ttn/lorawan/v3/identifiers.proto`](#ttn/lorawan/v3/identifiers.proto)
  - [Message `ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers)
  - [Message `ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers)
  - [Message `EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers)
  - [Message `EndDeviceIdentifiersList`](#ttn.lorawan.v3.EndDeviceIdentifiersList)
  - [Message `EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers)
  - [Message `EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers)
  - [Message `GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers)
  - [Message `GatewayIdentifiersList`](#ttn.lorawan.v3.GatewayIdentifiersList)
  - [Message `LoRaAllianceProfileIdentifiers`](#ttn.lorawan.v3.LoRaAllianceProfileIdentifiers)
  - [Message `NetworkIdentifiers`](#ttn.lorawan.v3.NetworkIdentifiers)
  - [Message `OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers)
  - [Message `OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers)
  - [Message `UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers)
- [File `ttn/lorawan/v3/identityserver.proto`](#ttn/lorawan/v3/identityserver.proto)
  - [Message `AuthInfoResponse`](#ttn.lorawan.v3.AuthInfoResponse)
  - [Message `AuthInfoResponse.APIKeyAccess`](#ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess)
  - [Message `AuthInfoResponse.GatewayToken`](#ttn.lorawan.v3.AuthInfoResponse.GatewayToken)
  - [Message `GetIsConfigurationRequest`](#ttn.lorawan.v3.GetIsConfigurationRequest)
  - [Message `GetIsConfigurationResponse`](#ttn.lorawan.v3.GetIsConfigurationResponse)
  - [Message `IsConfiguration`](#ttn.lorawan.v3.IsConfiguration)
  - [Message `IsConfiguration.AdminRights`](#ttn.lorawan.v3.IsConfiguration.AdminRights)
  - [Message `IsConfiguration.CollaboratorRights`](#ttn.lorawan.v3.IsConfiguration.CollaboratorRights)
  - [Message `IsConfiguration.EndDevicePicture`](#ttn.lorawan.v3.IsConfiguration.EndDevicePicture)
  - [Message `IsConfiguration.ProfilePicture`](#ttn.lorawan.v3.IsConfiguration.ProfilePicture)
  - [Message `IsConfiguration.UserLogin`](#ttn.lorawan.v3.IsConfiguration.UserLogin)
  - [Message `IsConfiguration.UserRegistration`](#ttn.lorawan.v3.IsConfiguration.UserRegistration)
  - [Message `IsConfiguration.UserRegistration.AdminApproval`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.AdminApproval)
  - [Message `IsConfiguration.UserRegistration.ContactInfoValidation`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.ContactInfoValidation)
  - [Message `IsConfiguration.UserRegistration.Invitation`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.Invitation)
  - [Message `IsConfiguration.UserRegistration.PasswordRequirements`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.PasswordRequirements)
  - [Message `IsConfiguration.UserRights`](#ttn.lorawan.v3.IsConfiguration.UserRights)
  - [Service `EntityAccess`](#ttn.lorawan.v3.EntityAccess)
  - [Service `Is`](#ttn.lorawan.v3.Is)
- [File `ttn/lorawan/v3/join.proto`](#ttn/lorawan/v3/join.proto)
  - [Message `JoinRequest`](#ttn.lorawan.v3.JoinRequest)
  - [Message `JoinResponse`](#ttn.lorawan.v3.JoinResponse)
- [File `ttn/lorawan/v3/joinserver.proto`](#ttn/lorawan/v3/joinserver.proto)
  - [Message `AppSKeyResponse`](#ttn.lorawan.v3.AppSKeyResponse)
  - [Message `ApplicationActivationSettings`](#ttn.lorawan.v3.ApplicationActivationSettings)
  - [Message `CryptoServicePayloadRequest`](#ttn.lorawan.v3.CryptoServicePayloadRequest)
  - [Message `CryptoServicePayloadResponse`](#ttn.lorawan.v3.CryptoServicePayloadResponse)
  - [Message `DeleteApplicationActivationSettingsRequest`](#ttn.lorawan.v3.DeleteApplicationActivationSettingsRequest)
  - [Message `DeriveSessionKeysRequest`](#ttn.lorawan.v3.DeriveSessionKeysRequest)
  - [Message `GetApplicationActivationSettingsRequest`](#ttn.lorawan.v3.GetApplicationActivationSettingsRequest)
  - [Message `GetDefaultJoinEUIResponse`](#ttn.lorawan.v3.GetDefaultJoinEUIResponse)
  - [Message `GetRootKeysRequest`](#ttn.lorawan.v3.GetRootKeysRequest)
  - [Message `JoinAcceptMICRequest`](#ttn.lorawan.v3.JoinAcceptMICRequest)
  - [Message `JoinEUIPrefix`](#ttn.lorawan.v3.JoinEUIPrefix)
  - [Message `JoinEUIPrefixes`](#ttn.lorawan.v3.JoinEUIPrefixes)
  - [Message `NwkSKeysResponse`](#ttn.lorawan.v3.NwkSKeysResponse)
  - [Message `ProvisionEndDevicesRequest`](#ttn.lorawan.v3.ProvisionEndDevicesRequest)
  - [Message `ProvisionEndDevicesRequest.IdentifiersFromData`](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData)
  - [Message `ProvisionEndDevicesRequest.IdentifiersList`](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList)
  - [Message `ProvisionEndDevicesRequest.IdentifiersRange`](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange)
  - [Message `SessionKeyRequest`](#ttn.lorawan.v3.SessionKeyRequest)
  - [Message `SetApplicationActivationSettingsRequest`](#ttn.lorawan.v3.SetApplicationActivationSettingsRequest)
  - [Service `AppJs`](#ttn.lorawan.v3.AppJs)
  - [Service `ApplicationActivationSettingRegistry`](#ttn.lorawan.v3.ApplicationActivationSettingRegistry)
  - [Service `ApplicationCryptoService`](#ttn.lorawan.v3.ApplicationCryptoService)
  - [Service `AsJs`](#ttn.lorawan.v3.AsJs)
  - [Service `Js`](#ttn.lorawan.v3.Js)
  - [Service `JsEndDeviceBatchRegistry`](#ttn.lorawan.v3.JsEndDeviceBatchRegistry)
  - [Service `JsEndDeviceRegistry`](#ttn.lorawan.v3.JsEndDeviceRegistry)
  - [Service `NetworkCryptoService`](#ttn.lorawan.v3.NetworkCryptoService)
  - [Service `NsJs`](#ttn.lorawan.v3.NsJs)
- [File `ttn/lorawan/v3/keys.proto`](#ttn/lorawan/v3/keys.proto)
  - [Message `KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope)
  - [Message `RootKeys`](#ttn.lorawan.v3.RootKeys)
  - [Message `SessionKeys`](#ttn.lorawan.v3.SessionKeys)
- [File `ttn/lorawan/v3/lorawan.proto`](#ttn/lorawan/v3/lorawan.proto)
  - [Message `ADRAckDelayExponentValue`](#ttn.lorawan.v3.ADRAckDelayExponentValue)
  - [Message `ADRAckLimitExponentValue`](#ttn.lorawan.v3.ADRAckLimitExponentValue)
  - [Message `AggregatedDutyCycleValue`](#ttn.lorawan.v3.AggregatedDutyCycleValue)
  - [Message `CFList`](#ttn.lorawan.v3.CFList)
  - [Message `ClassBCGatewayIdentifiers`](#ttn.lorawan.v3.ClassBCGatewayIdentifiers)
  - [Message `DLSettings`](#ttn.lorawan.v3.DLSettings)
  - [Message `DataRate`](#ttn.lorawan.v3.DataRate)
  - [Message `DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue)
  - [Message `DataRateOffsetValue`](#ttn.lorawan.v3.DataRateOffsetValue)
  - [Message `DeviceEIRPValue`](#ttn.lorawan.v3.DeviceEIRPValue)
  - [Message `DownlinkPath`](#ttn.lorawan.v3.DownlinkPath)
  - [Message `FCtrl`](#ttn.lorawan.v3.FCtrl)
  - [Message `FHDR`](#ttn.lorawan.v3.FHDR)
  - [Message `FSKDataRate`](#ttn.lorawan.v3.FSKDataRate)
  - [Message `FrequencyValue`](#ttn.lorawan.v3.FrequencyValue)
  - [Message `GatewayAntennaIdentifiers`](#ttn.lorawan.v3.GatewayAntennaIdentifiers)
  - [Message `JoinAcceptPayload`](#ttn.lorawan.v3.JoinAcceptPayload)
  - [Message `JoinRequestPayload`](#ttn.lorawan.v3.JoinRequestPayload)
  - [Message `LRFHSSDataRate`](#ttn.lorawan.v3.LRFHSSDataRate)
  - [Message `LoRaDataRate`](#ttn.lorawan.v3.LoRaDataRate)
  - [Message `MACCommand`](#ttn.lorawan.v3.MACCommand)
  - [Message `MACCommand.ADRParamSetupReq`](#ttn.lorawan.v3.MACCommand.ADRParamSetupReq)
  - [Message `MACCommand.BeaconFreqAns`](#ttn.lorawan.v3.MACCommand.BeaconFreqAns)
  - [Message `MACCommand.BeaconFreqReq`](#ttn.lorawan.v3.MACCommand.BeaconFreqReq)
  - [Message `MACCommand.BeaconTimingAns`](#ttn.lorawan.v3.MACCommand.BeaconTimingAns)
  - [Message `MACCommand.DLChannelAns`](#ttn.lorawan.v3.MACCommand.DLChannelAns)
  - [Message `MACCommand.DLChannelReq`](#ttn.lorawan.v3.MACCommand.DLChannelReq)
  - [Message `MACCommand.DevStatusAns`](#ttn.lorawan.v3.MACCommand.DevStatusAns)
  - [Message `MACCommand.DeviceModeConf`](#ttn.lorawan.v3.MACCommand.DeviceModeConf)
  - [Message `MACCommand.DeviceModeInd`](#ttn.lorawan.v3.MACCommand.DeviceModeInd)
  - [Message `MACCommand.DeviceTimeAns`](#ttn.lorawan.v3.MACCommand.DeviceTimeAns)
  - [Message `MACCommand.DutyCycleReq`](#ttn.lorawan.v3.MACCommand.DutyCycleReq)
  - [Message `MACCommand.ForceRejoinReq`](#ttn.lorawan.v3.MACCommand.ForceRejoinReq)
  - [Message `MACCommand.LinkADRAns`](#ttn.lorawan.v3.MACCommand.LinkADRAns)
  - [Message `MACCommand.LinkADRReq`](#ttn.lorawan.v3.MACCommand.LinkADRReq)
  - [Message `MACCommand.LinkCheckAns`](#ttn.lorawan.v3.MACCommand.LinkCheckAns)
  - [Message `MACCommand.NewChannelAns`](#ttn.lorawan.v3.MACCommand.NewChannelAns)
  - [Message `MACCommand.NewChannelReq`](#ttn.lorawan.v3.MACCommand.NewChannelReq)
  - [Message `MACCommand.PingSlotChannelAns`](#ttn.lorawan.v3.MACCommand.PingSlotChannelAns)
  - [Message `MACCommand.PingSlotChannelReq`](#ttn.lorawan.v3.MACCommand.PingSlotChannelReq)
  - [Message `MACCommand.PingSlotInfoReq`](#ttn.lorawan.v3.MACCommand.PingSlotInfoReq)
  - [Message `MACCommand.RejoinParamSetupAns`](#ttn.lorawan.v3.MACCommand.RejoinParamSetupAns)
  - [Message `MACCommand.RejoinParamSetupReq`](#ttn.lorawan.v3.MACCommand.RejoinParamSetupReq)
  - [Message `MACCommand.RekeyConf`](#ttn.lorawan.v3.MACCommand.RekeyConf)
  - [Message `MACCommand.RekeyInd`](#ttn.lorawan.v3.MACCommand.RekeyInd)
  - [Message `MACCommand.RelayConfAns`](#ttn.lorawan.v3.MACCommand.RelayConfAns)
  - [Message `MACCommand.RelayConfReq`](#ttn.lorawan.v3.MACCommand.RelayConfReq)
  - [Message `MACCommand.RelayConfReq.Configuration`](#ttn.lorawan.v3.MACCommand.RelayConfReq.Configuration)
  - [Message `MACCommand.RelayConfigureFwdLimitAns`](#ttn.lorawan.v3.MACCommand.RelayConfigureFwdLimitAns)
  - [Message `MACCommand.RelayConfigureFwdLimitReq`](#ttn.lorawan.v3.MACCommand.RelayConfigureFwdLimitReq)
  - [Message `MACCommand.RelayCtrlUplinkListAns`](#ttn.lorawan.v3.MACCommand.RelayCtrlUplinkListAns)
  - [Message `MACCommand.RelayCtrlUplinkListReq`](#ttn.lorawan.v3.MACCommand.RelayCtrlUplinkListReq)
  - [Message `MACCommand.RelayEndDeviceConfAns`](#ttn.lorawan.v3.MACCommand.RelayEndDeviceConfAns)
  - [Message `MACCommand.RelayEndDeviceConfReq`](#ttn.lorawan.v3.MACCommand.RelayEndDeviceConfReq)
  - [Message `MACCommand.RelayEndDeviceConfReq.Configuration`](#ttn.lorawan.v3.MACCommand.RelayEndDeviceConfReq.Configuration)
  - [Message `MACCommand.RelayNotifyNewEndDeviceReq`](#ttn.lorawan.v3.MACCommand.RelayNotifyNewEndDeviceReq)
  - [Message `MACCommand.RelayUpdateUplinkListAns`](#ttn.lorawan.v3.MACCommand.RelayUpdateUplinkListAns)
  - [Message `MACCommand.RelayUpdateUplinkListReq`](#ttn.lorawan.v3.MACCommand.RelayUpdateUplinkListReq)
  - [Message `MACCommand.ResetConf`](#ttn.lorawan.v3.MACCommand.ResetConf)
  - [Message `MACCommand.ResetInd`](#ttn.lorawan.v3.MACCommand.ResetInd)
  - [Message `MACCommand.RxParamSetupAns`](#ttn.lorawan.v3.MACCommand.RxParamSetupAns)
  - [Message `MACCommand.RxParamSetupReq`](#ttn.lorawan.v3.MACCommand.RxParamSetupReq)
  - [Message `MACCommand.RxTimingSetupReq`](#ttn.lorawan.v3.MACCommand.RxTimingSetupReq)
  - [Message `MACCommand.TxParamSetupReq`](#ttn.lorawan.v3.MACCommand.TxParamSetupReq)
  - [Message `MACCommands`](#ttn.lorawan.v3.MACCommands)
  - [Message `MACPayload`](#ttn.lorawan.v3.MACPayload)
  - [Message `MHDR`](#ttn.lorawan.v3.MHDR)
  - [Message `Message`](#ttn.lorawan.v3.Message)
  - [Message `PingSlotPeriodValue`](#ttn.lorawan.v3.PingSlotPeriodValue)
  - [Message `RejoinRequestPayload`](#ttn.lorawan.v3.RejoinRequestPayload)
  - [Message `RelayEndDeviceAlwaysMode`](#ttn.lorawan.v3.RelayEndDeviceAlwaysMode)
  - [Message `RelayEndDeviceControlledMode`](#ttn.lorawan.v3.RelayEndDeviceControlledMode)
  - [Message `RelayEndDeviceDynamicMode`](#ttn.lorawan.v3.RelayEndDeviceDynamicMode)
  - [Message `RelayForwardDownlinkReq`](#ttn.lorawan.v3.RelayForwardDownlinkReq)
  - [Message `RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits)
  - [Message `RelayForwardUplinkReq`](#ttn.lorawan.v3.RelayForwardUplinkReq)
  - [Message `RelaySecondChannel`](#ttn.lorawan.v3.RelaySecondChannel)
  - [Message `RelayUplinkForwardLimits`](#ttn.lorawan.v3.RelayUplinkForwardLimits)
  - [Message `RelayUplinkToken`](#ttn.lorawan.v3.RelayUplinkToken)
  - [Message `RxDelayValue`](#ttn.lorawan.v3.RxDelayValue)
  - [Message `TxRequest`](#ttn.lorawan.v3.TxRequest)
  - [Message `TxSettings`](#ttn.lorawan.v3.TxSettings)
  - [Message `TxSettings.Downlink`](#ttn.lorawan.v3.TxSettings.Downlink)
  - [Message `UplinkToken`](#ttn.lorawan.v3.UplinkToken)
  - [Message `ZeroableFrequencyValue`](#ttn.lorawan.v3.ZeroableFrequencyValue)
  - [Enum `ADRAckDelayExponent`](#ttn.lorawan.v3.ADRAckDelayExponent)
  - [Enum `ADRAckLimitExponent`](#ttn.lorawan.v3.ADRAckLimitExponent)
  - [Enum `AggregatedDutyCycle`](#ttn.lorawan.v3.AggregatedDutyCycle)
  - [Enum `CFListType`](#ttn.lorawan.v3.CFListType)
  - [Enum `Class`](#ttn.lorawan.v3.Class)
  - [Enum `DataRateIndex`](#ttn.lorawan.v3.DataRateIndex)
  - [Enum `DataRateOffset`](#ttn.lorawan.v3.DataRateOffset)
  - [Enum `DeviceEIRP`](#ttn.lorawan.v3.DeviceEIRP)
  - [Enum `JoinRequestType`](#ttn.lorawan.v3.JoinRequestType)
  - [Enum `MACCommandIdentifier`](#ttn.lorawan.v3.MACCommandIdentifier)
  - [Enum `MACVersion`](#ttn.lorawan.v3.MACVersion)
  - [Enum `MType`](#ttn.lorawan.v3.MType)
  - [Enum `Major`](#ttn.lorawan.v3.Major)
  - [Enum `Minor`](#ttn.lorawan.v3.Minor)
  - [Enum `PHYVersion`](#ttn.lorawan.v3.PHYVersion)
  - [Enum `PingSlotPeriod`](#ttn.lorawan.v3.PingSlotPeriod)
  - [Enum `RejoinCountExponent`](#ttn.lorawan.v3.RejoinCountExponent)
  - [Enum `RejoinPeriodExponent`](#ttn.lorawan.v3.RejoinPeriodExponent)
  - [Enum `RejoinRequestType`](#ttn.lorawan.v3.RejoinRequestType)
  - [Enum `RejoinTimeExponent`](#ttn.lorawan.v3.RejoinTimeExponent)
  - [Enum `RelayCADPeriodicity`](#ttn.lorawan.v3.RelayCADPeriodicity)
  - [Enum `RelayCtrlUplinkListAction`](#ttn.lorawan.v3.RelayCtrlUplinkListAction)
  - [Enum `RelayLimitBucketSize`](#ttn.lorawan.v3.RelayLimitBucketSize)
  - [Enum `RelayResetLimitCounter`](#ttn.lorawan.v3.RelayResetLimitCounter)
  - [Enum `RelaySecondChAckOffset`](#ttn.lorawan.v3.RelaySecondChAckOffset)
  - [Enum `RelaySmartEnableLevel`](#ttn.lorawan.v3.RelaySmartEnableLevel)
  - [Enum `RelayWORChannel`](#ttn.lorawan.v3.RelayWORChannel)
  - [Enum `RxDelay`](#ttn.lorawan.v3.RxDelay)
  - [Enum `TxSchedulePriority`](#ttn.lorawan.v3.TxSchedulePriority)
- [File `ttn/lorawan/v3/messages.proto`](#ttn/lorawan/v3/messages.proto)
  - [Message `ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink)
  - [Message `ApplicationDownlink.ClassBC`](#ttn.lorawan.v3.ApplicationDownlink.ClassBC)
  - [Message `ApplicationDownlink.ConfirmedRetry`](#ttn.lorawan.v3.ApplicationDownlink.ConfirmedRetry)
  - [Message `ApplicationDownlinkFailed`](#ttn.lorawan.v3.ApplicationDownlinkFailed)
  - [Message `ApplicationDownlinks`](#ttn.lorawan.v3.ApplicationDownlinks)
  - [Message `ApplicationInvalidatedDownlinks`](#ttn.lorawan.v3.ApplicationInvalidatedDownlinks)
  - [Message `ApplicationJoinAccept`](#ttn.lorawan.v3.ApplicationJoinAccept)
  - [Message `ApplicationLocation`](#ttn.lorawan.v3.ApplicationLocation)
  - [Message `ApplicationLocation.AttributesEntry`](#ttn.lorawan.v3.ApplicationLocation.AttributesEntry)
  - [Message `ApplicationServiceData`](#ttn.lorawan.v3.ApplicationServiceData)
  - [Message `ApplicationUp`](#ttn.lorawan.v3.ApplicationUp)
  - [Message `ApplicationUplink`](#ttn.lorawan.v3.ApplicationUplink)
  - [Message `ApplicationUplink.LocationsEntry`](#ttn.lorawan.v3.ApplicationUplink.LocationsEntry)
  - [Message `ApplicationUplinkNormalized`](#ttn.lorawan.v3.ApplicationUplinkNormalized)
  - [Message `ApplicationUplinkNormalized.LocationsEntry`](#ttn.lorawan.v3.ApplicationUplinkNormalized.LocationsEntry)
  - [Message `DownlinkMessage`](#ttn.lorawan.v3.DownlinkMessage)
  - [Message `DownlinkQueueOperationErrorDetails`](#ttn.lorawan.v3.DownlinkQueueOperationErrorDetails)
  - [Message `DownlinkQueueRequest`](#ttn.lorawan.v3.DownlinkQueueRequest)
  - [Message `GatewayTxAcknowledgment`](#ttn.lorawan.v3.GatewayTxAcknowledgment)
  - [Message `GatewayUplinkMessage`](#ttn.lorawan.v3.GatewayUplinkMessage)
  - [Message `MessagePayloadFormatters`](#ttn.lorawan.v3.MessagePayloadFormatters)
  - [Message `TxAcknowledgment`](#ttn.lorawan.v3.TxAcknowledgment)
  - [Message `UplinkMessage`](#ttn.lorawan.v3.UplinkMessage)
  - [Enum `PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter)
  - [Enum `TxAcknowledgment.Result`](#ttn.lorawan.v3.TxAcknowledgment.Result)
- [File `ttn/lorawan/v3/metadata.proto`](#ttn/lorawan/v3/metadata.proto)
  - [Message `Location`](#ttn.lorawan.v3.Location)
  - [Message `PacketBrokerMetadata`](#ttn.lorawan.v3.PacketBrokerMetadata)
  - [Message `PacketBrokerRouteHop`](#ttn.lorawan.v3.PacketBrokerRouteHop)
  - [Message `RelayMetadata`](#ttn.lorawan.v3.RelayMetadata)
  - [Message `RxMetadata`](#ttn.lorawan.v3.RxMetadata)
  - [Enum `LocationSource`](#ttn.lorawan.v3.LocationSource)
- [File `ttn/lorawan/v3/mqtt.proto`](#ttn/lorawan/v3/mqtt.proto)
  - [Message `MQTTConnectionInfo`](#ttn.lorawan.v3.MQTTConnectionInfo)
- [File `ttn/lorawan/v3/networkserver.proto`](#ttn/lorawan/v3/networkserver.proto)
  - [Message `GenerateDevAddrResponse`](#ttn.lorawan.v3.GenerateDevAddrResponse)
  - [Message `GetDefaultMACSettingsRequest`](#ttn.lorawan.v3.GetDefaultMACSettingsRequest)
  - [Message `GetDeviceAdressPrefixesResponse`](#ttn.lorawan.v3.GetDeviceAdressPrefixesResponse)
  - [Message `GetNetIDResponse`](#ttn.lorawan.v3.GetNetIDResponse)
  - [Service `AsNs`](#ttn.lorawan.v3.AsNs)
  - [Service `GsNs`](#ttn.lorawan.v3.GsNs)
  - [Service `Ns`](#ttn.lorawan.v3.Ns)
  - [Service `NsEndDeviceBatchRegistry`](#ttn.lorawan.v3.NsEndDeviceBatchRegistry)
  - [Service `NsEndDeviceRegistry`](#ttn.lorawan.v3.NsEndDeviceRegistry)
- [File `ttn/lorawan/v3/notification_service.proto`](#ttn/lorawan/v3/notification_service.proto)
  - [Message `CreateNotificationRequest`](#ttn.lorawan.v3.CreateNotificationRequest)
  - [Message `CreateNotificationResponse`](#ttn.lorawan.v3.CreateNotificationResponse)
  - [Message `EntityStateChangedNotification`](#ttn.lorawan.v3.EntityStateChangedNotification)
  - [Message `ListNotificationsRequest`](#ttn.lorawan.v3.ListNotificationsRequest)
  - [Message `ListNotificationsResponse`](#ttn.lorawan.v3.ListNotificationsResponse)
  - [Message `Notification`](#ttn.lorawan.v3.Notification)
  - [Message `UpdateNotificationStatusRequest`](#ttn.lorawan.v3.UpdateNotificationStatusRequest)
  - [Enum `NotificationReceiver`](#ttn.lorawan.v3.NotificationReceiver)
  - [Enum `NotificationStatus`](#ttn.lorawan.v3.NotificationStatus)
  - [Service `NotificationService`](#ttn.lorawan.v3.NotificationService)
- [File `ttn/lorawan/v3/oauth.proto`](#ttn/lorawan/v3/oauth.proto)
  - [Message `ListOAuthAccessTokensRequest`](#ttn.lorawan.v3.ListOAuthAccessTokensRequest)
  - [Message `ListOAuthClientAuthorizationsRequest`](#ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest)
  - [Message `OAuthAccessToken`](#ttn.lorawan.v3.OAuthAccessToken)
  - [Message `OAuthAccessTokenIdentifiers`](#ttn.lorawan.v3.OAuthAccessTokenIdentifiers)
  - [Message `OAuthAccessTokens`](#ttn.lorawan.v3.OAuthAccessTokens)
  - [Message `OAuthAuthorizationCode`](#ttn.lorawan.v3.OAuthAuthorizationCode)
  - [Message `OAuthClientAuthorization`](#ttn.lorawan.v3.OAuthClientAuthorization)
  - [Message `OAuthClientAuthorizationIdentifiers`](#ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers)
  - [Message `OAuthClientAuthorizations`](#ttn.lorawan.v3.OAuthClientAuthorizations)
- [File `ttn/lorawan/v3/oauth_services.proto`](#ttn/lorawan/v3/oauth_services.proto)
  - [Service `OAuthAuthorizationRegistry`](#ttn.lorawan.v3.OAuthAuthorizationRegistry)
- [File `ttn/lorawan/v3/organization.proto`](#ttn/lorawan/v3/organization.proto)
  - [Message `CreateOrganizationAPIKeyRequest`](#ttn.lorawan.v3.CreateOrganizationAPIKeyRequest)
  - [Message `CreateOrganizationRequest`](#ttn.lorawan.v3.CreateOrganizationRequest)
  - [Message `DeleteOrganizationAPIKeyRequest`](#ttn.lorawan.v3.DeleteOrganizationAPIKeyRequest)
  - [Message `DeleteOrganizationCollaboratorRequest`](#ttn.lorawan.v3.DeleteOrganizationCollaboratorRequest)
  - [Message `GetOrganizationAPIKeyRequest`](#ttn.lorawan.v3.GetOrganizationAPIKeyRequest)
  - [Message `GetOrganizationCollaboratorRequest`](#ttn.lorawan.v3.GetOrganizationCollaboratorRequest)
  - [Message `GetOrganizationRequest`](#ttn.lorawan.v3.GetOrganizationRequest)
  - [Message `ListOrganizationAPIKeysRequest`](#ttn.lorawan.v3.ListOrganizationAPIKeysRequest)
  - [Message `ListOrganizationCollaboratorsRequest`](#ttn.lorawan.v3.ListOrganizationCollaboratorsRequest)
  - [Message `ListOrganizationsRequest`](#ttn.lorawan.v3.ListOrganizationsRequest)
  - [Message `Organization`](#ttn.lorawan.v3.Organization)
  - [Message `Organization.AttributesEntry`](#ttn.lorawan.v3.Organization.AttributesEntry)
  - [Message `Organizations`](#ttn.lorawan.v3.Organizations)
  - [Message `SetOrganizationCollaboratorRequest`](#ttn.lorawan.v3.SetOrganizationCollaboratorRequest)
  - [Message `UpdateOrganizationAPIKeyRequest`](#ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest)
  - [Message `UpdateOrganizationRequest`](#ttn.lorawan.v3.UpdateOrganizationRequest)
- [File `ttn/lorawan/v3/organization_services.proto`](#ttn/lorawan/v3/organization_services.proto)
  - [Service `OrganizationAccess`](#ttn.lorawan.v3.OrganizationAccess)
  - [Service `OrganizationRegistry`](#ttn.lorawan.v3.OrganizationRegistry)
- [File `ttn/lorawan/v3/packetbrokeragent.proto`](#ttn/lorawan/v3/packetbrokeragent.proto)
  - [Message `ListForwarderRoutingPoliciesRequest`](#ttn.lorawan.v3.ListForwarderRoutingPoliciesRequest)
  - [Message `ListHomeNetworkRoutingPoliciesRequest`](#ttn.lorawan.v3.ListHomeNetworkRoutingPoliciesRequest)
  - [Message `ListPacketBrokerHomeNetworksRequest`](#ttn.lorawan.v3.ListPacketBrokerHomeNetworksRequest)
  - [Message `ListPacketBrokerNetworksRequest`](#ttn.lorawan.v3.ListPacketBrokerNetworksRequest)
  - [Message `PacketBrokerAgentCompoundUplinkToken`](#ttn.lorawan.v3.PacketBrokerAgentCompoundUplinkToken)
  - [Message `PacketBrokerAgentEncryptedPayload`](#ttn.lorawan.v3.PacketBrokerAgentEncryptedPayload)
  - [Message `PacketBrokerAgentGatewayUplinkToken`](#ttn.lorawan.v3.PacketBrokerAgentGatewayUplinkToken)
  - [Message `PacketBrokerAgentUplinkToken`](#ttn.lorawan.v3.PacketBrokerAgentUplinkToken)
  - [Message `PacketBrokerDefaultGatewayVisibility`](#ttn.lorawan.v3.PacketBrokerDefaultGatewayVisibility)
  - [Message `PacketBrokerDefaultRoutingPolicy`](#ttn.lorawan.v3.PacketBrokerDefaultRoutingPolicy)
  - [Message `PacketBrokerDevAddrBlock`](#ttn.lorawan.v3.PacketBrokerDevAddrBlock)
  - [Message `PacketBrokerGateway`](#ttn.lorawan.v3.PacketBrokerGateway)
  - [Message `PacketBrokerGateway.GatewayIdentifiers`](#ttn.lorawan.v3.PacketBrokerGateway.GatewayIdentifiers)
  - [Message `PacketBrokerGatewayVisibility`](#ttn.lorawan.v3.PacketBrokerGatewayVisibility)
  - [Message `PacketBrokerInfo`](#ttn.lorawan.v3.PacketBrokerInfo)
  - [Message `PacketBrokerNetwork`](#ttn.lorawan.v3.PacketBrokerNetwork)
  - [Message `PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier)
  - [Message `PacketBrokerNetworks`](#ttn.lorawan.v3.PacketBrokerNetworks)
  - [Message `PacketBrokerRegisterRequest`](#ttn.lorawan.v3.PacketBrokerRegisterRequest)
  - [Message `PacketBrokerRoutingPolicies`](#ttn.lorawan.v3.PacketBrokerRoutingPolicies)
  - [Message `PacketBrokerRoutingPolicy`](#ttn.lorawan.v3.PacketBrokerRoutingPolicy)
  - [Message `PacketBrokerRoutingPolicyDownlink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyDownlink)
  - [Message `PacketBrokerRoutingPolicyUplink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyUplink)
  - [Message `SetPacketBrokerDefaultGatewayVisibilityRequest`](#ttn.lorawan.v3.SetPacketBrokerDefaultGatewayVisibilityRequest)
  - [Message `SetPacketBrokerDefaultRoutingPolicyRequest`](#ttn.lorawan.v3.SetPacketBrokerDefaultRoutingPolicyRequest)
  - [Message `SetPacketBrokerRoutingPolicyRequest`](#ttn.lorawan.v3.SetPacketBrokerRoutingPolicyRequest)
  - [Message `UpdatePacketBrokerGatewayRequest`](#ttn.lorawan.v3.UpdatePacketBrokerGatewayRequest)
  - [Message `UpdatePacketBrokerGatewayResponse`](#ttn.lorawan.v3.UpdatePacketBrokerGatewayResponse)
  - [Service `GsPba`](#ttn.lorawan.v3.GsPba)
  - [Service `NsPba`](#ttn.lorawan.v3.NsPba)
  - [Service `Pba`](#ttn.lorawan.v3.Pba)
- [File `ttn/lorawan/v3/picture.proto`](#ttn/lorawan/v3/picture.proto)
  - [Message `Picture`](#ttn.lorawan.v3.Picture)
  - [Message `Picture.Embedded`](#ttn.lorawan.v3.Picture.Embedded)
  - [Message `Picture.SizesEntry`](#ttn.lorawan.v3.Picture.SizesEntry)
- [File `ttn/lorawan/v3/qrcodegenerator.proto`](#ttn/lorawan/v3/qrcodegenerator.proto)
  - [Message `GenerateEndDeviceQRCodeRequest`](#ttn.lorawan.v3.GenerateEndDeviceQRCodeRequest)
  - [Message `GenerateEndDeviceQRCodeRequest.Image`](#ttn.lorawan.v3.GenerateEndDeviceQRCodeRequest.Image)
  - [Message `GenerateQRCodeResponse`](#ttn.lorawan.v3.GenerateQRCodeResponse)
  - [Message `GetQRCodeFormatRequest`](#ttn.lorawan.v3.GetQRCodeFormatRequest)
  - [Message `ParseEndDeviceQRCodeRequest`](#ttn.lorawan.v3.ParseEndDeviceQRCodeRequest)
  - [Message `ParseEndDeviceQRCodeResponse`](#ttn.lorawan.v3.ParseEndDeviceQRCodeResponse)
  - [Message `QRCodeFormat`](#ttn.lorawan.v3.QRCodeFormat)
  - [Message `QRCodeFormats`](#ttn.lorawan.v3.QRCodeFormats)
  - [Message `QRCodeFormats.FormatsEntry`](#ttn.lorawan.v3.QRCodeFormats.FormatsEntry)
  - [Service `EndDeviceQRCodeGenerator`](#ttn.lorawan.v3.EndDeviceQRCodeGenerator)
- [File `ttn/lorawan/v3/regional.proto`](#ttn/lorawan/v3/regional.proto)
  - [Message `ConcentratorConfig`](#ttn.lorawan.v3.ConcentratorConfig)
  - [Message `ConcentratorConfig.Channel`](#ttn.lorawan.v3.ConcentratorConfig.Channel)
  - [Message `ConcentratorConfig.FSKChannel`](#ttn.lorawan.v3.ConcentratorConfig.FSKChannel)
  - [Message `ConcentratorConfig.LBTConfiguration`](#ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration)
  - [Message `ConcentratorConfig.LoRaStandardChannel`](#ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel)
- [File `ttn/lorawan/v3/rights.proto`](#ttn/lorawan/v3/rights.proto)
  - [Message `APIKey`](#ttn.lorawan.v3.APIKey)
  - [Message `APIKeys`](#ttn.lorawan.v3.APIKeys)
  - [Message `Collaborator`](#ttn.lorawan.v3.Collaborator)
  - [Message `Collaborators`](#ttn.lorawan.v3.Collaborators)
  - [Message `GetCollaboratorResponse`](#ttn.lorawan.v3.GetCollaboratorResponse)
  - [Message `Rights`](#ttn.lorawan.v3.Rights)
  - [Enum `Right`](#ttn.lorawan.v3.Right)
- [File `ttn/lorawan/v3/search_services.proto`](#ttn/lorawan/v3/search_services.proto)
  - [Message `SearchAccountsRequest`](#ttn.lorawan.v3.SearchAccountsRequest)
  - [Message `SearchAccountsResponse`](#ttn.lorawan.v3.SearchAccountsResponse)
  - [Message `SearchApplicationsRequest`](#ttn.lorawan.v3.SearchApplicationsRequest)
  - [Message `SearchApplicationsRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchApplicationsRequest.AttributesContainEntry)
  - [Message `SearchClientsRequest`](#ttn.lorawan.v3.SearchClientsRequest)
  - [Message `SearchClientsRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchClientsRequest.AttributesContainEntry)
  - [Message `SearchEndDevicesRequest`](#ttn.lorawan.v3.SearchEndDevicesRequest)
  - [Message `SearchEndDevicesRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry)
  - [Message `SearchGatewaysRequest`](#ttn.lorawan.v3.SearchGatewaysRequest)
  - [Message `SearchGatewaysRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchGatewaysRequest.AttributesContainEntry)
  - [Message `SearchOrganizationsRequest`](#ttn.lorawan.v3.SearchOrganizationsRequest)
  - [Message `SearchOrganizationsRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchOrganizationsRequest.AttributesContainEntry)
  - [Message `SearchUsersRequest`](#ttn.lorawan.v3.SearchUsersRequest)
  - [Message `SearchUsersRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchUsersRequest.AttributesContainEntry)
  - [Service `EndDeviceRegistrySearch`](#ttn.lorawan.v3.EndDeviceRegistrySearch)
  - [Service `EntityRegistrySearch`](#ttn.lorawan.v3.EntityRegistrySearch)
- [File `ttn/lorawan/v3/secrets.proto`](#ttn/lorawan/v3/secrets.proto)
  - [Message `Secret`](#ttn.lorawan.v3.Secret)
- [File `ttn/lorawan/v3/simulate.proto`](#ttn/lorawan/v3/simulate.proto)
  - [Message `SimulateDataUplinkParams`](#ttn.lorawan.v3.SimulateDataUplinkParams)
  - [Message `SimulateJoinRequestParams`](#ttn.lorawan.v3.SimulateJoinRequestParams)
  - [Message `SimulateMetadataParams`](#ttn.lorawan.v3.SimulateMetadataParams)
- [File `ttn/lorawan/v3/user.proto`](#ttn/lorawan/v3/user.proto)
  - [Message `CreateLoginTokenRequest`](#ttn.lorawan.v3.CreateLoginTokenRequest)
  - [Message `CreateLoginTokenResponse`](#ttn.lorawan.v3.CreateLoginTokenResponse)
  - [Message `CreateTemporaryPasswordRequest`](#ttn.lorawan.v3.CreateTemporaryPasswordRequest)
  - [Message `CreateUserAPIKeyRequest`](#ttn.lorawan.v3.CreateUserAPIKeyRequest)
  - [Message `CreateUserRequest`](#ttn.lorawan.v3.CreateUserRequest)
  - [Message `DeleteInvitationRequest`](#ttn.lorawan.v3.DeleteInvitationRequest)
  - [Message `DeleteUserAPIKeyRequest`](#ttn.lorawan.v3.DeleteUserAPIKeyRequest)
  - [Message `GetUserAPIKeyRequest`](#ttn.lorawan.v3.GetUserAPIKeyRequest)
  - [Message `GetUserRequest`](#ttn.lorawan.v3.GetUserRequest)
  - [Message `Invitation`](#ttn.lorawan.v3.Invitation)
  - [Message `Invitations`](#ttn.lorawan.v3.Invitations)
  - [Message `ListInvitationsRequest`](#ttn.lorawan.v3.ListInvitationsRequest)
  - [Message `ListUserAPIKeysRequest`](#ttn.lorawan.v3.ListUserAPIKeysRequest)
  - [Message `ListUserSessionsRequest`](#ttn.lorawan.v3.ListUserSessionsRequest)
  - [Message `ListUsersRequest`](#ttn.lorawan.v3.ListUsersRequest)
  - [Message `LoginToken`](#ttn.lorawan.v3.LoginToken)
  - [Message `SendInvitationRequest`](#ttn.lorawan.v3.SendInvitationRequest)
  - [Message `UpdateUserAPIKeyRequest`](#ttn.lorawan.v3.UpdateUserAPIKeyRequest)
  - [Message `UpdateUserPasswordRequest`](#ttn.lorawan.v3.UpdateUserPasswordRequest)
  - [Message `UpdateUserRequest`](#ttn.lorawan.v3.UpdateUserRequest)
  - [Message `User`](#ttn.lorawan.v3.User)
  - [Message `User.AttributesEntry`](#ttn.lorawan.v3.User.AttributesEntry)
  - [Message `UserSession`](#ttn.lorawan.v3.UserSession)
  - [Message `UserSessionIdentifiers`](#ttn.lorawan.v3.UserSessionIdentifiers)
  - [Message `UserSessions`](#ttn.lorawan.v3.UserSessions)
  - [Message `Users`](#ttn.lorawan.v3.Users)
- [File `ttn/lorawan/v3/user_services.proto`](#ttn/lorawan/v3/user_services.proto)
  - [Service `UserAccess`](#ttn.lorawan.v3.UserAccess)
  - [Service `UserInvitationRegistry`](#ttn.lorawan.v3.UserInvitationRegistry)
  - [Service `UserRegistry`](#ttn.lorawan.v3.UserRegistry)
  - [Service `UserSessionRegistry`](#ttn.lorawan.v3.UserSessionRegistry)
- [Scalar Value Types](#scalar-value-types)

## <a name="ttn/lorawan/v3/_api.proto">File `ttn/lorawan/v3/_api.proto`</a>

## <a name="ttn/lorawan/v3/application.proto">File `ttn/lorawan/v3/application.proto`</a>

### <a name="ttn.lorawan.v3.Application">Message `Application`</a>

Application is the message that defines an Application in the network.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | The identifiers of the application. These are public and can be seen by any authenticated user in the network. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the application was created. This information is public and can be seen by any authenticated user in the network. |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the application was last updated. This information is public and can be seen by any authenticated user in the network. |
| `deleted_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the application was deleted. This information is public and can be seen by any authenticated user in the network. |
| `name` | [`string`](#string) |  | The name of the application. |
| `description` | [`string`](#string) |  | A description for the application. |
| `attributes` | [`Application.AttributesEntry`](#ttn.lorawan.v3.Application.AttributesEntry) | repeated | Key-value attributes for this application. Typically used for organizing applications or for storing integration-specific data. |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated | Contact information for this application. Typically used to indicate who to contact with technical/security questions about the application. This field is deprecated. Use administrative_contact and technical_contact instead. |
| `administrative_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `technical_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `network_server_address` | [`string`](#string) |  | The address of the Network Server where this application is supposed to be registered. If set, this fields indicates where end devices for this application should be registered. Stored in Entity Registry. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| `application_server_address` | [`string`](#string) |  | The address of the Application Server where this application is supposed to be registered. If set, this fields indicates where end devices for this application should be registered. Stored in Entity Registry. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| `join_server_address` | [`string`](#string) |  | The address of the Join Server where this application is supposed to be registered. If set, this fields indicates where end devices for this application should be registered. Stored in Entity Registry. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| `dev_eui_counter` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `description` | <p>`string.max_len`: `2000`</p> |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |
| `contact_info` | <p>`repeated.max_items`: `10`</p> |
| `network_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `application_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `join_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |

### <a name="ttn.lorawan.v3.Application.AttributesEntry">Message `Application.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.Applications">Message `Applications`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `applications` | [`Application`](#ttn.lorawan.v3.Application) | repeated |  |

### <a name="ttn.lorawan.v3.CreateApplicationAPIKeyRequest">Message `CreateApplicationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `name` | [`string`](#string) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `rights` | <p>`repeated.min_items`: `1`</p><p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `expires_at` | <p>`timestamp.gt_now`: `true`</p> |

### <a name="ttn.lorawan.v3.CreateApplicationRequest">Message `CreateApplicationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application` | [`Application`](#ttn.lorawan.v3.Application) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created application. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteApplicationAPIKeyRequest">Message `DeleteApplicationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteApplicationCollaboratorRequest">Message `DeleteApplicationCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `collaborator_ids` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `collaborator_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetApplicationAPIKeyRequest">Message `GetApplicationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  | Unique public identifier for the API key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetApplicationCollaboratorRequest">Message `GetApplicationCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetApplicationRequest">Message `GetApplicationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the application fields that should be returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.IssueDevEUIResponse">Message `IssueDevEUIResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.ListApplicationAPIKeysRequest">Message `ListApplicationAPIKeysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path. Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `order` | <p>`string.in`: `[ api_key_id -api_key_id name -name created_at -created_at expires_at -expires_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListApplicationCollaboratorsRequest">Message `ListApplicationCollaboratorsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `order` | <p>`string.in`: `[ id -id -rights rights]`</p> |

### <a name="ttn.lorawan.v3.ListApplicationsRequest">Message `ListApplicationsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | By default we list all applications the caller has rights on. Set the user or the organization (not both) to instead list the applications where the user or organization is collaborator on. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the application fields that should be returned. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted applications. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `order` | <p>`string.in`: `[ application_id -application_id name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SetApplicationCollaboratorRequest">Message `SetApplicationCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `collaborator` | [`Collaborator`](#ttn.lorawan.v3.Collaborator) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateApplicationAPIKeyRequest">Message `UpdateApplicationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `api_key` | [`APIKey`](#ttn.lorawan.v3.APIKey) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the api key fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `api_key` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateApplicationRequest">Message `UpdateApplicationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application` | [`Application`](#ttn.lorawan.v3.Application) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the application fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application` | <p>`message.required`: `true`</p> |

## <a name="ttn/lorawan/v3/application_services.proto">File `ttn/lorawan/v3/application_services.proto`</a>

### <a name="ttn.lorawan.v3.ApplicationAccess">Service `ApplicationAccess`</a>

The ApplicationAcces service, exposed by the Identity Server, is used to manage
API keys and collaborators of applications.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListRights` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`Rights`](#ttn.lorawan.v3.Rights) | List the rights the caller has on this application. |
| `CreateAPIKey` | [`CreateApplicationAPIKeyRequest`](#ttn.lorawan.v3.CreateApplicationAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Create an API key scoped to this application. |
| `ListAPIKeys` | [`ListApplicationAPIKeysRequest`](#ttn.lorawan.v3.ListApplicationAPIKeysRequest) | [`APIKeys`](#ttn.lorawan.v3.APIKeys) | List the API keys for this application. |
| `GetAPIKey` | [`GetApplicationAPIKeyRequest`](#ttn.lorawan.v3.GetApplicationAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Get a single API key of this application. |
| `UpdateAPIKey` | [`UpdateApplicationAPIKeyRequest`](#ttn.lorawan.v3.UpdateApplicationAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Update the rights of an API key of the application. This method can also be used to delete the API key, by giving it no rights. The caller is required to have all assigned or/and removed rights. |
| `DeleteAPIKey` | [`DeleteApplicationAPIKeyRequest`](#ttn.lorawan.v3.DeleteApplicationAPIKeyRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a single API key of this application. |
| `GetCollaborator` | [`GetApplicationCollaboratorRequest`](#ttn.lorawan.v3.GetApplicationCollaboratorRequest) | [`GetCollaboratorResponse`](#ttn.lorawan.v3.GetCollaboratorResponse) | Get the rights of a collaborator (member) of the application. Pseudo-rights in the response (such as the "_ALL" right) are not expanded. |
| `SetCollaborator` | [`SetApplicationCollaboratorRequest`](#ttn.lorawan.v3.SetApplicationCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Set the rights of a collaborator (member) on the application. This method can also be used to delete the collaborator, by giving them no rights. The caller is required to have all assigned or/and removed rights. |
| `ListCollaborators` | [`ListApplicationCollaboratorsRequest`](#ttn.lorawan.v3.ListApplicationCollaboratorsRequest) | [`Collaborators`](#ttn.lorawan.v3.Collaborators) | List the collaborators on this application. |
| `DeleteCollaborator` | [`DeleteApplicationCollaboratorRequest`](#ttn.lorawan.v3.DeleteApplicationCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | DeleteCollaborator removes a collaborator from an application. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListRights` | `GET` | `/api/v3/applications/{application_id}/rights` |  |
| `CreateAPIKey` | `POST` | `/api/v3/applications/{application_ids.application_id}/api-keys` | `*` |
| `ListAPIKeys` | `GET` | `/api/v3/applications/{application_ids.application_id}/api-keys` |  |
| `GetAPIKey` | `GET` | `/api/v3/applications/{application_ids.application_id}/api-keys/{key_id}` |  |
| `UpdateAPIKey` | `PUT` | `/api/v3/applications/{application_ids.application_id}/api-keys/{api_key.id}` | `*` |
| `DeleteAPIKey` | `DELETE` | `/api/v3/applications/{application_ids.application_id}/api-keys/{key_id}` |  |
| `GetCollaborator` | `` | `/api/v3` |  |
| `GetCollaborator` | `GET` | `/api/v3/applications/{application_ids.application_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
| `GetCollaborator` | `GET` | `/api/v3/applications/{application_ids.application_id}/collaborator/organization/{collaborator.organization_ids.organization_id}` |  |
| `SetCollaborator` | `PUT` | `/api/v3/applications/{application_ids.application_id}/collaborators` | `*` |
| `ListCollaborators` | `GET` | `/api/v3/applications/{application_ids.application_id}/collaborators` |  |
| `DeleteCollaborator` | `` | `/api/v3` |  |
| `DeleteCollaborator` | `DELETE` | `/api/v3/applications/{application_ids.application_id}/collaborator/user/{collaborator_ids.user_ids.user_id}` |  |
| `DeleteCollaborator` | `DELETE` | `/api/v3/applications/{application_ids.application_id}/collaborator/organization/{collaborator_ids.organization_ids.organization_id}` |  |

### <a name="ttn.lorawan.v3.ApplicationRegistry">Service `ApplicationRegistry`</a>

The ApplicationRegistry service, exposed by the Identity Server, is used to manage
application registrations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateApplicationRequest`](#ttn.lorawan.v3.CreateApplicationRequest) | [`Application`](#ttn.lorawan.v3.Application) | Create a new application. This also sets the given organization or user as first collaborator with all possible rights. |
| `Get` | [`GetApplicationRequest`](#ttn.lorawan.v3.GetApplicationRequest) | [`Application`](#ttn.lorawan.v3.Application) | Get the application with the given identifiers, selecting the fields specified in the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `List` | [`ListApplicationsRequest`](#ttn.lorawan.v3.ListApplicationsRequest) | [`Applications`](#ttn.lorawan.v3.Applications) | List applications where the given user or organization is a direct collaborator. If no user or organization is given, this returns the applications the caller has access to. Similar to Get, this selects the fields given by the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `Update` | [`UpdateApplicationRequest`](#ttn.lorawan.v3.UpdateApplicationRequest) | [`Application`](#ttn.lorawan.v3.Application) | Update the application, changing the fields specified by the field mask to the provided values. |
| `Delete` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the application. This may not release the application ID for reuse. All end devices must be deleted from the application before it can be deleted. |
| `Restore` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Restore a recently deleted application. Deployment configuration may specify if, and for how long after deletion, entities can be restored. |
| `Purge` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Purge the application. This will release the application ID for reuse. All end devices must be deleted from the application before it can be deleted. The application owner is responsible for clearing data from any (external) integrations that may store and expose data by application ID |
| `IssueDevEUI` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`IssueDevEUIResponse`](#ttn.lorawan.v3.IssueDevEUIResponse) | Request DevEUI from the configured address block for a device inside the application. The maximum number of DevEUI's issued per application can be configured. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Create` | `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/applications` | `*` |
| `Create` | `POST` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/applications` | `*` |
| `Get` | `GET` | `/api/v3/applications/{application_ids.application_id}` |  |
| `List` | `GET` | `/api/v3/applications` |  |
| `List` | `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/applications` |  |
| `List` | `GET` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/applications` |  |
| `Update` | `PUT` | `/api/v3/applications/{application.ids.application_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/applications/{application_id}` |  |
| `Restore` | `POST` | `/api/v3/applications/{application_id}/restore` |  |
| `Purge` | `DELETE` | `/api/v3/applications/{application_id}/purge` |  |
| `IssueDevEUI` | `POST` | `/api/v3/applications/{application_id}/dev-eui` |  |

## <a name="ttn/lorawan/v3/applicationserver.proto">File `ttn/lorawan/v3/applicationserver.proto`</a>

### <a name="ttn.lorawan.v3.ApplicationLink">Message `ApplicationLink`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default_formatters` | [`MessagePayloadFormatters`](#ttn.lorawan.v3.MessagePayloadFormatters) |  | Default message payload formatters to use when there are no formatters defined on the end device level. |
| `skip_payload_crypto` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  | Skip decryption of uplink payloads and encryption of downlink payloads. Leave empty for the using the Application Server's default setting. |

### <a name="ttn.lorawan.v3.ApplicationLinkStats">Message `ApplicationLinkStats`</a>

Link stats as monitored by the Application Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `linked_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `network_server_address` | [`string`](#string) |  |  |
| `last_up_received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp when the last upstream message has been received from a Network Server. This can be a join-accept, uplink message or downlink message event. |
| `up_count` | [`uint64`](#uint64) |  | Number of upstream messages received. |
| `last_downlink_forwarded_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp when the last downlink message has been forwarded to a Network Server. |
| `downlink_count` | [`uint64`](#uint64) |  | Number of downlink messages forwarded. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `network_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |

### <a name="ttn.lorawan.v3.AsConfiguration">Message `AsConfiguration`</a>

Application Server configuration.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pubsub` | [`AsConfiguration.PubSub`](#ttn.lorawan.v3.AsConfiguration.PubSub) |  |  |
| `webhooks` | [`AsConfiguration.Webhooks`](#ttn.lorawan.v3.AsConfiguration.Webhooks) |  |  |

### <a name="ttn.lorawan.v3.AsConfiguration.PubSub">Message `AsConfiguration.PubSub`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `providers` | [`AsConfiguration.PubSub.Providers`](#ttn.lorawan.v3.AsConfiguration.PubSub.Providers) |  |  |

### <a name="ttn.lorawan.v3.AsConfiguration.PubSub.Providers">Message `AsConfiguration.PubSub.Providers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mqtt` | [`AsConfiguration.PubSub.Providers.Status`](#ttn.lorawan.v3.AsConfiguration.PubSub.Providers.Status) |  |  |
| `nats` | [`AsConfiguration.PubSub.Providers.Status`](#ttn.lorawan.v3.AsConfiguration.PubSub.Providers.Status) |  |  |

### <a name="ttn.lorawan.v3.AsConfiguration.Webhooks">Message `AsConfiguration.Webhooks`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `unhealthy_attempts_threshold` | [`int64`](#int64) |  |  |
| `unhealthy_retry_interval` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |

### <a name="ttn.lorawan.v3.DecodeDownlinkRequest">Message `DecodeDownlinkRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  |  |
| `downlink` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| `formatter` | [`PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter) |  |  |
| `parameter` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |
| `downlink` | <p>`message.required`: `true`</p> |
| `formatter` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.DecodeDownlinkResponse">Message `DecodeDownlinkResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `downlink` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |

### <a name="ttn.lorawan.v3.DecodeUplinkRequest">Message `DecodeUplinkRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  |  |
| `uplink` | [`ApplicationUplink`](#ttn.lorawan.v3.ApplicationUplink) |  |  |
| `formatter` | [`PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter) |  |  |
| `parameter` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |
| `uplink` | <p>`message.required`: `true`</p> |
| `formatter` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.DecodeUplinkResponse">Message `DecodeUplinkResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uplink` | [`ApplicationUplink`](#ttn.lorawan.v3.ApplicationUplink) |  |  |

### <a name="ttn.lorawan.v3.EncodeDownlinkRequest">Message `EncodeDownlinkRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  |  |
| `downlink` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| `formatter` | [`PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter) |  |  |
| `parameter` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |
| `downlink` | <p>`message.required`: `true`</p> |
| `formatter` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.EncodeDownlinkResponse">Message `EncodeDownlinkResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `downlink` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |

### <a name="ttn.lorawan.v3.GetApplicationLinkRequest">Message `GetApplicationLinkRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetAsConfigurationRequest">Message `GetAsConfigurationRequest`</a>

### <a name="ttn.lorawan.v3.GetAsConfigurationResponse">Message `GetAsConfigurationResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `configuration` | [`AsConfiguration`](#ttn.lorawan.v3.AsConfiguration) |  |  |

### <a name="ttn.lorawan.v3.NsAsHandleUplinkRequest">Message `NsAsHandleUplinkRequest`</a>

Container for multiple Application uplink messages.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ups` | [`ApplicationUp`](#ttn.lorawan.v3.ApplicationUp) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ups` | <p>`repeated.min_items`: `1`</p> |

### <a name="ttn.lorawan.v3.SetApplicationLinkRequest">Message `SetApplicationLinkRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `link` | [`ApplicationLink`](#ttn.lorawan.v3.ApplicationLink) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `link` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.AsConfiguration.PubSub.Providers.Status">Enum `AsConfiguration.PubSub.Providers.Status`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `ENABLED` | 0 | No restrictions are in place. |
| `WARNING` | 1 | Warnings are being emitted that the provider will be deprecated in the future. |
| `DISABLED` | 2 | New integrations cannot be set up, and old ones do not start. |

### <a name="ttn.lorawan.v3.AppAs">Service `AppAs`</a>

The AppAs service connects an application or integration to an Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Subscribe` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`ApplicationUp`](#ttn.lorawan.v3.ApplicationUp) _stream_ | Subscribe to upstream messages. |
| `DownlinkQueuePush` | [`DownlinkQueueRequest`](#ttn.lorawan.v3.DownlinkQueueRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Push downlink messages to the end of the downlink queue. |
| `DownlinkQueueReplace` | [`DownlinkQueueRequest`](#ttn.lorawan.v3.DownlinkQueueRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Replace the entire downlink queue with the specified messages. This can also be used to empty the queue by specifying no messages. |
| `DownlinkQueueList` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`ApplicationDownlinks`](#ttn.lorawan.v3.ApplicationDownlinks) | List the items currently in the downlink queue. |
| `GetMQTTConnectionInfo` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`MQTTConnectionInfo`](#ttn.lorawan.v3.MQTTConnectionInfo) | Get connection information to connect an MQTT client. |
| `SimulateUplink` | [`ApplicationUp`](#ttn.lorawan.v3.ApplicationUp) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Simulate an upstream message. This can be used to test integrations. |
| `EncodeDownlink` | [`EncodeDownlinkRequest`](#ttn.lorawan.v3.EncodeDownlinkRequest) | [`EncodeDownlinkResponse`](#ttn.lorawan.v3.EncodeDownlinkResponse) |  |
| `DecodeUplink` | [`DecodeUplinkRequest`](#ttn.lorawan.v3.DecodeUplinkRequest) | [`DecodeUplinkResponse`](#ttn.lorawan.v3.DecodeUplinkResponse) |  |
| `DecodeDownlink` | [`DecodeDownlinkRequest`](#ttn.lorawan.v3.DecodeDownlinkRequest) | [`DecodeDownlinkResponse`](#ttn.lorawan.v3.DecodeDownlinkResponse) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `DownlinkQueuePush` | `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/down/push` | `*` |
| `DownlinkQueueReplace` | `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/down/replace` | `*` |
| `DownlinkQueueList` | `GET` | `/api/v3/as/applications/{application_ids.application_id}/devices/{device_id}/down` |  |
| `GetMQTTConnectionInfo` | `GET` | `/api/v3/as/applications/{application_id}/mqtt-connection-info` |  |
| `SimulateUplink` | `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/up/simulate` | `*` |
| `EncodeDownlink` | `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/down/encode` | `*` |
| `DecodeUplink` | `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/up/decode` | `*` |
| `DecodeDownlink` | `POST` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/down/decode` | `*` |

### <a name="ttn.lorawan.v3.As">Service `As`</a>

The As service manages the Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetLink` | [`GetApplicationLinkRequest`](#ttn.lorawan.v3.GetApplicationLinkRequest) | [`ApplicationLink`](#ttn.lorawan.v3.ApplicationLink) | Get a link configuration from the Application Server to Network Server. This only contains the configuration. Use GetLinkStats to view statistics and any link errors. |
| `SetLink` | [`SetApplicationLinkRequest`](#ttn.lorawan.v3.SetApplicationLinkRequest) | [`ApplicationLink`](#ttn.lorawan.v3.ApplicationLink) | Set a link configuration from the Application Server a Network Server. This call returns immediately after setting the link configuration; it does not wait for a link to establish. To get link statistics or errors, use GetLinkStats. Note that there can only be one Application Server instance linked to a Network Server for a given application at a time. |
| `DeleteLink` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the link between the Application Server and Network Server for the specified application. |
| `GetLinkStats` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`ApplicationLinkStats`](#ttn.lorawan.v3.ApplicationLinkStats) | GetLinkStats returns the link statistics. This call returns a NotFound error code if there is no link for the given application identifiers. This call returns the error code of the link error if linking to a Network Server failed. |
| `GetConfiguration` | [`GetAsConfigurationRequest`](#ttn.lorawan.v3.GetAsConfigurationRequest) | [`GetAsConfigurationResponse`](#ttn.lorawan.v3.GetAsConfigurationResponse) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetLink` | `GET` | `/api/v3/as/applications/{application_ids.application_id}/link` |  |
| `SetLink` | `PUT` | `/api/v3/as/applications/{application_ids.application_id}/link` | `*` |
| `DeleteLink` | `DELETE` | `/api/v3/as/applications/{application_id}/link` |  |
| `GetLinkStats` | `GET` | `/api/v3/as/applications/{application_id}/link/stats` |  |
| `GetConfiguration` | `GET` | `/api/v3/as/configuration` |  |

### <a name="ttn.lorawan.v3.AsEndDeviceBatchRegistry">Service `AsEndDeviceBatchRegistry`</a>

The AsEndDeviceBatchRegistry service allows clients to manage batches end devices on the Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Delete` | [`BatchDeleteEndDevicesRequest`](#ttn.lorawan.v3.BatchDeleteEndDevicesRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a list of devices within the same application. This operation is atomic; either all devices are deleted or none. Devices not found are skipped and no error is returned. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Delete` | `DELETE` | `/api/v3/as/applications/{application_ids.application_id}/devices/batch` |  |

### <a name="ttn.lorawan.v3.AsEndDeviceRegistry">Service `AsEndDeviceRegistry`</a>

The AsEndDeviceRegistry service allows clients to manage their end devices on the Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Get` | [`GetEndDeviceRequest`](#ttn.lorawan.v3.GetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Get returns the device that matches the given identifiers. If there are multiple matches, an error will be returned. |
| `Set` | [`SetEndDeviceRequest`](#ttn.lorawan.v3.SetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Set creates or updates the device. |
| `Delete` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete deletes the device that matches the given identifiers. If there are multiple matches, an error will be returned. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Get` | `GET` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |
| `Set` | `PUT` | `/api/v3/as/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | `*` |
| `Set` | `POST` | `/api/v3/as/applications/{end_device.ids.application_ids.application_id}/devices` | `*` |
| `Delete` | `DELETE` | `/api/v3/as/applications/{application_ids.application_id}/devices/{device_id}` |  |

### <a name="ttn.lorawan.v3.NsAs">Service `NsAs`</a>

The NsAs service connects a Network Server to an Application Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `HandleUplink` | [`NsAsHandleUplinkRequest`](#ttn.lorawan.v3.NsAsHandleUplinkRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Handle Application uplink messages. |

## <a name="ttn/lorawan/v3/applicationserver_integrations_alcsync.proto">File `ttn/lorawan/v3/applicationserver_integrations_alcsync.proto`</a>

### <a name="ttn.lorawan.v3.ALCSyncCommand">Message `ALCSyncCommand`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cid` | [`ALCSyncCommandIdentifier`](#ttn.lorawan.v3.ALCSyncCommandIdentifier) |  |  |
| `app_time_req` | [`ALCSyncCommand.AppTimeReq`](#ttn.lorawan.v3.ALCSyncCommand.AppTimeReq) |  |  |
| `app_time_ans` | [`ALCSyncCommand.AppTimeAns`](#ttn.lorawan.v3.ALCSyncCommand.AppTimeAns) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `cid` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.ALCSyncCommand.AppTimeAns">Message `ALCSyncCommand.AppTimeAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `TimeCorrection` | [`int32`](#int32) |  |  |
| `TokenAns` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `TokenAns` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.ALCSyncCommand.AppTimeReq">Message `ALCSyncCommand.AppTimeReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `DeviceTime` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `TokenReq` | [`uint32`](#uint32) |  |  |
| `AnsRequired` | [`bool`](#bool) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `DeviceTime` | <p>`timestamp.required`: `true`</p> |
| `TokenReq` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.ALCSyncCommandIdentifier">Enum `ALCSyncCommandIdentifier`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `ALCSYNC_CID_PKG_VERSION` | 0 |  |
| `ALCSYNC_CID_APP_TIME` | 1 |  |
| `ALCSYNC_CID_APP_DEV_TIME_PERIODICITY` | 2 |  |
| `ALCSYNC_CID_FORCE_DEV_RESYNC` | 3 |  |

## <a name="ttn/lorawan/v3/applicationserver_integrations_storage.proto">File `ttn/lorawan/v3/applicationserver_integrations_storage.proto`</a>

### <a name="ttn.lorawan.v3.ContinuationTokenPayload">Message `ContinuationTokenPayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `after` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `before` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `f_port` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `order` | [`string`](#string) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `last` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `last_received_id` | [`int64`](#int64) |  |  |

### <a name="ttn.lorawan.v3.GetStoredApplicationUpCountRequest">Message `GetStoredApplicationUpCountRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Count upstream messages from all end devices of an application. Cannot be used in conjunction with end_device_ids. |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  | Count upstream messages from a single end device. Cannot be used in conjunction with application_ids. |
| `type` | [`string`](#string) |  | Count upstream messages of a specific type. If not set, then all upstream messages are returned. |
| `after` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Count upstream messages after this timestamp only. Cannot be used in conjunction with last. |
| `before` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Count upstream messages before this timestamp only. Cannot be used in conjunction with last. |
| `f_port` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Count uplinks on a specific FPort only. |
| `last` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Count upstream messages that have arrived in the last minutes or hours. Cannot be used in conjunction with after and before. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `type` | <p>`string.in`: `[ uplink_message join_accept downlink_ack downlink_nack downlink_sent downlink_failed downlink_queued downlink_queue_invalidated location_solved service_data]`</p> |

### <a name="ttn.lorawan.v3.GetStoredApplicationUpCountResponse">Message `GetStoredApplicationUpCountResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `count` | [`GetStoredApplicationUpCountResponse.CountEntry`](#ttn.lorawan.v3.GetStoredApplicationUpCountResponse.CountEntry) | repeated | Number of stored messages by end device ID. |

### <a name="ttn.lorawan.v3.GetStoredApplicationUpCountResponse.CountEntry">Message `GetStoredApplicationUpCountResponse.CountEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`uint32`](#uint32) |  |  |

### <a name="ttn.lorawan.v3.GetStoredApplicationUpRequest">Message `GetStoredApplicationUpRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Query upstream messages from all end devices of an application. Cannot be used in conjunction with end_device_ids. |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  | Query upstream messages from a single end device. Cannot be used in conjunction with application_ids. |
| `type` | [`string`](#string) |  | Query upstream messages of a specific type. If not set, then all upstream messages are returned. |
| `limit` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Limit number of results. |
| `after` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Query upstream messages after this timestamp only. Cannot be used in conjunction with last. |
| `before` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Query upstream messages before this timestamp only. Cannot be used in conjunction with last. |
| `f_port` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Query uplinks on a specific FPort only. |
| `order` | [`string`](#string) |  | Order results. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the upstream message fields that should be returned. See the API reference for allowed field names for each type of upstream message. |
| `last` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Query upstream messages that have arrived in the last minutes or hours. Cannot be used in conjunction with after and before. |
| `continuation_token` | [`string`](#string) |  | The continuation token, which is used to retrieve the next page. If provided, other fields are ignored. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `type` | <p>`string.in`: `[ uplink_message uplink_normalized join_accept downlink_ack downlink_nack downlink_sent downlink_failed downlink_queued downlink_queue_invalidated location_solved service_data]`</p> |
| `order` | <p>`string.in`: `[ -received_at received_at]`</p> |
| `continuation_token` | <p>`string.max_len`: `16000`</p> |

### <a name="ttn.lorawan.v3.ApplicationUpStorage">Service `ApplicationUpStorage`</a>

The ApplicationUpStorage service can be used to query stored application upstream messages.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetStoredApplicationUp` | [`GetStoredApplicationUpRequest`](#ttn.lorawan.v3.GetStoredApplicationUpRequest) | [`ApplicationUp`](#ttn.lorawan.v3.ApplicationUp) _stream_ | Returns a stream of application messages that have been stored in the database. |
| `GetStoredApplicationUpCount` | [`GetStoredApplicationUpCountRequest`](#ttn.lorawan.v3.GetStoredApplicationUpCountRequest) | [`GetStoredApplicationUpCountResponse`](#ttn.lorawan.v3.GetStoredApplicationUpCountResponse) | Returns how many application messages have been stored in the database for an application or end device. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetStoredApplicationUp` | `GET` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/packages/storage/{type}` |  |
| `GetStoredApplicationUp` | `GET` | `/api/v3/as/applications/{application_ids.application_id}/packages/storage/{type}` |  |
| `GetStoredApplicationUpCount` | `GET` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/packages/storage/{type}/count` |  |
| `GetStoredApplicationUpCount` | `GET` | `/api/v3/as/applications/{application_ids.application_id}/packages/storage/{type}/count` |  |

## <a name="ttn/lorawan/v3/applicationserver_packages.proto">File `ttn/lorawan/v3/applicationserver_packages.proto`</a>

### <a name="ttn.lorawan.v3.ApplicationPackage">Message `ApplicationPackage`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [`string`](#string) |  |  |
| `default_f_port` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `name` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `default_f_port` | <p>`uint32.lte`: `255`</p><p>`uint32.gte`: `1`</p> |

### <a name="ttn.lorawan.v3.ApplicationPackageAssociation">Message `ApplicationPackageAssociation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationPackageAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageAssociationIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `package_name` | [`string`](#string) |  |  |
| `data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `package_name` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ApplicationPackageAssociationIdentifiers">Message `ApplicationPackageAssociationIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `f_port` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |
| `f_port` | <p>`uint32.lte`: `255`</p><p>`uint32.gte`: `1`</p> |

### <a name="ttn.lorawan.v3.ApplicationPackageAssociations">Message `ApplicationPackageAssociations`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `associations` | [`ApplicationPackageAssociation`](#ttn.lorawan.v3.ApplicationPackageAssociation) | repeated |  |

### <a name="ttn.lorawan.v3.ApplicationPackageDefaultAssociation">Message `ApplicationPackageDefaultAssociation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationPackageDefaultAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociationIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `package_name` | [`string`](#string) |  |  |
| `data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `package_name` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ApplicationPackageDefaultAssociationIdentifiers">Message `ApplicationPackageDefaultAssociationIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `f_port` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `f_port` | <p>`uint32.lte`: `255`</p><p>`uint32.gte`: `1`</p> |

### <a name="ttn.lorawan.v3.ApplicationPackageDefaultAssociations">Message `ApplicationPackageDefaultAssociations`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `defaults` | [`ApplicationPackageDefaultAssociation`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociation) | repeated |  |

### <a name="ttn.lorawan.v3.ApplicationPackages">Message `ApplicationPackages`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `packages` | [`ApplicationPackage`](#ttn.lorawan.v3.ApplicationPackage) | repeated |  |

### <a name="ttn.lorawan.v3.GetApplicationPackageAssociationRequest">Message `GetApplicationPackageAssociationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationPackageAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageAssociationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetApplicationPackageDefaultAssociationRequest">Message `GetApplicationPackageDefaultAssociationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationPackageDefaultAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ListApplicationPackageAssociationRequest">Message `ListApplicationPackageAssociationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. Each page is ordered by the FPort. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListApplicationPackageDefaultAssociationRequest">Message `ListApplicationPackageDefaultAssociationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. Each page is ordered by the FPort. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SetApplicationPackageAssociationRequest">Message `SetApplicationPackageAssociationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `association` | [`ApplicationPackageAssociation`](#ttn.lorawan.v3.ApplicationPackageAssociation) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `association` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.SetApplicationPackageDefaultAssociationRequest">Message `SetApplicationPackageDefaultAssociationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default` | [`ApplicationPackageDefaultAssociation`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociation) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `default` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationPackageRegistry">Service `ApplicationPackageRegistry`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `List` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`ApplicationPackages`](#ttn.lorawan.v3.ApplicationPackages) | List returns the available packages for the end device. |
| `GetAssociation` | [`GetApplicationPackageAssociationRequest`](#ttn.lorawan.v3.GetApplicationPackageAssociationRequest) | [`ApplicationPackageAssociation`](#ttn.lorawan.v3.ApplicationPackageAssociation) | GetAssociation returns the association registered on the FPort of the end device. |
| `ListAssociations` | [`ListApplicationPackageAssociationRequest`](#ttn.lorawan.v3.ListApplicationPackageAssociationRequest) | [`ApplicationPackageAssociations`](#ttn.lorawan.v3.ApplicationPackageAssociations) | ListAssociations returns all of the associations of the end device. |
| `SetAssociation` | [`SetApplicationPackageAssociationRequest`](#ttn.lorawan.v3.SetApplicationPackageAssociationRequest) | [`ApplicationPackageAssociation`](#ttn.lorawan.v3.ApplicationPackageAssociation) | SetAssociation updates or creates the association on the FPort of the end device. |
| `DeleteAssociation` | [`ApplicationPackageAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageAssociationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | DeleteAssociation removes the association on the FPort of the end device. |
| `GetDefaultAssociation` | [`GetApplicationPackageDefaultAssociationRequest`](#ttn.lorawan.v3.GetApplicationPackageDefaultAssociationRequest) | [`ApplicationPackageDefaultAssociation`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociation) | GetDefaultAssociation returns the default association registered on the FPort of the application. |
| `ListDefaultAssociations` | [`ListApplicationPackageDefaultAssociationRequest`](#ttn.lorawan.v3.ListApplicationPackageDefaultAssociationRequest) | [`ApplicationPackageDefaultAssociations`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociations) | ListDefaultAssociations returns all of the default associations of the application. |
| `SetDefaultAssociation` | [`SetApplicationPackageDefaultAssociationRequest`](#ttn.lorawan.v3.SetApplicationPackageDefaultAssociationRequest) | [`ApplicationPackageDefaultAssociation`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociation) | SetDefaultAssociation updates or creates the default association on the FPort of the application. |
| `DeleteDefaultAssociation` | [`ApplicationPackageDefaultAssociationIdentifiers`](#ttn.lorawan.v3.ApplicationPackageDefaultAssociationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | DeleteDefaultAssociation removes the default association on the FPort of the application. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `List` | `GET` | `/api/v3/as/applications/{application_ids.application_id}/devices/{device_id}/packages` |  |
| `GetAssociation` | `GET` | `/api/v3/as/applications/{ids.end_device_ids.application_ids.application_id}/devices/{ids.end_device_ids.device_id}/packages/associations/{ids.f_port}` |  |
| `ListAssociations` | `GET` | `/api/v3/as/applications/{ids.application_ids.application_id}/devices/{ids.device_id}/packages/associations` |  |
| `SetAssociation` | `PUT` | `/api/v3/as/applications/{association.ids.end_device_ids.application_ids.application_id}/devices/{association.ids.end_device_ids.device_id}/packages/associations/{association.ids.f_port}` | `*` |
| `DeleteAssociation` | `DELETE` | `/api/v3/as/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}/packages/associations/{f_port}` |  |
| `GetDefaultAssociation` | `GET` | `/api/v3/as/applications/{ids.application_ids.application_id}/packages/associations/{ids.f_port}` |  |
| `ListDefaultAssociations` | `GET` | `/api/v3/as/applications/{ids.application_id}/packages/associations` |  |
| `SetDefaultAssociation` | `PUT` | `/api/v3/as/applications/{default.ids.application_ids.application_id}/packages/associations/{default.ids.f_port}` | `*` |
| `DeleteDefaultAssociation` | `DELETE` | `/api/v3/as/applications/{application_ids.application_id}/packages/associations/{f_port}` |  |

## <a name="ttn/lorawan/v3/applicationserver_pubsub.proto">File `ttn/lorawan/v3/applicationserver_pubsub.proto`</a>

### <a name="ttn.lorawan.v3.ApplicationPubSub">Message `ApplicationPubSub`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationPubSubIdentifiers`](#ttn.lorawan.v3.ApplicationPubSubIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `format` | [`string`](#string) |  | The format to use for the body. Supported values depend on the Application Server configuration. |
| `nats` | [`ApplicationPubSub.NATSProvider`](#ttn.lorawan.v3.ApplicationPubSub.NATSProvider) |  |  |
| `mqtt` | [`ApplicationPubSub.MQTTProvider`](#ttn.lorawan.v3.ApplicationPubSub.MQTTProvider) |  |  |
| `aws_iot` | [`ApplicationPubSub.AWSIoTProvider`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider) |  |  |
| `base_topic` | [`string`](#string) |  | Base topic name to which the messages topic is appended. |
| `downlink_push` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  | The topic to which the Application Server subscribes for downlink queue push operations. |
| `downlink_replace` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  | The topic to which the Application Server subscribes for downlink queue replace operations. |
| `uplink_message` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `uplink_normalized` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `join_accept` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `downlink_ack` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `downlink_nack` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `downlink_sent` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `downlink_failed` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `downlink_queued` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `downlink_queue_invalidated` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `location_solved` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |
| `service_data` | [`ApplicationPubSub.Message`](#ttn.lorawan.v3.ApplicationPubSub.Message) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `format` | <p>`string.max_len`: `20`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `base_topic` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider">Message `ApplicationPubSub.AWSIoTProvider`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `region` | [`string`](#string) |  | The AWS region. |
| `access_key` | [`ApplicationPubSub.AWSIoTProvider.AccessKey`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.AccessKey) |  | If set, the integration will use an AWS access key. |
| `assume_role` | [`ApplicationPubSub.AWSIoTProvider.AssumeRole`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.AssumeRole) |  | If set, the integration will assume the given role during operation. |
| `endpoint_address` | [`string`](#string) |  | The endpoint address to connect to. If the endpoint address is left empty, the integration will try to discover it. |
| `default` | [`ApplicationPubSub.AWSIoTProvider.DefaultIntegration`](#ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.DefaultIntegration) |  | Enable the default integration. This overrides custom base topic and message topics of the pub/sub integration. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `region` | <p>`string.in`: `[af-south-1 ap-east-1 ap-northeast-1 ap-northeast-2 ap-south-1 ap-southeast-1 ap-southeast-2 ca-central-1 eu-central-1 eu-north-1 eu-south-1 eu-west-1 eu-west-2 eu-west-3 me-south-1 sa-east-1 us-east-1 us-east-2 us-west-1 us-west-2]`</p> |
| `endpoint_address` | <p>`string.max_len`: `128`</p><p>`string.pattern`: `^((([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])|)$`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.AccessKey">Message `ApplicationPubSub.AWSIoTProvider.AccessKey`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `access_key_id` | [`string`](#string) |  |  |
| `secret_access_key` | [`string`](#string) |  |  |
| `session_token` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `access_key_id` | <p>`string.min_len`: `16`</p><p>`string.max_len`: `128`</p><p>`string.pattern`: `^[\w]*$`</p> |
| `secret_access_key` | <p>`string.max_len`: `40`</p> |
| `session_token` | <p>`string.max_len`: `256`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.AssumeRole">Message `ApplicationPubSub.AWSIoTProvider.AssumeRole`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `arn` | [`string`](#string) |  |  |
| `external_id` | [`string`](#string) |  |  |
| `session_duration` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `arn` | <p>`string.pattern`: `^arn:aws:iam::[0-9]{12}:role\/[A-Za-z0-9_+=,.@-]+$`</p> |
| `external_id` | <p>`string.max_len`: `1224`</p><p>`string.pattern`: `^[\w+=,.@:\/-]*$`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.AWSIoTProvider.DefaultIntegration">Message `ApplicationPubSub.AWSIoTProvider.DefaultIntegration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `stack_name` | [`string`](#string) |  | The stack name that is associated with the CloudFormation deployment of The Things Stack Enterprise integration. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `stack_name` | <p>`string.max_len`: `128`</p><p>`string.pattern`: `^[A-Za-z][A-Za-z0-9\-]*$`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.MQTTProvider">Message `ApplicationPubSub.MQTTProvider`</a>

The MQTT provider settings.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `server_url` | [`string`](#string) |  |  |
| `client_id` | [`string`](#string) |  |  |
| `username` | [`string`](#string) |  |  |
| `password` | [`string`](#string) |  |  |
| `subscribe_qos` | [`ApplicationPubSub.MQTTProvider.QoS`](#ttn.lorawan.v3.ApplicationPubSub.MQTTProvider.QoS) |  |  |
| `publish_qos` | [`ApplicationPubSub.MQTTProvider.QoS`](#ttn.lorawan.v3.ApplicationPubSub.MQTTProvider.QoS) |  |  |
| `use_tls` | [`bool`](#bool) |  |  |
| `tls_ca` | [`bytes`](#bytes) |  | The server Root CA certificate. PEM formatted. |
| `tls_client_cert` | [`bytes`](#bytes) |  | The client certificate. PEM formatted. |
| `tls_client_key` | [`bytes`](#bytes) |  | The client private key. PEM formatted. |
| `headers` | [`ApplicationPubSub.MQTTProvider.HeadersEntry`](#ttn.lorawan.v3.ApplicationPubSub.MQTTProvider.HeadersEntry) | repeated | HTTP headers to use on MQTT-over-Websocket connections. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `server_url` | <p>`string.uri`: `true`</p> |
| `client_id` | <p>`string.max_len`: `23`</p> |
| `username` | <p>`string.max_len`: `100`</p> |
| `password` | <p>`string.max_len`: `100`</p> |
| `tls_ca` | <p>`bytes.max_len`: `8192`</p> |
| `tls_client_cert` | <p>`bytes.max_len`: `8192`</p> |
| `tls_client_key` | <p>`bytes.max_len`: `8192`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.MQTTProvider.HeadersEntry">Message `ApplicationPubSub.MQTTProvider.HeadersEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.ApplicationPubSub.Message">Message `ApplicationPubSub.Message`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `topic` | [`string`](#string) |  | The topic on which the Application Server publishes or receives the messages. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `topic` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.NATSProvider">Message `ApplicationPubSub.NATSProvider`</a>

The NATS provider settings.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `server_url` | [`string`](#string) |  | The server connection URL. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `server_url` | <p>`string.uri`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSubFormats">Message `ApplicationPubSubFormats`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `formats` | [`ApplicationPubSubFormats.FormatsEntry`](#ttn.lorawan.v3.ApplicationPubSubFormats.FormatsEntry) | repeated | Format and description. |

### <a name="ttn.lorawan.v3.ApplicationPubSubFormats.FormatsEntry">Message `ApplicationPubSubFormats.FormatsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.ApplicationPubSubIdentifiers">Message `ApplicationPubSubIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `pub_sub_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `pub_sub_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSubs">Message `ApplicationPubSubs`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pubsubs` | [`ApplicationPubSub`](#ttn.lorawan.v3.ApplicationPubSub) | repeated |  |

### <a name="ttn.lorawan.v3.GetApplicationPubSubRequest">Message `GetApplicationPubSubRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationPubSubIdentifiers`](#ttn.lorawan.v3.ApplicationPubSubIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ListApplicationPubSubsRequest">Message `ListApplicationPubSubsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.SetApplicationPubSubRequest">Message `SetApplicationPubSubRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pubsub` | [`ApplicationPubSub`](#ttn.lorawan.v3.ApplicationPubSub) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `pubsub` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationPubSub.MQTTProvider.QoS">Enum `ApplicationPubSub.MQTTProvider.QoS`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `AT_MOST_ONCE` | 0 |  |
| `AT_LEAST_ONCE` | 1 |  |
| `EXACTLY_ONCE` | 2 |  |

### <a name="ttn.lorawan.v3.ApplicationPubSubRegistry">Service `ApplicationPubSubRegistry`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetFormats` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`ApplicationPubSubFormats`](#ttn.lorawan.v3.ApplicationPubSubFormats) |  |
| `Get` | [`GetApplicationPubSubRequest`](#ttn.lorawan.v3.GetApplicationPubSubRequest) | [`ApplicationPubSub`](#ttn.lorawan.v3.ApplicationPubSub) |  |
| `List` | [`ListApplicationPubSubsRequest`](#ttn.lorawan.v3.ListApplicationPubSubsRequest) | [`ApplicationPubSubs`](#ttn.lorawan.v3.ApplicationPubSubs) |  |
| `Set` | [`SetApplicationPubSubRequest`](#ttn.lorawan.v3.SetApplicationPubSubRequest) | [`ApplicationPubSub`](#ttn.lorawan.v3.ApplicationPubSub) |  |
| `Delete` | [`ApplicationPubSubIdentifiers`](#ttn.lorawan.v3.ApplicationPubSubIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetFormats` | `GET` | `/api/v3/as/pubsub-formats` |  |
| `Get` | `GET` | `/api/v3/as/pubsub/{ids.application_ids.application_id}/{ids.pub_sub_id}` |  |
| `List` | `GET` | `/api/v3/as/pubsub/{application_ids.application_id}` |  |
| `Set` | `PUT` | `/api/v3/as/pubsub/{pubsub.ids.application_ids.application_id}/{pubsub.ids.pub_sub_id}` | `*` |
| `Set` | `POST` | `/api/v3/as/pubsub/{pubsub.ids.application_ids.application_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/as/pubsub/{application_ids.application_id}/{pub_sub_id}` |  |

## <a name="ttn/lorawan/v3/applicationserver_web.proto">File `ttn/lorawan/v3/applicationserver_web.proto`</a>

### <a name="ttn.lorawan.v3.ApplicationWebhook">Message `ApplicationWebhook`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationWebhookIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `base_url` | [`string`](#string) |  | Base URL to which the message's path is appended. |
| `headers` | [`ApplicationWebhook.HeadersEntry`](#ttn.lorawan.v3.ApplicationWebhook.HeadersEntry) | repeated | HTTP headers to use. |
| `format` | [`string`](#string) |  | The format to use for the body. Supported values depend on the Application Server configuration. |
| `template_ids` | [`ApplicationWebhookTemplateIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookTemplateIdentifiers) |  | The ID of the template that was used to create the Webhook. |
| `template_fields` | [`ApplicationWebhook.TemplateFieldsEntry`](#ttn.lorawan.v3.ApplicationWebhook.TemplateFieldsEntry) | repeated | The value of the fields used by the template. Maps field.id to the value. |
| `downlink_api_key` | [`string`](#string) |  | The API key to be used for downlink queue operations. The field is provided for convenience reasons, and can contain API keys with additional rights (albeit this is discouraged). |
| `uplink_message` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `uplink_normalized` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `join_accept` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `downlink_ack` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `downlink_nack` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `downlink_sent` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `downlink_failed` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `downlink_queued` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `downlink_queue_invalidated` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `location_solved` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `service_data` | [`ApplicationWebhook.Message`](#ttn.lorawan.v3.ApplicationWebhook.Message) |  |  |
| `health_status` | [`ApplicationWebhookHealth`](#ttn.lorawan.v3.ApplicationWebhookHealth) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `base_url` | <p>`string.uri`: `true`</p> |
| `headers` | <p>`map.max_pairs`: `50`</p><p>`map.keys.string.max_len`: `64`</p><p>`map.values.string.max_len`: `4096`</p> |
| `format` | <p>`string.max_len`: `20`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `downlink_api_key` | <p>`string.max_len`: `128`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhook.HeadersEntry">Message `ApplicationWebhook.HeadersEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.ApplicationWebhook.Message">Message `ApplicationWebhook.Message`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [`string`](#string) |  | Path to append to the base URL. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `path` | <p>`string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhook.TemplateFieldsEntry">Message `ApplicationWebhook.TemplateFieldsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.ApplicationWebhookFormats">Message `ApplicationWebhookFormats`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `formats` | [`ApplicationWebhookFormats.FormatsEntry`](#ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry) | repeated | Format and description. |

### <a name="ttn.lorawan.v3.ApplicationWebhookFormats.FormatsEntry">Message `ApplicationWebhookFormats.FormatsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.ApplicationWebhookHealth">Message `ApplicationWebhookHealth`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `healthy` | [`ApplicationWebhookHealth.WebhookHealthStatusHealthy`](#ttn.lorawan.v3.ApplicationWebhookHealth.WebhookHealthStatusHealthy) |  |  |
| `unhealthy` | [`ApplicationWebhookHealth.WebhookHealthStatusUnhealthy`](#ttn.lorawan.v3.ApplicationWebhookHealth.WebhookHealthStatusUnhealthy) |  |  |

### <a name="ttn.lorawan.v3.ApplicationWebhookHealth.WebhookHealthStatusHealthy">Message `ApplicationWebhookHealth.WebhookHealthStatusHealthy`</a>

### <a name="ttn.lorawan.v3.ApplicationWebhookHealth.WebhookHealthStatusUnhealthy">Message `ApplicationWebhookHealth.WebhookHealthStatusUnhealthy`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `failed_attempts` | [`uint64`](#uint64) |  |  |
| `last_failed_attempt_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `last_failed_attempt_details` | [`ErrorDetails`](#ttn.lorawan.v3.ErrorDetails) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `last_failed_attempt_at` | <p>`timestamp.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhookIdentifiers">Message `ApplicationWebhookIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `webhook_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `webhook_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhookTemplate">Message `ApplicationWebhookTemplate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationWebhookTemplateIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookTemplateIdentifiers) |  |  |
| `name` | [`string`](#string) |  |  |
| `description` | [`string`](#string) |  |  |
| `logo_url` | [`string`](#string) |  |  |
| `info_url` | [`string`](#string) |  |  |
| `documentation_url` | [`string`](#string) |  |  |
| `base_url` | [`string`](#string) |  | The base URL of the template. Can contain template fields, in RFC 6570 format. |
| `headers` | [`ApplicationWebhookTemplate.HeadersEntry`](#ttn.lorawan.v3.ApplicationWebhookTemplate.HeadersEntry) | repeated | The HTTP headers used by the template. Both the key and the value can contain template fields. |
| `format` | [`string`](#string) |  |  |
| `fields` | [`ApplicationWebhookTemplateField`](#ttn.lorawan.v3.ApplicationWebhookTemplateField) | repeated |  |
| `create_downlink_api_key` | [`bool`](#bool) |  | Control the creation of the downlink queue operations API key. |
| `uplink_message` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `uplink_normalized` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `join_accept` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `downlink_ack` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `downlink_nack` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `downlink_sent` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `downlink_failed` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `downlink_queued` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `downlink_queue_invalidated` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `location_solved` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `service_data` | [`ApplicationWebhookTemplate.Message`](#ttn.lorawan.v3.ApplicationWebhookTemplate.Message) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `20`</p> |
| `description` | <p>`string.max_len`: `100`</p> |
| `logo_url` | <p>`string.uri`: `true`</p> |
| `info_url` | <p>`string.uri`: `true`</p> |
| `documentation_url` | <p>`string.uri`: `true`</p> |
| `base_url` | <p>`string.uri`: `true`</p> |
| `headers` | <p>`map.max_pairs`: `50`</p><p>`map.keys.string.max_len`: `64`</p><p>`map.values.string.max_len`: `256`</p> |
| `format` | <p>`string.max_len`: `20`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhookTemplate.HeadersEntry">Message `ApplicationWebhookTemplate.HeadersEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.ApplicationWebhookTemplate.Message">Message `ApplicationWebhookTemplate.Message`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path` | [`string`](#string) |  | Path to append to the base URL. Can contain template fields, in RFC 6570 format. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `path` | <p>`string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhookTemplateField">Message `ApplicationWebhookTemplateField`</a>

ApplicationWebhookTemplateField represents a custom field that needs to be filled by the user in order to use the template.
A field can be an API key, an username or password, or any custom platform specific field (such as region).
The fields are meant to be replaced inside the URLs and headers when the webhook is created.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  |  |
| `name` | [`string`](#string) |  |  |
| `description` | [`string`](#string) |  |  |
| `secret` | [`bool`](#bool) |  | Secret decides if the field should be shown in plain-text or should stay hidden. |
| `default_value` | [`string`](#string) |  |  |
| `optional` | [`bool`](#bool) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `name` | <p>`string.max_len`: `20`</p> |
| `description` | <p>`string.max_len`: `100`</p> |
| `default_value` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhookTemplateIdentifiers">Message `ApplicationWebhookTemplateIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `template_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `template_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhookTemplates">Message `ApplicationWebhookTemplates`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `templates` | [`ApplicationWebhookTemplate`](#ttn.lorawan.v3.ApplicationWebhookTemplate) | repeated |  |

### <a name="ttn.lorawan.v3.ApplicationWebhooks">Message `ApplicationWebhooks`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `webhooks` | [`ApplicationWebhook`](#ttn.lorawan.v3.ApplicationWebhook) | repeated |  |

### <a name="ttn.lorawan.v3.GetApplicationWebhookRequest">Message `GetApplicationWebhookRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationWebhookIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetApplicationWebhookTemplateRequest">Message `GetApplicationWebhookTemplateRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ApplicationWebhookTemplateIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookTemplateIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ListApplicationWebhookTemplatesRequest">Message `ListApplicationWebhookTemplatesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

### <a name="ttn.lorawan.v3.ListApplicationWebhooksRequest">Message `ListApplicationWebhooksRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.SetApplicationWebhookRequest">Message `SetApplicationWebhookRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `webhook` | [`ApplicationWebhook`](#ttn.lorawan.v3.ApplicationWebhook) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `webhook` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationWebhookRegistry">Service `ApplicationWebhookRegistry`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetFormats` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`ApplicationWebhookFormats`](#ttn.lorawan.v3.ApplicationWebhookFormats) |  |
| `GetTemplate` | [`GetApplicationWebhookTemplateRequest`](#ttn.lorawan.v3.GetApplicationWebhookTemplateRequest) | [`ApplicationWebhookTemplate`](#ttn.lorawan.v3.ApplicationWebhookTemplate) |  |
| `ListTemplates` | [`ListApplicationWebhookTemplatesRequest`](#ttn.lorawan.v3.ListApplicationWebhookTemplatesRequest) | [`ApplicationWebhookTemplates`](#ttn.lorawan.v3.ApplicationWebhookTemplates) |  |
| `Get` | [`GetApplicationWebhookRequest`](#ttn.lorawan.v3.GetApplicationWebhookRequest) | [`ApplicationWebhook`](#ttn.lorawan.v3.ApplicationWebhook) |  |
| `List` | [`ListApplicationWebhooksRequest`](#ttn.lorawan.v3.ListApplicationWebhooksRequest) | [`ApplicationWebhooks`](#ttn.lorawan.v3.ApplicationWebhooks) |  |
| `Set` | [`SetApplicationWebhookRequest`](#ttn.lorawan.v3.SetApplicationWebhookRequest) | [`ApplicationWebhook`](#ttn.lorawan.v3.ApplicationWebhook) |  |
| `Delete` | [`ApplicationWebhookIdentifiers`](#ttn.lorawan.v3.ApplicationWebhookIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetFormats` | `GET` | `/api/v3/as/webhook-formats` |  |
| `GetTemplate` | `GET` | `/api/v3/as/webhook-templates/{ids.template_id}` |  |
| `ListTemplates` | `GET` | `/api/v3/as/webhook-templates` |  |
| `Get` | `GET` | `/api/v3/as/webhooks/{ids.application_ids.application_id}/{ids.webhook_id}` |  |
| `List` | `GET` | `/api/v3/as/webhooks/{application_ids.application_id}` |  |
| `Set` | `PUT` | `/api/v3/as/webhooks/{webhook.ids.application_ids.application_id}/{webhook.ids.webhook_id}` | `*` |
| `Set` | `POST` | `/api/v3/as/webhooks/{webhook.ids.application_ids.application_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/as/webhooks/{application_ids.application_id}/{webhook_id}` |  |

## <a name="ttn/lorawan/v3/client.proto">File `ttn/lorawan/v3/client.proto`</a>

### <a name="ttn.lorawan.v3.Client">Message `Client`</a>

An OAuth client on the network.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  | The identifiers of the OAuth client. These are public and can be seen by any authenticated user in the network. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the OAuth client was created. This information is public and can be seen by any authenticated user in the network. |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the OAuth client was last updated. This information is public and can be seen by any authenticated user in the network. |
| `deleted_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the OAuth client was deleted. This information is public and can be seen by any authenticated user in the network. |
| `name` | [`string`](#string) |  | The name of the OAuth client. This information is public and can be seen by any authenticated user in the network. |
| `description` | [`string`](#string) |  | A description for the OAuth client. This information is public and can be seen by any authenticated user in the network. |
| `attributes` | [`Client.AttributesEntry`](#ttn.lorawan.v3.Client.AttributesEntry) | repeated | Key-value attributes for this client. Typically used for organizing clients or for storing integration-specific data. |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated | Contact information for this client. Typically used to indicate who to contact with technical/security questions about the application. This information is public and can be seen by any authenticated user in the network. This field is deprecated. Use administrative_contact and technical_contact instead. |
| `administrative_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `technical_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `secret` | [`string`](#string) |  | The client secret is only visible to collaborators of the client. |
| `redirect_uris` | [`string`](#string) | repeated | The allowed redirect URIs against which authorization requests are checked. If the authorization request does not pass a redirect URI, the first one from this list is taken. This information is public and can be seen by any authenticated user in the network. |
| `logout_redirect_uris` | [`string`](#string) | repeated | The allowed logout redirect URIs against which client initiated logout requests are checked. If the authorization request does not pass a redirect URI, the first one from this list is taken. This information is public and can be seen by any authenticated user in the network. |
| `state` | [`State`](#ttn.lorawan.v3.State) |  | The reviewing state of the client. This information is public and can be seen by any authenticated user in the network. This field can only be modified by admins. If state_description is not updated when updating state, state_description is cleared. |
| `state_description` | [`string`](#string) |  | A description for the state field. This field can only be modified by admins, and should typically only be updated when also updating `state`. |
| `skip_authorization` | [`bool`](#bool) |  | If set, the authorization page will be skipped. This information is public and can be seen by any authenticated user in the network. This field can only be modified by admins. |
| `endorsed` | [`bool`](#bool) |  | If set, the authorization page will show endorsement. This information is public and can be seen by any authenticated user in the network. This field can only be modified by admins. |
| `grants` | [`GrantType`](#ttn.lorawan.v3.GrantType) | repeated | OAuth flows that can be used for the client to get a token. This information is public and can be seen by any authenticated user in the network. After a client is created, this field can only be modified by admins. |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated | Rights denotes what rights the client will have access to. This information is public and can be seen by any authenticated user in the network. Users that previously authorized this client will have to re-authorize the client after rights are added to this list. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `description` | <p>`string.max_len`: `2000`</p> |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |
| `contact_info` | <p>`repeated.max_items`: `10`</p> |
| `secret` | <p>`string.max_len`: `128`</p> |
| `redirect_uris` | <p>`repeated.max_items`: `10`</p><p>`repeated.items.string.max_len`: `128`</p> |
| `logout_redirect_uris` | <p>`repeated.max_items`: `10`</p><p>`repeated.items.string.max_len`: `128`</p> |
| `state` | <p>`enum.defined_only`: `true`</p> |
| `state_description` | <p>`string.max_len`: `128`</p> |
| `grants` | <p>`repeated.items.enum.defined_only`: `true`</p> |
| `rights` | <p>`repeated.items.enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.Client.AttributesEntry">Message `Client.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.Clients">Message `Clients`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `clients` | [`Client`](#ttn.lorawan.v3.Client) | repeated |  |

### <a name="ttn.lorawan.v3.CreateClientRequest">Message `CreateClientRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client` | [`Client`](#ttn.lorawan.v3.Client) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created client. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteClientCollaboratorRequest">Message `DeleteClientCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `collaborator_ids` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client_ids` | <p>`message.required`: `true`</p> |
| `collaborator_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetClientCollaboratorRequest">Message `GetClientCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetClientRequest">Message `GetClientRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the client fields that should be returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ListClientCollaboratorsRequest">Message `ListClientCollaboratorsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client_ids` | <p>`message.required`: `true`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `order` | <p>`string.in`: `[ id -id -rights rights]`</p> |

### <a name="ttn.lorawan.v3.ListClientsRequest">Message `ListClientsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | By default we list all OAuth clients the caller has rights on. Set the user or the organization (not both) to instead list the OAuth clients where the user or organization is collaborator on. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the client fields that should be returned. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted clients. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `order` | <p>`string.in`: `[ client_id -client_id name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SetClientCollaboratorRequest">Message `SetClientCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `collaborator` | [`Collaborator`](#ttn.lorawan.v3.Collaborator) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateClientRequest">Message `UpdateClientRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client` | [`Client`](#ttn.lorawan.v3.Client) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the client fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GrantType">Enum `GrantType`</a>

The OAuth2 flows an OAuth client can use to get an access token.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `GRANT_AUTHORIZATION_CODE` | 0 | Grant type used to exchange an authorization code for an access token. |
| `GRANT_PASSWORD` | 1 | Grant type used to exchange a user ID and password for an access token. |
| `GRANT_REFRESH_TOKEN` | 2 | Grant type used to exchange a refresh token for an access token. |

## <a name="ttn/lorawan/v3/client_services.proto">File `ttn/lorawan/v3/client_services.proto`</a>

### <a name="ttn.lorawan.v3.ClientAccess">Service `ClientAccess`</a>

The ClientAcces service, exposed by the Identity Server, is used to manage
collaborators of OAuth clients.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListRights` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) | [`Rights`](#ttn.lorawan.v3.Rights) | List the rights the caller has on this application. |
| `GetCollaborator` | [`GetClientCollaboratorRequest`](#ttn.lorawan.v3.GetClientCollaboratorRequest) | [`GetCollaboratorResponse`](#ttn.lorawan.v3.GetCollaboratorResponse) | Get the rights of a collaborator (member) of the client. Pseudo-rights in the response (such as the "_ALL" right) are not expanded. |
| `SetCollaborator` | [`SetClientCollaboratorRequest`](#ttn.lorawan.v3.SetClientCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Set the rights of a collaborator (member) on the OAuth client. This method can also be used to delete the collaborator, by giving them no rights. The caller is required to have all assigned or/and removed rights. |
| `ListCollaborators` | [`ListClientCollaboratorsRequest`](#ttn.lorawan.v3.ListClientCollaboratorsRequest) | [`Collaborators`](#ttn.lorawan.v3.Collaborators) | List the collaborators on this OAuth client. |
| `DeleteCollaborator` | [`DeleteClientCollaboratorRequest`](#ttn.lorawan.v3.DeleteClientCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | DeleteCollaborator removes a collaborator from a client. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListRights` | `GET` | `/api/v3/clients/{client_id}/rights` |  |
| `GetCollaborator` | `` | `/api/v3` |  |
| `GetCollaborator` | `GET` | `/api/v3/clients/{client_ids.client_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
| `GetCollaborator` | `GET` | `/api/v3/clients/{client_ids.client_id}/collaborator/organization/{collaborator.organization_ids.organization_id}` |  |
| `SetCollaborator` | `PUT` | `/api/v3/clients/{client_ids.client_id}/collaborators` | `*` |
| `ListCollaborators` | `GET` | `/api/v3/clients/{client_ids.client_id}/collaborators` |  |
| `DeleteCollaborator` | `` | `/api/v3` |  |
| `DeleteCollaborator` | `DELETE` | `/api/v3/clients/{client_ids.client_id}/collaborators/user/{collaborator_ids.user_ids.user_id}` |  |
| `DeleteCollaborator` | `DELETE` | `/api/v3/clients/{client_ids.client_id}/collaborators/organization/{collaborator_ids.organization_ids.organization_id}` |  |

### <a name="ttn.lorawan.v3.ClientRegistry">Service `ClientRegistry`</a>

The ClientRegistry service, exposed by the Identity Server, is used to manage
OAuth client registrations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateClientRequest`](#ttn.lorawan.v3.CreateClientRequest) | [`Client`](#ttn.lorawan.v3.Client) | Create a new OAuth client. This also sets the given organization or user as first collaborator with all possible rights. |
| `Get` | [`GetClientRequest`](#ttn.lorawan.v3.GetClientRequest) | [`Client`](#ttn.lorawan.v3.Client) | Get the OAuth client with the given identifiers, selecting the fields specified in the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `List` | [`ListClientsRequest`](#ttn.lorawan.v3.ListClientsRequest) | [`Clients`](#ttn.lorawan.v3.Clients) | List OAuth clients where the given user or organization is a direct collaborator. If no user or organization is given, this returns the OAuth clients the caller has access to. Similar to Get, this selects the fields specified in the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `Update` | [`UpdateClientRequest`](#ttn.lorawan.v3.UpdateClientRequest) | [`Client`](#ttn.lorawan.v3.Client) | Update the OAuth client, changing the fields specified by the field mask to the provided values. |
| `Delete` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the OAuth client. This may not release the client ID for reuse. |
| `Restore` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Restore a recently deleted client. Deployment configuration may specify if, and for how long after deletion, entities can be restored. |
| `Purge` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Purge the client. This will release the client ID for reuse. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Create` | `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/clients` | `*` |
| `Create` | `POST` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/clients` | `*` |
| `Get` | `GET` | `/api/v3/clients/{client_ids.client_id}` |  |
| `List` | `GET` | `/api/v3/clients` |  |
| `List` | `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/clients` |  |
| `List` | `GET` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/clients` |  |
| `Update` | `PUT` | `/api/v3/clients/{client.ids.client_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/clients/{client_id}` |  |
| `Restore` | `POST` | `/api/v3/clients/{client_id}/restore` |  |
| `Purge` | `DELETE` | `/api/v3/clients/{client_id}/purge` |  |

## <a name="ttn/lorawan/v3/cluster.proto">File `ttn/lorawan/v3/cluster.proto`</a>

### <a name="ttn.lorawan.v3.PeerInfo">Message `PeerInfo`</a>

PeerInfo

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `grpc_port` | [`uint32`](#uint32) |  | Port on which the gRPC server is exposed. |
| `tls` | [`bool`](#bool) |  | Indicates whether the gRPC server uses TLS. |
| `roles` | [`ClusterRole`](#ttn.lorawan.v3.ClusterRole) | repeated | Roles of the peer. |
| `tags` | [`PeerInfo.TagsEntry`](#ttn.lorawan.v3.PeerInfo.TagsEntry) | repeated | Tags of the peer |

### <a name="ttn.lorawan.v3.PeerInfo.TagsEntry">Message `PeerInfo.TagsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

## <a name="ttn/lorawan/v3/configuration_services.proto">File `ttn/lorawan/v3/configuration_services.proto`</a>

### <a name="ttn.lorawan.v3.BandDescription">Message `BandDescription`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  |  |
| `beacon` | [`BandDescription.Beacon`](#ttn.lorawan.v3.BandDescription.Beacon) |  |  |
| `ping_slot_frequencies` | [`uint64`](#uint64) | repeated |  |
| `max_uplink_channels` | [`uint32`](#uint32) |  |  |
| `uplink_channels` | [`BandDescription.Channel`](#ttn.lorawan.v3.BandDescription.Channel) | repeated |  |
| `max_downlink_channels` | [`uint32`](#uint32) |  |  |
| `downlink_channels` | [`BandDescription.Channel`](#ttn.lorawan.v3.BandDescription.Channel) | repeated |  |
| `sub_bands` | [`BandDescription.SubBandParameters`](#ttn.lorawan.v3.BandDescription.SubBandParameters) | repeated |  |
| `data_rates` | [`BandDescription.DataRatesEntry`](#ttn.lorawan.v3.BandDescription.DataRatesEntry) | repeated |  |
| `freq_multiplier` | [`uint64`](#uint64) |  |  |
| `implements_cf_list` | [`bool`](#bool) |  |  |
| `cf_list_type` | [`CFListType`](#ttn.lorawan.v3.CFListType) |  |  |
| `receive_delay_1` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `receive_delay_2` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `join_accept_delay_1` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `join_accept_delay_2` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `max_fcnt_gap` | [`uint64`](#uint64) |  |  |
| `supports_dynamic_adr` | [`bool`](#bool) |  |  |
| `adr_ack_limit` | [`ADRAckLimitExponent`](#ttn.lorawan.v3.ADRAckLimitExponent) |  |  |
| `min_retransmit_timeout` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `max_retransmit_timeout` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `tx_offset` | [`float`](#float) | repeated |  |
| `max_adr_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `relay_forward_delay` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `relay_receive_delay` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `tx_param_setup_req_support` | [`bool`](#bool) |  |  |
| `default_max_eirp` | [`float`](#float) |  |  |
| `default_rx2_parameters` | [`BandDescription.Rx2Parameters`](#ttn.lorawan.v3.BandDescription.Rx2Parameters) |  |  |
| `boot_dwell_time` | [`BandDescription.DwellTime`](#ttn.lorawan.v3.BandDescription.DwellTime) |  |  |
| `relay` | [`BandDescription.RelayParameters`](#ttn.lorawan.v3.BandDescription.RelayParameters) |  |  |

### <a name="ttn.lorawan.v3.BandDescription.BandDataRate">Message `BandDescription.BandDataRate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rate` | [`DataRate`](#ttn.lorawan.v3.DataRate) |  |  |

### <a name="ttn.lorawan.v3.BandDescription.Beacon">Message `BandDescription.Beacon`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `coding_rate` | [`string`](#string) |  |  |
| `frequencies` | [`uint64`](#uint64) | repeated |  |

### <a name="ttn.lorawan.v3.BandDescription.Channel">Message `BandDescription.Channel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency` | [`uint64`](#uint64) |  |  |
| `min_data_rate` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `max_data_rate` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |

### <a name="ttn.lorawan.v3.BandDescription.DataRatesEntry">Message `BandDescription.DataRatesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`uint32`](#uint32) |  |  |
| `value` | [`BandDescription.BandDataRate`](#ttn.lorawan.v3.BandDescription.BandDataRate) |  |  |

### <a name="ttn.lorawan.v3.BandDescription.DwellTime">Message `BandDescription.DwellTime`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uplinks` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `downlinks` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.BandDescription.RelayParameters">Message `BandDescription.RelayParameters`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `wor_channels` | [`BandDescription.RelayParameters.RelayWORChannel`](#ttn.lorawan.v3.BandDescription.RelayParameters.RelayWORChannel) | repeated |  |

### <a name="ttn.lorawan.v3.BandDescription.RelayParameters.RelayWORChannel">Message `BandDescription.RelayParameters.RelayWORChannel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency` | [`uint64`](#uint64) |  |  |
| `ack_frequency` | [`uint64`](#uint64) |  |  |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |

### <a name="ttn.lorawan.v3.BandDescription.Rx2Parameters">Message `BandDescription.Rx2Parameters`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `frequency` | [`uint64`](#uint64) |  |  |

### <a name="ttn.lorawan.v3.BandDescription.SubBandParameters">Message `BandDescription.SubBandParameters`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_frequency` | [`uint64`](#uint64) |  |  |
| `max_frequency` | [`uint64`](#uint64) |  |  |
| `duty_cycle` | [`float`](#float) |  |  |
| `max_eirp` | [`float`](#float) |  |  |

### <a name="ttn.lorawan.v3.FrequencyPlanDescription">Message `FrequencyPlanDescription`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  |  |
| `base_id` | [`string`](#string) |  | The ID of the frequency that the current frequency plan is based on. |
| `name` | [`string`](#string) |  |  |
| `base_frequency` | [`uint32`](#uint32) |  | Base frequency in MHz for hardware support (433, 470, 868 or 915) |
| `band_id` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.GetPhyVersionsRequest">Message `GetPhyVersionsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `band_id` | [`string`](#string) |  | Optional Band ID to filter the results. If unused, all supported Bands and their versions are returned. |

### <a name="ttn.lorawan.v3.GetPhyVersionsResponse">Message `GetPhyVersionsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `version_info` | [`GetPhyVersionsResponse.VersionInfo`](#ttn.lorawan.v3.GetPhyVersionsResponse.VersionInfo) | repeated |  |

### <a name="ttn.lorawan.v3.GetPhyVersionsResponse.VersionInfo">Message `GetPhyVersionsResponse.VersionInfo`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `band_id` | [`string`](#string) |  |  |
| `phy_versions` | [`PHYVersion`](#ttn.lorawan.v3.PHYVersion) | repeated |  |

### <a name="ttn.lorawan.v3.ListBandsRequest">Message `ListBandsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `band_id` | [`string`](#string) |  | Optional Band ID to filter the results. If unused, all supported Bands are returned. |
| `phy_version` | [`PHYVersion`](#ttn.lorawan.v3.PHYVersion) |  | Optional PHY version to filter the results. If unused, all supported versions are returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `phy_version` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.ListBandsResponse">Message `ListBandsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `descriptions` | [`ListBandsResponse.DescriptionsEntry`](#ttn.lorawan.v3.ListBandsResponse.DescriptionsEntry) | repeated |  |

### <a name="ttn.lorawan.v3.ListBandsResponse.DescriptionsEntry">Message `ListBandsResponse.DescriptionsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`ListBandsResponse.VersionedBandDescription`](#ttn.lorawan.v3.ListBandsResponse.VersionedBandDescription) |  |  |

### <a name="ttn.lorawan.v3.ListBandsResponse.VersionedBandDescription">Message `ListBandsResponse.VersionedBandDescription`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `band` | [`ListBandsResponse.VersionedBandDescription.BandEntry`](#ttn.lorawan.v3.ListBandsResponse.VersionedBandDescription.BandEntry) | repeated |  |

### <a name="ttn.lorawan.v3.ListBandsResponse.VersionedBandDescription.BandEntry">Message `ListBandsResponse.VersionedBandDescription.BandEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`BandDescription`](#ttn.lorawan.v3.BandDescription) |  |  |

### <a name="ttn.lorawan.v3.ListFrequencyPlansRequest">Message `ListFrequencyPlansRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_frequency` | [`uint32`](#uint32) |  | Optional base frequency in MHz for hardware support (433, 470, 868 or 915) |
| `band_id` | [`string`](#string) |  | Optional Band ID to filter the results. |

### <a name="ttn.lorawan.v3.ListFrequencyPlansResponse">Message `ListFrequencyPlansResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency_plans` | [`FrequencyPlanDescription`](#ttn.lorawan.v3.FrequencyPlanDescription) | repeated |  |

### <a name="ttn.lorawan.v3.Configuration">Service `Configuration`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListFrequencyPlans` | [`ListFrequencyPlansRequest`](#ttn.lorawan.v3.ListFrequencyPlansRequest) | [`ListFrequencyPlansResponse`](#ttn.lorawan.v3.ListFrequencyPlansResponse) |  |
| `GetPhyVersions` | [`GetPhyVersionsRequest`](#ttn.lorawan.v3.GetPhyVersionsRequest) | [`GetPhyVersionsResponse`](#ttn.lorawan.v3.GetPhyVersionsResponse) | Returns a list of supported LoRaWAN PHY Versions for the given Band ID. |
| `ListBands` | [`ListBandsRequest`](#ttn.lorawan.v3.ListBandsRequest) | [`ListBandsResponse`](#ttn.lorawan.v3.ListBandsResponse) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListFrequencyPlans` | `GET` | `/api/v3/configuration/frequency-plans` |  |
| `GetPhyVersions` | `GET` | `/api/v3/configuration/phy-versions` |  |
| `ListBands` | `GET` | `/api/v3/configuration/bands` |  |
| `ListBands` | `GET` | `/api/v3/configuration/bands/{band_id}` |  |
| `ListBands` | `GET` | `/api/v3/configuration/bands/{band_id}/{phy_version}` |  |

## <a name="ttn/lorawan/v3/contact_info.proto">File `ttn/lorawan/v3/contact_info.proto`</a>

### <a name="ttn.lorawan.v3.ContactInfo">Message `ContactInfo`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contact_type` | [`ContactType`](#ttn.lorawan.v3.ContactType) |  |  |
| `contact_method` | [`ContactMethod`](#ttn.lorawan.v3.ContactMethod) |  |  |
| `value` | [`string`](#string) |  |  |
| `public` | [`bool`](#bool) |  |  |
| `validated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `contact_type` | <p>`enum.defined_only`: `true`</p> |
| `contact_method` | <p>`enum.defined_only`: `true`</p> |
| `value` | <p>`string.max_len`: `256`</p> |

### <a name="ttn.lorawan.v3.ContactInfoValidation">Message `ContactInfoValidation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  |  |
| `token` | [`string`](#string) |  |  |
| `entity` | [`EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers) |  |  |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `id` | <p>`string.min_len`: `1`</p><p>`string.max_len`: `64`</p> |
| `token` | <p>`string.min_len`: `1`</p><p>`string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.ContactMethod">Enum `ContactMethod`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `CONTACT_METHOD_OTHER` | 0 |  |
| `CONTACT_METHOD_EMAIL` | 1 |  |
| `CONTACT_METHOD_PHONE` | 2 |  |

### <a name="ttn.lorawan.v3.ContactType">Enum `ContactType`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `CONTACT_TYPE_OTHER` | 0 |  |
| `CONTACT_TYPE_ABUSE` | 1 |  |
| `CONTACT_TYPE_BILLING` | 2 |  |
| `CONTACT_TYPE_TECHNICAL` | 3 |  |

### <a name="ttn.lorawan.v3.ContactInfoRegistry">Service `ContactInfoRegistry`</a>

The ContactInfoRegistry service, exposed by the Identity Server, is used for
validating contact information of registered entities.

The actual contact information can be managed with the different registry services:
ApplicationRegistry, ClientRegistry, GatewayRegistry, OrganizationRegistry and UserRegistry.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `RequestValidation` | [`EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers) | [`ContactInfoValidation`](#ttn.lorawan.v3.ContactInfoValidation) | Request validation for the non-validated contact info for the given entity. |
| `Validate` | [`ContactInfoValidation`](#ttn.lorawan.v3.ContactInfoValidation) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Validate confirms a contact info validation. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `RequestValidation` | `POST` | `/api/v3/contact_info/validation` | `*` |
| `Validate` | `PATCH` | `/api/v3/contact_info/validation` | `*` |

## <a name="ttn/lorawan/v3/deviceclaimingserver.proto">File `ttn/lorawan/v3/deviceclaimingserver.proto`</a>

### <a name="ttn.lorawan.v3.AuthorizeApplicationRequest">Message `AuthorizeApplicationRequest`</a>

DEPRECATED: Device claiming that transfers devices between applications is no longer supported and will be removed
in a future version of The Things Stack.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `api_key` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `api_key` | <p>`string.min_len`: `1`</p><p>`string.max_len`: `128`</p> |

### <a name="ttn.lorawan.v3.AuthorizeGatewayRequest">Message `AuthorizeGatewayRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `api_key` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `api_key` | <p>`string.min_len`: `1`</p> |

### <a name="ttn.lorawan.v3.BatchUnclaimEndDevicesRequest">Message `BatchUnclaimEndDevicesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `device_ids` | [`string`](#string) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `device_ids` | <p>`repeated.min_items`: `1`</p><p>`repeated.max_items`: `20`</p><p>`repeated.items.string.max_len`: `36`</p><p>`repeated.items.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.BatchUnclaimEndDevicesResponse">Message `BatchUnclaimEndDevicesResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `failed` | [`BatchUnclaimEndDevicesResponse.FailedEntry`](#ttn.lorawan.v3.BatchUnclaimEndDevicesResponse.FailedEntry) | repeated | End devices that could not be unclaimed. The key is the device ID. |

### <a name="ttn.lorawan.v3.BatchUnclaimEndDevicesResponse.FailedEntry">Message `BatchUnclaimEndDevicesResponse.FailedEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`ErrorDetails`](#ttn.lorawan.v3.ErrorDetails) |  |  |

### <a name="ttn.lorawan.v3.CUPSRedirection">Message `CUPSRedirection`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `target_cups_uri` | [`string`](#string) |  | CUPS URI for LoRa Basics Station CUPS redirection. |
| `current_gateway_key` | [`string`](#string) |  | The key set in the gateway to authenticate itself. |
| `target_cups_trust` | [`bytes`](#bytes) |  | Optional PEM encoded CA Root certificate. If this field is empty, DCS will attempt to dial the Target CUPS server and fetch the CA. |
| `client_tls` | [`CUPSRedirection.ClientTLS`](#ttn.lorawan.v3.CUPSRedirection.ClientTLS) |  | TODO: Support mTLS (https://github.com/TheThingsNetwork/lorawan-stack/issues/137) |
| `auth_token` | [`string`](#string) |  | The Device Claiming Server will fill this field with a The Things Stack API Key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `target_cups_uri` | <p>`string.max_len`: `256`</p><p>`string.pattern`: `^https`</p><p>`string.uri`: `true`</p> |
| `current_gateway_key` | <p>`string.max_len`: `2048`</p> |
| `auth_token` | <p>`string.max_len`: `2048`</p> |

### <a name="ttn.lorawan.v3.CUPSRedirection.ClientTLS">Message `CUPSRedirection.ClientTLS`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cert` | [`bytes`](#bytes) |  | PEM encoded Client Certificate. |
| `key` | [`bytes`](#bytes) |  | PEM encoded Client Private Key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `cert` | <p>`bytes.max_len`: `8192`</p> |
| `key` | <p>`bytes.max_len`: `8192`</p> |

### <a name="ttn.lorawan.v3.ClaimEndDeviceRequest">Message `ClaimEndDeviceRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authenticated_identifiers` | [`ClaimEndDeviceRequest.AuthenticatedIdentifiers`](#ttn.lorawan.v3.ClaimEndDeviceRequest.AuthenticatedIdentifiers) |  | Authenticated identifiers. |
| `qr_code` | [`bytes`](#bytes) |  | Raw QR code contents. |
| `target_application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Application identifiers of the target end device. |
| `target_device_id` | [`string`](#string) |  | End device ID of the target end device. If empty, use the source device ID. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `qr_code` | <p>`bytes.min_len`: `0`</p><p>`bytes.max_len`: `1024`</p> |
| `target_application_ids` | <p>`message.required`: `true`</p> |
| `target_device_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |

### <a name="ttn.lorawan.v3.ClaimEndDeviceRequest.AuthenticatedIdentifiers">Message `ClaimEndDeviceRequest.AuthenticatedIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  | JoinEUI (or AppEUI) of the device to claim. |
| `dev_eui` | [`bytes`](#bytes) |  | DevEUI of the device to claim. |
| `authentication_code` | [`string`](#string) |  | Authentication code to prove ownership. In the LoRa Alliance TR005 specification, this equals the OwnerToken. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |
| `dev_eui` | <p>`bytes.len`: `8`</p> |
| `authentication_code` | <p>`string.pattern`: `^[A-Z0-9]{1,32}$`</p> |

### <a name="ttn.lorawan.v3.ClaimGatewayRequest">Message `ClaimGatewayRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authenticated_identifiers` | [`ClaimGatewayRequest.AuthenticatedIdentifiers`](#ttn.lorawan.v3.ClaimGatewayRequest.AuthenticatedIdentifiers) |  |  |
| `qr_code` | [`bytes`](#bytes) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the target gateway. |
| `target_gateway_id` | [`string`](#string) |  | Gateway ID for the target gateway. This must be a unique value. If this is not set, the target ID for the target gateway will be set to `<gateway-eui>`. |
| `target_gateway_server_address` | [`string`](#string) |  | Target Gateway Server Address for the target gateway. |
| `cups_redirection` | [`CUPSRedirection`](#ttn.lorawan.v3.CUPSRedirection) |  | Parameters to set CUPS redirection for the gateway. |
| `target_frequency_plan_id` | [`string`](#string) |  | Frequency plan ID of the target gateway. This equals the first element of the frequency_plan_ids field. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `qr_code` | <p>`bytes.min_len`: `0`</p><p>`bytes.max_len`: `1024`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |
| `target_gateway_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |
| `target_gateway_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `target_frequency_plan_id` | <p>`string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.ClaimGatewayRequest.AuthenticatedIdentifiers">Message `ClaimGatewayRequest.AuthenticatedIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_eui` | [`bytes`](#bytes) |  |  |
| `authentication_code` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_eui` | <p>`bytes.len`: `8`</p> |
| `authentication_code` | <p>`bytes.max_len`: `2048`</p> |

### <a name="ttn.lorawan.v3.GetClaimStatusResponse">Message `GetClaimStatusResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `home_net_id` | [`bytes`](#bytes) |  |  |
| `home_ns_id` | [`bytes`](#bytes) |  |  |
| `vendor_specific` | [`GetClaimStatusResponse.VendorSpecific`](#ttn.lorawan.v3.GetClaimStatusResponse.VendorSpecific) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |
| `home_net_id` | <p>`bytes.len`: `3`</p> |
| `home_ns_id` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetClaimStatusResponse.VendorSpecific">Message `GetClaimStatusResponse.VendorSpecific`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_unique_identifier` | [`uint32`](#uint32) |  |  |
| `data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Vendor Specific data in JSON format. |

### <a name="ttn.lorawan.v3.GetInfoByGatewayEUIRequest">Message `GetInfoByGatewayEUIRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetInfoByGatewayEUIResponse">Message `GetInfoByGatewayEUIResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `eui` | [`bytes`](#bytes) |  |  |
| `supports_claiming` | [`bool`](#bool) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetInfoByJoinEUIRequest">Message `GetInfoByJoinEUIRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetInfoByJoinEUIResponse">Message `GetInfoByJoinEUIResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `supports_claiming` | [`bool`](#bool) |  | If set, this Join EUI is available for claiming on one of the configured Join Servers. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetInfoByJoinEUIsRequest">Message `GetInfoByJoinEUIsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `requests` | [`GetInfoByJoinEUIRequest`](#ttn.lorawan.v3.GetInfoByJoinEUIRequest) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `requests` | <p>`repeated.max_items`: `20`</p> |

### <a name="ttn.lorawan.v3.GetInfoByJoinEUIsResponse">Message `GetInfoByJoinEUIsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `infos` | [`GetInfoByJoinEUIResponse`](#ttn.lorawan.v3.GetInfoByJoinEUIResponse) | repeated |  |

### <a name="ttn.lorawan.v3.EndDeviceBatchClaimingServer">Service `EndDeviceBatchClaimingServer`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Unclaim` | [`BatchUnclaimEndDevicesRequest`](#ttn.lorawan.v3.BatchUnclaimEndDevicesRequest) | [`BatchUnclaimEndDevicesResponse`](#ttn.lorawan.v3.BatchUnclaimEndDevicesResponse) | Unclaims multiple end devices on an external Join Server. All devices must have the same application ID. Check the response for devices that could not be unclaimed. |
| `GetInfoByJoinEUIs` | [`GetInfoByJoinEUIsRequest`](#ttn.lorawan.v3.GetInfoByJoinEUIsRequest) | [`GetInfoByJoinEUIsResponse`](#ttn.lorawan.v3.GetInfoByJoinEUIsResponse) | Return whether claiming is supported for each Join EUI in a given list. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Unclaim` | `DELETE` | `/api/v3/edcs/claim/{application_ids.application_id}/devices/batch` |  |
| `GetInfoByJoinEUIs` | `POST` | `/api/v3/edcs/claim/info/batch` | `*` |

### <a name="ttn.lorawan.v3.EndDeviceClaimingServer">Service `EndDeviceClaimingServer`</a>

The EndDeviceClaimingServer service configures authorization to claim end devices registered in an application,
and allows clients to claim end devices.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Claim` | [`ClaimEndDeviceRequest`](#ttn.lorawan.v3.ClaimEndDeviceRequest) | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | Claims the end device on a Join Server by claim authentication code or QR code. |
| `Unclaim` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Unclaims the end device on a Join Server. EUIs provided in the request are ignored and the end device is looked up by the given identifiers. |
| `GetInfoByJoinEUI` | [`GetInfoByJoinEUIRequest`](#ttn.lorawan.v3.GetInfoByJoinEUIRequest) | [`GetInfoByJoinEUIResponse`](#ttn.lorawan.v3.GetInfoByJoinEUIResponse) | Return whether claiming is available for a given JoinEUI. |
| `GetClaimStatus` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`GetClaimStatusResponse`](#ttn.lorawan.v3.GetClaimStatusResponse) | Gets the claim status of an end device. EUIs provided in the request are ignored and the end device is looked up by the given identifiers. |
| `AuthorizeApplication` | [`AuthorizeApplicationRequest`](#ttn.lorawan.v3.AuthorizeApplicationRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Authorize the End Device Claiming Server to claim devices registered in the given application. The application identifiers are the source application, where the devices are registered before they are claimed. The API key is used to access the application, find the device, verify the claim request and delete the end device from the source application. DEPRECATED: Device claiming that transfers devices between applications is no longer supported and will be removed in a future version of The Things Stack. |
| `UnauthorizeApplication` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Unauthorize the End Device Claiming Server to claim devices in the given application. This reverts the authorization given with rpc AuthorizeApplication. DEPRECATED: Device claiming that transfers devices between applications is no longer supported and will be removed in a future version of The Things Stack. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Claim` | `POST` | `/api/v3/edcs/claim` | `*` |
| `Unclaim` | `DELETE` | `/api/v3/edcs/claim/{application_ids.application_id}/devices/{device_id}` |  |
| `GetInfoByJoinEUI` | `POST` | `/api/v3/edcs/claim/info` | `*` |
| `GetClaimStatus` | `GET` | `/api/v3/edcs/claim/{application_ids.application_id}/devices/{device_id}` |  |
| `AuthorizeApplication` | `POST` | `/api/v3/edcs/applications/{application_ids.application_id}/authorize` | `*` |
| `UnauthorizeApplication` | `DELETE` | `/api/v3/edcs/applications/{application_id}/authorize` |  |

### <a name="ttn.lorawan.v3.GatewayClaimingServer">Service `GatewayClaimingServer`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Claim` | [`ClaimGatewayRequest`](#ttn.lorawan.v3.ClaimGatewayRequest) | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | Claims a gateway by claim authentication code or QR code and transfers the gateway to the target user. |
| `AuthorizeGateway` | [`AuthorizeGatewayRequest`](#ttn.lorawan.v3.AuthorizeGatewayRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | AuthorizeGateway allows a gateway to be claimed. |
| `UnauthorizeGateway` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | UnauthorizeGateway prevents a gateway from being claimed. |
| `GetInfoByGatewayEUI` | [`GetInfoByGatewayEUIRequest`](#ttn.lorawan.v3.GetInfoByGatewayEUIRequest) | [`GetInfoByGatewayEUIResponse`](#ttn.lorawan.v3.GetInfoByGatewayEUIResponse) | Return whether claiming is available for a given gateway EUI. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Claim` | `POST` | `/api/v3/gcls/claim` | `*` |
| `AuthorizeGateway` | `POST` | `/api/v3/gcls/gateways/{gateway_ids.gateway_id}/authorize` | `*` |
| `UnauthorizeGateway` | `DELETE` | `/api/v3/gcls/gateways/{gateway_id}/authorize` |  |
| `GetInfoByGatewayEUI` | `POST` | `/api/v3/gcls/claim/info` | `*` |

## <a name="ttn/lorawan/v3/devicerepository.proto">File `ttn/lorawan/v3/devicerepository.proto`</a>

### <a name="ttn.lorawan.v3.DecodedMessagePayload">Message `DecodedMessagePayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  |  |
| `warnings` | [`string`](#string) | repeated |  |
| `errors` | [`string`](#string) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `warnings` | <p>`repeated.max_items`: `10`</p><p>`repeated.items.string.max_len`: `100`</p> |
| `errors` | <p>`repeated.max_items`: `10`</p><p>`repeated.items.string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.EncodedMessagePayload">Message `EncodedMessagePayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `f_port` | [`uint32`](#uint32) |  |  |
| `frm_payload` | [`bytes`](#bytes) |  |  |
| `warnings` | [`string`](#string) | repeated |  |
| `errors` | [`string`](#string) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `f_port` | <p>`uint32.lte`: `255`</p> |
| `warnings` | <p>`repeated.max_items`: `10`</p><p>`repeated.items.string.max_len`: `100`</p> |
| `errors` | <p>`repeated.max_items`: `10`</p><p>`repeated.items.string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.EndDeviceBrand">Message `EndDeviceBrand`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `brand_id` | [`string`](#string) |  | Brand identifier, as specified in the Device Repository. |
| `name` | [`string`](#string) |  | Brand name. |
| `private_enterprise_number` | [`uint32`](#uint32) |  | Private Enterprise Number (PEN) assigned by IANA. |
| `organization_unique_identifiers` | [`string`](#string) | repeated | Organization Unique Identifiers (OUI) assigned by IEEE. |
| `lora_alliance_vendor_id` | [`uint32`](#uint32) |  | VendorID managed by the LoRa Alliance, as defined in TR005. |
| `website` | [`string`](#string) |  | Brand website URL. |
| `email` | [`string`](#string) |  | Contact email address. |
| `logo` | [`string`](#string) |  | Path to brand logo. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `organization_unique_identifiers` | <p>`repeated.items.string.pattern`: `[0-9A-F]{6}`</p> |
| `website` | <p>`string.uri`: `true`</p> |
| `email` | <p>`string.email`: `true`</p> |
| `logo` | <p>`string.pattern`: `^$|^(([a-z0-9-]+\/)+)?([a-z0-9_-]+\.)+(png|svg)$`</p> |

### <a name="ttn.lorawan.v3.EndDeviceModel">Message `EndDeviceModel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `brand_id` | [`string`](#string) |  | Brand identifier, as defined in the Device Repository. |
| `model_id` | [`string`](#string) |  | Model identifier, as defined in the Device Repository. |
| `name` | [`string`](#string) |  | Model name, as defined in the Device Repository. |
| `description` | [`string`](#string) |  | Model description. |
| `hardware_versions` | [`EndDeviceModel.HardwareVersion`](#ttn.lorawan.v3.EndDeviceModel.HardwareVersion) | repeated | Available hardware versions. |
| `firmware_versions` | [`EndDeviceModel.FirmwareVersion`](#ttn.lorawan.v3.EndDeviceModel.FirmwareVersion) | repeated | Available firmware versions. |
| `sensors` | [`string`](#string) | repeated | List of sensors included in the device. |
| `dimensions` | [`EndDeviceModel.Dimensions`](#ttn.lorawan.v3.EndDeviceModel.Dimensions) |  | Device dimensions. |
| `weight` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Device weight (gram). |
| `battery` | [`EndDeviceModel.Battery`](#ttn.lorawan.v3.EndDeviceModel.Battery) |  | Device battery information. |
| `operating_conditions` | [`EndDeviceModel.OperatingConditions`](#ttn.lorawan.v3.EndDeviceModel.OperatingConditions) |  | Device operating conditions. |
| `ip_code` | [`string`](#string) |  | Device IP rating code. |
| `key_provisioning` | [`KeyProvisioning`](#ttn.lorawan.v3.KeyProvisioning) | repeated | Supported key provisioning methods. |
| `key_security` | [`KeySecurity`](#ttn.lorawan.v3.KeySecurity) |  | Device key security. |
| `photos` | [`EndDeviceModel.Photos`](#ttn.lorawan.v3.EndDeviceModel.Photos) |  | Device photos. |
| `videos` | [`EndDeviceModel.Videos`](#ttn.lorawan.v3.EndDeviceModel.Videos) |  | Device videos. |
| `product_url` | [`string`](#string) |  | Device information page URL. |
| `datasheet_url` | [`string`](#string) |  | Device datasheet URL. |
| `resellers` | [`EndDeviceModel.Reseller`](#ttn.lorawan.v3.EndDeviceModel.Reseller) | repeated | Reseller URLs. |
| `compliances` | [`EndDeviceModel.Compliances`](#ttn.lorawan.v3.EndDeviceModel.Compliances) |  | List of standards the device is compliant with. |
| `additional_radios` | [`string`](#string) | repeated | List of any additional radios included in the device. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `model_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `sensors` | <p>`repeated.unique`: `true`</p> |
| `key_provisioning` | <p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `key_security` | <p>`enum.defined_only`: `true`</p> |
| `product_url` | <p>`string.uri`: `true`</p> |
| `datasheet_url` | <p>`string.uri`: `true`</p> |
| `additional_radios` | <p>`repeated.unique`: `true`</p> |

### <a name="ttn.lorawan.v3.EndDeviceModel.Battery">Message `EndDeviceModel.Battery`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `replaceable` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  | Whether the device battery can be replaced. |
| `type` | [`string`](#string) |  | Battery type. |

### <a name="ttn.lorawan.v3.EndDeviceModel.Compliances">Message `EndDeviceModel.Compliances`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `safety` | [`EndDeviceModel.Compliances.Compliance`](#ttn.lorawan.v3.EndDeviceModel.Compliances.Compliance) | repeated | List of safety standards the device is compliant with. |
| `radio_equipment` | [`EndDeviceModel.Compliances.Compliance`](#ttn.lorawan.v3.EndDeviceModel.Compliances.Compliance) | repeated | List of radio equipment standards the device is compliant with. |

### <a name="ttn.lorawan.v3.EndDeviceModel.Compliances.Compliance">Message `EndDeviceModel.Compliances.Compliance`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `body` | [`string`](#string) |  |  |
| `norm` | [`string`](#string) |  |  |
| `standard` | [`string`](#string) |  |  |
| `version` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.EndDeviceModel.Dimensions">Message `EndDeviceModel.Dimensions`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `width` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Device width (mm). |
| `height` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Device height (mm). |
| `diameter` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Device diameter (mm). |
| `length` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Device length (mm). |

### <a name="ttn.lorawan.v3.EndDeviceModel.FirmwareVersion">Message `EndDeviceModel.FirmwareVersion`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `version` | [`string`](#string) |  | Firmware version string. |
| `numeric` | [`uint32`](#uint32) |  | Numeric firmware revision number. |
| `supported_hardware_versions` | [`string`](#string) | repeated | Hardware versions supported by this firmware version. |
| `profiles` | [`EndDeviceModel.FirmwareVersion.ProfilesEntry`](#ttn.lorawan.v3.EndDeviceModel.FirmwareVersion.ProfilesEntry) | repeated | Device profiles for each supported region (band). |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `supported_hardware_versions` | <p>`repeated.unique`: `true`</p> |

### <a name="ttn.lorawan.v3.EndDeviceModel.FirmwareVersion.Profile">Message `EndDeviceModel.FirmwareVersion.Profile`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vendor_id` | [`string`](#string) |  | Vendor ID of the profile, as defined in the Device Repository. If this value is set, the profile is loaded from this vendor's folder. If this value is not set, the profile is loaded from the current (end device's) vendor. |
| `profile_id` | [`string`](#string) |  | Profile identifier, as defined in the Device Repository. |
| `lorawan_certified` | [`bool`](#bool) |  | Whether the device is LoRaWAN certified. |
| `codec_id` | [`string`](#string) |  | Payload formatter codec identifier, as defined in the Device Repository. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `vendor_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^$|^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `profile_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^$|^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `codec_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^$|^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.EndDeviceModel.FirmwareVersion.ProfilesEntry">Message `EndDeviceModel.FirmwareVersion.ProfilesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`EndDeviceModel.FirmwareVersion.Profile`](#ttn.lorawan.v3.EndDeviceModel.FirmwareVersion.Profile) |  |  |

### <a name="ttn.lorawan.v3.EndDeviceModel.HardwareVersion">Message `EndDeviceModel.HardwareVersion`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `version` | [`string`](#string) |  | Hardware version string. |
| `numeric` | [`uint32`](#uint32) |  | Numberic hardware revision number. |
| `part_number` | [`string`](#string) |  | Hardware part number. |

### <a name="ttn.lorawan.v3.EndDeviceModel.OperatingConditions">Message `EndDeviceModel.OperatingConditions`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `temperature` | [`EndDeviceModel.OperatingConditions.Limits`](#ttn.lorawan.v3.EndDeviceModel.OperatingConditions.Limits) |  | Temperature operating conditions (Celsius). |
| `relative_humidity` | [`EndDeviceModel.OperatingConditions.Limits`](#ttn.lorawan.v3.EndDeviceModel.OperatingConditions.Limits) |  | Relative humidity operating conditions (Fraction, in range [0, 1]). |

### <a name="ttn.lorawan.v3.EndDeviceModel.OperatingConditions.Limits">Message `EndDeviceModel.OperatingConditions.Limits`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Min value of operating conditions range. |
| `max` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Max value of operating conditions range. |

### <a name="ttn.lorawan.v3.EndDeviceModel.Photos">Message `EndDeviceModel.Photos`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `main` | [`string`](#string) |  | Main device photo. |
| `other` | [`string`](#string) | repeated | List of other device photos. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `main` | <p>`string.pattern`: `^$|^(([a-z0-9-]+\/)+)?([a-z0-9_-]+\.)+(png|jpg|jpeg)$`</p> |
| `other` | <p>`repeated.unique`: `true`</p><p>`repeated.items.string.pattern`: `^$|^(([a-z0-9-]+\/)+)?([a-z0-9_-]+\.)+(png|jpg|jpeg)$`</p> |

### <a name="ttn.lorawan.v3.EndDeviceModel.Reseller">Message `EndDeviceModel.Reseller`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [`string`](#string) |  | Reseller name. |
| `region` | [`string`](#string) | repeated | Reseller regions. |
| `url` | [`string`](#string) |  | Reseller URL. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `url` | <p>`string.uri`: `true`</p> |

### <a name="ttn.lorawan.v3.EndDeviceModel.Videos">Message `EndDeviceModel.Videos`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `main` | [`string`](#string) |  | Link to main device video. |
| `other` | [`string`](#string) | repeated | Links to other device videos. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `main` | <p>`string.pattern`: `^(?:https?:\/\/(?:www\.)?youtu(?:be\.com\/watch\?v=|\.be\/)(?:[\w\-_]*)(?:&(amp;)?[\w\?=]*)?)$|^(?:https?:\/\/(?:www\.)?vimeo\.com\/(?:channels\/(?:\w+\/)?|groups\/([^\/]*)\/videos\/|)(?:\d+)(?:|\/\?))$`</p> |
| `other` | <p>`repeated.unique`: `true`</p><p>`repeated.items.string.pattern`: `^(?:https?:\/\/(?:www\.)?youtu(?:be\.com\/watch\?v=|\.be\/)(?:[\w\-_]*)(?:&(amp;)?[\w\?=]*)?)$|^(?:https?:\/\/(?:www\.)?vimeo\.com\/(?:channels\/(?:\w+\/)?|groups\/([^\/]*)\/videos\/|)(?:\d+)(?:|\/\?))$`</p> |

### <a name="ttn.lorawan.v3.GetEndDeviceBrandRequest">Message `GetEndDeviceBrandRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Application identifiers. |
| `brand_id` | [`string`](#string) |  | Brand identifier, as defined in the Device Repository. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | Field mask paths. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.GetEndDeviceModelRequest">Message `GetEndDeviceModelRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Application identifiers. |
| `brand_id` | [`string`](#string) |  | Brand identifier, as defined in the Device Repository. |
| `model_id` | [`string`](#string) |  | Model identifier, as defined in the Device Repository. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | Field mask paths. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `model_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.GetPayloadFormatterRequest">Message `GetPayloadFormatterRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Application identifiers. |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | End device version information. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | Field mask paths. |

### <a name="ttn.lorawan.v3.GetTemplateRequest">Message `GetTemplateRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Application identifiers. |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | End device version information. Use either EndDeviceVersionIdentifiers or EndDeviceProfileIdentifiers. |
| `end_device_profile_ids` | [`GetTemplateRequest.EndDeviceProfileIdentifiers`](#ttn.lorawan.v3.GetTemplateRequest.EndDeviceProfileIdentifiers) |  | End device profile identifiers. |

### <a name="ttn.lorawan.v3.GetTemplateRequest.EndDeviceProfileIdentifiers">Message `GetTemplateRequest.EndDeviceProfileIdentifiers`</a>

Identifiers to uniquely identify a LoRaWAN end device profile.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vendor_id` | [`uint32`](#uint32) |  | VendorID managed by the LoRa Alliance, as defined in TR005. |
| `vendor_profile_id` | [`uint32`](#uint32) |  | ID of the LoRaWAN end device profile assigned by the vendor. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `vendor_id` | <p>`uint32.gte`: `1`</p> |

### <a name="ttn.lorawan.v3.ListEndDeviceBrandsRequest">Message `ListEndDeviceBrandsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Application identifiers. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `order_by` | [`string`](#string) |  | Order (for pagination) |
| `search` | [`string`](#string) |  | Search for brands matching a query string. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | Field mask paths. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `order_by` | <p>`string.in`: `[ brand_id -brand_id name -name]`</p> |
| `search` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ListEndDeviceBrandsResponse">Message `ListEndDeviceBrandsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `brands` | [`EndDeviceBrand`](#ttn.lorawan.v3.EndDeviceBrand) | repeated |  |

### <a name="ttn.lorawan.v3.ListEndDeviceModelsRequest">Message `ListEndDeviceModelsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  | Application identifiers. |
| `brand_id` | [`string`](#string) |  | List end devices from a specific brand. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `order_by` | [`string`](#string) |  | Order end devices |
| `search` | [`string`](#string) |  | List end devices matching a query string. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | Field mask paths. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^([a-z0-9](?:[-]?[a-z0-9]){2,}|)?$`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `order_by` | <p>`string.in`: `[ brand_id -brand_id model_id -model_id name -name]`</p> |
| `search` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ListEndDeviceModelsResponse">Message `ListEndDeviceModelsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `models` | [`EndDeviceModel`](#ttn.lorawan.v3.EndDeviceModel) | repeated |  |

### <a name="ttn.lorawan.v3.MessagePayloadDecoder">Message `MessagePayloadDecoder`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `formatter` | [`PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter type. |
| `formatter_parameter` | [`string`](#string) |  | Parameter for the formatter, must be set together. |
| `codec_id` | [`string`](#string) |  |  |
| `examples` | [`MessagePayloadDecoder.Example`](#ttn.lorawan.v3.MessagePayloadDecoder.Example) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `formatter` | <p>`enum.defined_only`: `true`</p> |
| `codec_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^([a-z0-9](?:[-]?[a-z0-9]){2,}|)?$`</p> |
| `examples` | <p>`repeated.max_items`: `20`</p> |

### <a name="ttn.lorawan.v3.MessagePayloadDecoder.Example">Message `MessagePayloadDecoder.Example`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `description` | [`string`](#string) |  |  |
| `input` | [`EncodedMessagePayload`](#ttn.lorawan.v3.EncodedMessagePayload) |  |  |
| `output` | [`DecodedMessagePayload`](#ttn.lorawan.v3.DecodedMessagePayload) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `description` | <p>`string.max_len`: `200`</p> |

### <a name="ttn.lorawan.v3.MessagePayloadEncoder">Message `MessagePayloadEncoder`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `formatter` | [`PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter type. |
| `formatter_parameter` | [`string`](#string) |  | Parameter for the formatter, must be set together. |
| `codec_id` | [`string`](#string) |  |  |
| `examples` | [`MessagePayloadEncoder.Example`](#ttn.lorawan.v3.MessagePayloadEncoder.Example) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `formatter` | <p>`enum.defined_only`: `true`</p> |
| `codec_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^([a-z0-9](?:[-]?[a-z0-9]){2,}|)?$`</p> |
| `examples` | <p>`repeated.max_items`: `20`</p> |

### <a name="ttn.lorawan.v3.MessagePayloadEncoder.Example">Message `MessagePayloadEncoder.Example`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `description` | [`string`](#string) |  |  |
| `input` | [`DecodedMessagePayload`](#ttn.lorawan.v3.DecodedMessagePayload) |  |  |
| `output` | [`EncodedMessagePayload`](#ttn.lorawan.v3.EncodedMessagePayload) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `description` | <p>`string.max_len`: `200`</p> |

### <a name="ttn.lorawan.v3.KeyProvisioning">Enum `KeyProvisioning`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `KEY_PROVISIONING_UNKNOWN` | 0 | Unknown Key Provisioning. |
| `KEY_PROVISIONING_CUSTOM` | 1 | Custom Key Provisioning. |
| `KEY_PROVISIONING_JOIN_SERVER` | 2 | Key Provisioning from the Global Join Server. |
| `KEY_PROVISIONING_MANIFEST` | 3 | Key Provisioning from Manifest. |

### <a name="ttn.lorawan.v3.KeySecurity">Enum `KeySecurity`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `KEY_SECURITY_UNKNOWN` | 0 | Unknown key security. |
| `KEY_SECURITY_NONE` | 1 | No key security. |
| `KEY_SECURITY_READ_PROTECTED` | 2 | Read Protected key security. |
| `KEY_SECURITY_SECURE_ELEMENT` | 3 | Key security using the Security Element. |

### <a name="ttn.lorawan.v3.DeviceRepository">Service `DeviceRepository`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListBrands` | [`ListEndDeviceBrandsRequest`](#ttn.lorawan.v3.ListEndDeviceBrandsRequest) | [`ListEndDeviceBrandsResponse`](#ttn.lorawan.v3.ListEndDeviceBrandsResponse) |  |
| `GetBrand` | [`GetEndDeviceBrandRequest`](#ttn.lorawan.v3.GetEndDeviceBrandRequest) | [`EndDeviceBrand`](#ttn.lorawan.v3.EndDeviceBrand) |  |
| `ListModels` | [`ListEndDeviceModelsRequest`](#ttn.lorawan.v3.ListEndDeviceModelsRequest) | [`ListEndDeviceModelsResponse`](#ttn.lorawan.v3.ListEndDeviceModelsResponse) |  |
| `GetModel` | [`GetEndDeviceModelRequest`](#ttn.lorawan.v3.GetEndDeviceModelRequest) | [`EndDeviceModel`](#ttn.lorawan.v3.EndDeviceModel) |  |
| `GetTemplate` | [`GetTemplateRequest`](#ttn.lorawan.v3.GetTemplateRequest) | [`EndDeviceTemplate`](#ttn.lorawan.v3.EndDeviceTemplate) |  |
| `GetUplinkDecoder` | [`GetPayloadFormatterRequest`](#ttn.lorawan.v3.GetPayloadFormatterRequest) | [`MessagePayloadDecoder`](#ttn.lorawan.v3.MessagePayloadDecoder) |  |
| `GetDownlinkDecoder` | [`GetPayloadFormatterRequest`](#ttn.lorawan.v3.GetPayloadFormatterRequest) | [`MessagePayloadDecoder`](#ttn.lorawan.v3.MessagePayloadDecoder) |  |
| `GetDownlinkEncoder` | [`GetPayloadFormatterRequest`](#ttn.lorawan.v3.GetPayloadFormatterRequest) | [`MessagePayloadEncoder`](#ttn.lorawan.v3.MessagePayloadEncoder) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListBrands` | `GET` | `/api/v3/dr/brands` |  |
| `ListBrands` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands` |  |
| `GetBrand` | `GET` | `/api/v3/dr/brands/{brand_id}` |  |
| `GetBrand` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands/{brand_id}` |  |
| `ListModels` | `GET` | `/api/v3/dr/models` |  |
| `ListModels` | `GET` | `/api/v3/dr/brands/{brand_id}/models` |  |
| `ListModels` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/models` |  |
| `ListModels` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands/{brand_id}/models` |  |
| `GetModel` | `GET` | `/api/v3/dr/brands/{brand_id}/models/{model_id}` |  |
| `GetModel` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands/{brand_id}/models/{model_id}` |  |
| `GetTemplate` | `GET` | `/api/v3/dr/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/template` |  |
| `GetTemplate` | `GET` | `/api/v3/dr/vendors/{end_device_profile_ids.vendor_id}/profiles/{end_device_profile_ids.vendor_profile_id}/template` |  |
| `GetTemplate` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/template` |  |
| `GetTemplate` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/vendors/{end_device_profile_ids.vendor_id}/profiles/{end_device_profile_ids.vendor_profile_id}/template` |  |
| `GetUplinkDecoder` | `GET` | `/api/v3/dr/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/formatters/uplink/decoder` |  |
| `GetUplinkDecoder` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/formatters/uplink/decoder` |  |
| `GetDownlinkDecoder` | `GET` | `/api/v3/dr/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/formatters/downlink/decoder` |  |
| `GetDownlinkDecoder` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/formatters/downlink/decoder` |  |
| `GetDownlinkEncoder` | `GET` | `/api/v3/dr/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/formatters/downlink/encoder` |  |
| `GetDownlinkEncoder` | `GET` | `/api/v3/dr/applications/{application_ids.application_id}/brands/{version_ids.brand_id}/models/{version_ids.model_id}/{version_ids.firmware_version}/{version_ids.band_id}/formatters/downlink/encoder` |  |

## <a name="ttn/lorawan/v3/email_messages.proto">File `ttn/lorawan/v3/email_messages.proto`</a>

### <a name="ttn.lorawan.v3.CreateClientEmailMessage">Message `CreateClientEmailMessage`</a>

CreateClientEmailMessage is used as a wrapper for handling the email regarding the create client procedure.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `create_client_request` | [`CreateClientRequest`](#ttn.lorawan.v3.CreateClientRequest) |  |  |
| `api_key` | [`APIKey`](#ttn.lorawan.v3.APIKey) |  |  |

## <a name="ttn/lorawan/v3/end_device.proto">File `ttn/lorawan/v3/end_device.proto`</a>

### <a name="ttn.lorawan.v3.ADRSettings">Message `ADRSettings`</a>

Adaptive Data Rate settings.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `static` | [`ADRSettings.StaticMode`](#ttn.lorawan.v3.ADRSettings.StaticMode) |  |  |
| `dynamic` | [`ADRSettings.DynamicMode`](#ttn.lorawan.v3.ADRSettings.DynamicMode) |  |  |
| `disabled` | [`ADRSettings.DisabledMode`](#ttn.lorawan.v3.ADRSettings.DisabledMode) |  |  |

### <a name="ttn.lorawan.v3.ADRSettings.DisabledMode">Message `ADRSettings.DisabledMode`</a>

Configuration options for cases in which ADR is to be disabled
completely.

### <a name="ttn.lorawan.v3.ADRSettings.DynamicMode">Message `ADRSettings.DynamicMode`</a>

Configuration options for dynamic ADR.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `margin` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | The ADR margin (dB) tells the network server how much margin it should add in ADR requests. A bigger margin is less efficient, but gives a better chance of successful reception. If unset, the default value from Network Server configuration will be used. |
| `min_data_rate_index` | [`DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue) |  | Minimum data rate index. If unset, the default value from Network Server configuration will be used. |
| `max_data_rate_index` | [`DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue) |  | Maximum data rate index. If unset, the default value from Network Server configuration will be used. |
| `min_tx_power_index` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Minimum transmission power index. If unset, the default value from Network Server configuration will be used. |
| `max_tx_power_index` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Maximum transmission power index. If unset, the default value from Network Server configuration will be used. |
| `min_nb_trans` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Minimum number of retransmissions. If unset, the default value from Network Server configuration will be used. |
| `max_nb_trans` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Maximum number of retransmissions. If unset, the default value from Network Server configuration will be used. |
| `channel_steering` | [`ADRSettings.DynamicMode.ChannelSteeringSettings`](#ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `min_tx_power_index` | <p>`uint32.lte`: `15`</p> |
| `max_tx_power_index` | <p>`uint32.lte`: `15`</p> |
| `min_nb_trans` | <p>`uint32.lte`: `3`</p><p>`uint32.gte`: `1`</p> |
| `max_nb_trans` | <p>`uint32.lte`: `3`</p><p>`uint32.gte`: `1`</p> |

### <a name="ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings">Message `ADRSettings.DynamicMode.ChannelSteeringSettings`</a>

EXPERIMENTAL: Channel steering settings.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lora_narrow` | [`ADRSettings.DynamicMode.ChannelSteeringSettings.LoRaNarrowMode`](#ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings.LoRaNarrowMode) |  |  |
| `disabled` | [`ADRSettings.DynamicMode.ChannelSteeringSettings.DisabledMode`](#ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings.DisabledMode) |  |  |

### <a name="ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings.DisabledMode">Message `ADRSettings.DynamicMode.ChannelSteeringSettings.DisabledMode`</a>

Configuration options for cases in which ADR is not supposed to steer the end device
to another set of channels.

### <a name="ttn.lorawan.v3.ADRSettings.DynamicMode.ChannelSteeringSettings.LoRaNarrowMode">Message `ADRSettings.DynamicMode.ChannelSteeringSettings.LoRaNarrowMode`</a>

Configuration options for LoRa narrow channels steering.
The narrow mode attempts to steer the end device towards
using the LoRa modulated, 125kHz bandwidth channels.

### <a name="ttn.lorawan.v3.ADRSettings.StaticMode">Message `ADRSettings.StaticMode`</a>

Configuration options for static ADR.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  | Data rate index to use. |
| `tx_power_index` | [`uint32`](#uint32) |  | Transmission power index to use. |
| `nb_trans` | [`uint32`](#uint32) |  | Number of retransmissions. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `tx_power_index` | <p>`uint32.lte`: `15`</p> |
| `nb_trans` | <p>`uint32.lte`: `15`</p><p>`uint32.gte`: `1`</p> |

### <a name="ttn.lorawan.v3.BatchDeleteEndDevicesRequest">Message `BatchDeleteEndDevicesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `device_ids` | [`string`](#string) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `device_ids` | <p>`repeated.min_items`: `1`</p><p>`repeated.max_items`: `20`</p><p>`repeated.items.string.max_len`: `36`</p><p>`repeated.items.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.BatchGetEndDevicesRequest">Message `BatchGetEndDevicesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `device_ids` | [`string`](#string) | repeated |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the end device fields that should be returned. This mask is applied on all the end devices in the result. See the API reference for which fields can be returned by the different services. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `device_ids` | <p>`repeated.min_items`: `1`</p><p>`repeated.max_items`: `20`</p><p>`repeated.items.string.max_len`: `36`</p><p>`repeated.items.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.BatchUpdateEndDeviceLastSeenRequest">Message `BatchUpdateEndDeviceLastSeenRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `updates` | [`BatchUpdateEndDeviceLastSeenRequest.EndDeviceLastSeenUpdate`](#ttn.lorawan.v3.BatchUpdateEndDeviceLastSeenRequest.EndDeviceLastSeenUpdate) | repeated | The last seen timestamp needs to be set per end device. |

### <a name="ttn.lorawan.v3.BatchUpdateEndDeviceLastSeenRequest.EndDeviceLastSeenUpdate">Message `BatchUpdateEndDeviceLastSeenRequest.EndDeviceLastSeenUpdate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `last_seen_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.BoolValue">Message `BoolValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.ConvertEndDeviceTemplateRequest">Message `ConvertEndDeviceTemplateRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `format_id` | [`string`](#string) |  | ID of the format. |
| `data` | [`bytes`](#bytes) |  | Data to convert. |
| `end_device_version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | End device profile identifiers. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `format_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.CreateEndDeviceRequest">Message `CreateEndDeviceRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device` | [`EndDevice`](#ttn.lorawan.v3.EndDevice) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DevAddrPrefix">Message `DevAddrPrefix`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr` | [`bytes`](#bytes) |  | DevAddr base. |
| `length` | [`uint32`](#uint32) |  | Number of most significant bits from dev_addr that are used as prefix. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr` | <p>`bytes.len`: `4`</p> |

### <a name="ttn.lorawan.v3.EndDevice">Message `EndDevice`</a>

Defines an End Device registration and its state on the network.
The persistence of the EndDevice is divided between the Network Server, Application Server and Join Server.
SDKs are responsible for combining (if desired) the three.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `name` | [`string`](#string) |  | Friendly name of the device. Stored in Entity Registry. |
| `description` | [`string`](#string) |  | Description of the device. Stored in Entity Registry. |
| `attributes` | [`EndDevice.AttributesEntry`](#ttn.lorawan.v3.EndDevice.AttributesEntry) | repeated | Key-value attributes for this end device. Typically used for organizing end devices or for storing integration-specific data. Stored in Entity Registry. |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | Version Identifiers. Stored in Entity Registry, Network Server and Application Server. |
| `service_profile_id` | [`string`](#string) |  | Default service profile. Stored in Entity Registry. |
| `network_server_address` | [`string`](#string) |  | The address of the Network Server where this device is supposed to be registered. Stored in Entity Registry and Join Server. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| `network_server_kek_label` | [`string`](#string) |  | The KEK label of the Network Server to use for wrapping network session keys. Stored in Join Server. |
| `application_server_address` | [`string`](#string) |  | The address of the Application Server where this device is supposed to be registered. Stored in Entity Registry and Join Server. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| `application_server_kek_label` | [`string`](#string) |  | The KEK label of the Application Server to use for wrapping the application session key. Stored in Join Server. |
| `application_server_id` | [`string`](#string) |  | The AS-ID of the Application Server to use. Stored in Join Server. |
| `join_server_address` | [`string`](#string) |  | The address of the Join Server where this device is supposed to be registered. Stored in Entity Registry. The typical format of the address is "host:port". If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| `locations` | [`EndDevice.LocationsEntry`](#ttn.lorawan.v3.EndDevice.LocationsEntry) | repeated | Location of the device. Stored in Entity Registry. |
| `picture` | [`Picture`](#ttn.lorawan.v3.Picture) |  | Stored in Entity Registry. |
| `supports_class_b` | [`bool`](#bool) |  | Whether the device supports class B. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `supports_class_c` | [`bool`](#bool) |  | Whether the device supports class C. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `lorawan_version` | [`MACVersion`](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `lorawan_phy_version` | [`PHYVersion`](#ttn.lorawan.v3.PHYVersion) |  | LoRaWAN PHY version. Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `frequency_plan_id` | [`string`](#string) |  | ID of the frequency plan used by this device. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `min_frequency` | [`uint64`](#uint64) |  | Minimum frequency the device is capable of using (Hz). Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `max_frequency` | [`uint64`](#uint64) |  | Maximum frequency the device is capable of using (Hz). Stored in Network Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `supports_join` | [`bool`](#bool) |  | The device supports join (it's OTAA). Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `resets_join_nonces` | [`bool`](#bool) |  | Whether the device resets the join and dev nonces (not LoRaWAN compliant). Stored in Join Server. Copied on creation from template identified by version_ids, if any or from the home Network Server device profile, if any. |
| `root_keys` | [`RootKeys`](#ttn.lorawan.v3.RootKeys) |  | Device root keys. Stored in Join Server. |
| `net_id` | [`bytes`](#bytes) |  | Home NetID. Stored in Join Server. |
| `mac_settings` | [`MACSettings`](#ttn.lorawan.v3.MACSettings) |  | Settings for how the Network Server handles MAC layer for this device. Stored in Network Server. |
| `mac_state` | [`MACState`](#ttn.lorawan.v3.MACState) |  | MAC state of the device. Stored in Network Server. |
| `pending_mac_state` | [`MACState`](#ttn.lorawan.v3.MACState) |  | Pending MAC state of the device. Stored in Network Server. |
| `session` | [`Session`](#ttn.lorawan.v3.Session) |  | Current session of the device. Stored in Network Server and Application Server. |
| `pending_session` | [`Session`](#ttn.lorawan.v3.Session) |  | Pending session. Stored in Network Server and Application Server until RekeyInd is received. |
| `last_dev_nonce` | [`uint32`](#uint32) |  | Last DevNonce used. This field is only used for devices using LoRaWAN version 1.1 and later. Stored in Join Server. |
| `used_dev_nonces` | [`uint32`](#uint32) | repeated | Used DevNonces sorted in ascending order. This field is only used for devices using LoRaWAN versions preceding 1.1. Stored in Join Server. |
| `last_join_nonce` | [`uint32`](#uint32) |  | Last JoinNonce/AppNonce(for devices using LoRaWAN versions preceding 1.1) used. Stored in Join Server. |
| `last_rj_count_0` | [`uint32`](#uint32) |  | Last Rejoin counter value used (type 0/2). Stored in Join Server. |
| `last_rj_count_1` | [`uint32`](#uint32) |  | Last Rejoin counter value used (type 1). Stored in Join Server. |
| `last_dev_status_received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time when last DevStatus MAC command was received. Stored in Network Server. |
| `power_state` | [`PowerState`](#ttn.lorawan.v3.PowerState) |  | The power state of the device; whether it is battery-powered or connected to an external power source. Received via the DevStatus MAC command at status_received_at. Stored in Network Server. |
| `battery_percentage` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Latest-known battery percentage of the device. Received via the DevStatus MAC command at last_dev_status_received_at or earlier. Stored in Network Server. |
| `downlink_margin` | [`int32`](#int32) |  | Demodulation signal-to-noise ratio (dB). Received via the DevStatus MAC command at last_dev_status_received_at. Stored in Network Server. |
| `queued_application_downlinks` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) | repeated | Queued Application downlink messages. Stored in Application Server, which sets them on the Network Server. This field is deprecated and is always set equal to session.queued_application_downlinks. |
| `formatters` | [`MessagePayloadFormatters`](#ttn.lorawan.v3.MessagePayloadFormatters) |  | The payload formatters for this end device. Stored in Application Server. Copied on creation from template identified by version_ids. |
| `provisioner_id` | [`string`](#string) |  | ID of the provisioner. Stored in Join Server. |
| `provisioning_data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Vendor-specific provisioning data. Stored in Join Server. |
| `multicast` | [`bool`](#bool) |  | Indicates whether this device represents a multicast group. |
| `claim_authentication_code` | [`EndDeviceAuthenticationCode`](#ttn.lorawan.v3.EndDeviceAuthenticationCode) |  | Authentication code to claim ownership of the end device. From TTS v3.21.0 this field is stored in the Identity Server. For TTS versions < 3.21.0, this field is stored in the Join Server. The value stored on the Identity Server takes precedence. |
| `skip_payload_crypto` | [`bool`](#bool) |  | Skip decryption of uplink payloads and encryption of downlink payloads. This field is deprecated, use skip_payload_crypto_override instead. |
| `skip_payload_crypto_override` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  | Skip decryption of uplink payloads and encryption of downlink payloads. This field overrides the application-level setting. |
| `activated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp when the device has been activated. Stored in the Entity Registry. This field is set by the Application Server when an end device sends its first uplink. The Application Server will use the field in order to avoid repeated calls to the Entity Registry. The field cannot be unset once set. |
| `last_seen_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp when a device uplink has been last observed. This field is set by the Application Server and stored in the Identity Server. |
| `serial_number` | [`string`](#string) |  |  |
| `lora_alliance_profile_ids` | [`LoRaAllianceProfileIdentifiers`](#ttn.lorawan.v3.LoRaAllianceProfileIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `description` | <p>`string.max_len`: `2000`</p> |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |
| `service_profile_id` | <p>`string.max_len`: `64`</p> |
| `network_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `network_server_kek_label` | <p>`string.max_len`: `2048`</p> |
| `application_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `application_server_kek_label` | <p>`string.max_len`: `2048`</p> |
| `application_server_id` | <p>`string.max_len`: `100`</p> |
| `join_server_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `locations` | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `lorawan_version` | <p>`enum.defined_only`: `true`</p> |
| `lorawan_phy_version` | <p>`enum.defined_only`: `true`</p> |
| `frequency_plan_id` | <p>`string.max_len`: `64`</p> |
| `net_id` | <p>`bytes.len`: `3`</p> |
| `power_state` | <p>`enum.defined_only`: `true`</p> |
| `battery_percentage` | <p>`float.lte`: `1`</p><p>`float.gte`: `0`</p> |
| `provisioner_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |
| `serial_number` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.EndDevice.AttributesEntry">Message `EndDevice.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.EndDevice.LocationsEntry">Message `EndDevice.LocationsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`Location`](#ttn.lorawan.v3.Location) |  |  |

### <a name="ttn.lorawan.v3.EndDeviceAuthenticationCode">Message `EndDeviceAuthenticationCode`</a>

Authentication code for end devices.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`string`](#string) |  |  |
| `valid_from` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `valid_to` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`string.pattern`: `^[a-zA-Z0-9]{1,32}$`</p> |

### <a name="ttn.lorawan.v3.EndDeviceTemplate">Message `EndDeviceTemplate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device` | [`EndDevice`](#ttn.lorawan.v3.EndDevice) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `mapping_key` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device` | <p>`message.required`: `true`</p> |
| `mapping_key` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.EndDeviceTemplateFormat">Message `EndDeviceTemplateFormat`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [`string`](#string) |  |  |
| `description` | [`string`](#string) |  |  |
| `file_extensions` | [`string`](#string) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `name` | <p>`string.max_len`: `100`</p> |
| `description` | <p>`string.max_len`: `200`</p> |
| `file_extensions` | <p>`repeated.max_items`: `100`</p><p>`repeated.unique`: `true`</p><p>`repeated.items.string.pattern`: `^(?:\.[a-z0-9]{1,16}){1,2}$`</p> |

### <a name="ttn.lorawan.v3.EndDeviceTemplateFormats">Message `EndDeviceTemplateFormats`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `formats` | [`EndDeviceTemplateFormats.FormatsEntry`](#ttn.lorawan.v3.EndDeviceTemplateFormats.FormatsEntry) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `formats` | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.EndDeviceTemplateFormats.FormatsEntry">Message `EndDeviceTemplateFormats.FormatsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`EndDeviceTemplateFormat`](#ttn.lorawan.v3.EndDeviceTemplateFormat) |  |  |

### <a name="ttn.lorawan.v3.EndDeviceVersion">Message `EndDeviceVersion`</a>

Template for creating end devices.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | Version identifiers. |
| `lorawan_version` | [`MACVersion`](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. |
| `lorawan_phy_version` | [`PHYVersion`](#ttn.lorawan.v3.PHYVersion) |  | LoRaWAN PHY version. |
| `frequency_plan_id` | [`string`](#string) |  | ID of the frequency plan used by this device. |
| `photos` | [`string`](#string) | repeated | Photos contains file names of device photos. |
| `supports_class_b` | [`bool`](#bool) |  | Whether the device supports class B. |
| `supports_class_c` | [`bool`](#bool) |  | Whether the device supports class C. |
| `default_mac_settings` | [`MACSettings`](#ttn.lorawan.v3.MACSettings) |  | Default MAC layer settings of the device. |
| `min_frequency` | [`uint64`](#uint64) |  | Minimum frequency the device is capable of using (Hz). |
| `max_frequency` | [`uint64`](#uint64) |  | Maximum frequency the device is capable of using (Hz). |
| `supports_join` | [`bool`](#bool) |  | The device supports join (it's OTAA). |
| `resets_join_nonces` | [`bool`](#bool) |  | Whether the device resets the join and dev nonces (not LoRaWAN compliant). |
| `default_formatters` | [`MessagePayloadFormatters`](#ttn.lorawan.v3.MessagePayloadFormatters) |  | Default formatters defining the payload formats for this end device. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `lorawan_version` | <p>`enum.defined_only`: `true`</p> |
| `lorawan_phy_version` | <p>`enum.defined_only`: `true`</p> |
| `frequency_plan_id` | <p>`string.max_len`: `64`</p> |
| `photos` | <p>`repeated.max_items`: `10`</p> |
| `default_formatters` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.EndDevices">Message `EndDevices`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_devices` | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | repeated |  |

### <a name="ttn.lorawan.v3.GetEndDeviceIdentifiersForEUIsRequest">Message `GetEndDeviceIdentifiersForEUIsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `dev_eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |
| `dev_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetEndDeviceRequest">Message `GetEndDeviceRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the end device fields that should be returned. See the API reference for which fields can be returned by the different services. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ListEndDevicesRequest">Message `ListEndDevicesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the end device fields that should be returned. See the API reference for which fields can be returned by the different services. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `order` | <p>`string.in`: `[ device_id -device_id join_eui -join_eui dev_eui -dev_eui name -name description -description created_at -created_at last_seen_at -last_seen_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.MACParameters">Message `MACParameters`</a>

MACParameters represent the parameters of the device's MAC layer (active or desired).
This is used internally by the Network Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_eirp` | [`float`](#float) |  | Maximum EIRP (dBm). |
| `adr_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  | ADR: data rate index to use. |
| `adr_tx_power_index` | [`uint32`](#uint32) |  | ADR: transmission power index to use. |
| `adr_nb_trans` | [`uint32`](#uint32) |  | ADR: number of retransmissions. |
| `adr_ack_limit` | [`uint32`](#uint32) |  | ADR: number of messages to wait before setting ADRAckReq. This field is deprecated, use adr_ack_limit_exponent instead. |
| `adr_ack_delay` | [`uint32`](#uint32) |  | ADR: number of messages to wait after setting ADRAckReq and before changing TxPower or DataRate. This field is deprecated, use adr_ack_delay_exponent instead. |
| `rx1_delay` | [`RxDelay`](#ttn.lorawan.v3.RxDelay) |  | Rx1 delay (Rx2 delay is Rx1 delay + 1 second). |
| `rx1_data_rate_offset` | [`DataRateOffset`](#ttn.lorawan.v3.DataRateOffset) |  | Data rate offset for Rx1. |
| `rx2_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  | Data rate index for Rx2. |
| `rx2_frequency` | [`uint64`](#uint64) |  | Frequency for Rx2 (Hz). |
| `max_duty_cycle` | [`AggregatedDutyCycle`](#ttn.lorawan.v3.AggregatedDutyCycle) |  | Maximum uplink duty cycle (of all channels). |
| `rejoin_time_periodicity` | [`RejoinTimeExponent`](#ttn.lorawan.v3.RejoinTimeExponent) |  | Time within which a rejoin-request must be sent. |
| `rejoin_count_periodicity` | [`RejoinCountExponent`](#ttn.lorawan.v3.RejoinCountExponent) |  | Message count within which a rejoin-request must be sent. |
| `ping_slot_frequency` | [`uint64`](#uint64) |  | Frequency of the class B ping slot (Hz). |
| `ping_slot_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  | Data rate index of the class B ping slot. This field is deprecated, use ping_slot_data_rate_index_value instead. |
| `beacon_frequency` | [`uint64`](#uint64) |  | Frequency of the class B beacon (Hz). |
| `channels` | [`MACParameters.Channel`](#ttn.lorawan.v3.MACParameters.Channel) | repeated | Configured uplink channels and optionally Rx1 frequency. |
| `uplink_dwell_time` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether uplink dwell time is set (400ms). If unset, then the value is either unknown or irrelevant(Network Server cannot modify it). |
| `downlink_dwell_time` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether downlink dwell time is set (400ms). If unset, then the value is either unknown or irrelevant(Network Server cannot modify it). |
| `adr_ack_limit_exponent` | [`ADRAckLimitExponentValue`](#ttn.lorawan.v3.ADRAckLimitExponentValue) |  | ADR: number of messages to wait before setting ADRAckReq. |
| `adr_ack_delay_exponent` | [`ADRAckDelayExponentValue`](#ttn.lorawan.v3.ADRAckDelayExponentValue) |  | ADR: number of messages to wait after setting ADRAckReq and before changing TxPower or DataRate. |
| `ping_slot_data_rate_index_value` | [`DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue) |  | Data rate index of the class B ping slot. |
| `relay` | [`RelayParameters`](#ttn.lorawan.v3.RelayParameters) |  | Relay parameters. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `adr_data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `adr_tx_power_index` | <p>`uint32.lte`: `15`</p> |
| `adr_nb_trans` | <p>`uint32.lte`: `15`</p> |
| `rx1_delay` | <p>`enum.defined_only`: `true`</p> |
| `rx1_data_rate_offset` | <p>`enum.defined_only`: `true`</p> |
| `rx2_data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `rx2_frequency` | <p>`uint64.gte`: `100000`</p> |
| `max_duty_cycle` | <p>`enum.defined_only`: `true`</p> |
| `rejoin_time_periodicity` | <p>`enum.defined_only`: `true`</p> |
| `rejoin_count_periodicity` | <p>`enum.defined_only`: `true`</p> |
| `ping_slot_frequency` | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p> |
| `beacon_frequency` | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p> |
| `channels` | <p>`repeated.min_items`: `1`</p> |

### <a name="ttn.lorawan.v3.MACParameters.Channel">Message `MACParameters.Channel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uplink_frequency` | [`uint64`](#uint64) |  | Uplink frequency of the channel (Hz). |
| `downlink_frequency` | [`uint64`](#uint64) |  | Downlink frequency of the channel (Hz). |
| `min_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  | Index of the minimum data rate for uplink. |
| `max_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  | Index of the maximum data rate for uplink. |
| `enable_uplink` | [`bool`](#bool) |  | Channel can be used by device for uplink. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `uplink_frequency` | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p> |
| `downlink_frequency` | <p>`uint64.gte`: `100000`</p> |
| `min_data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `max_data_rate_index` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACSettings">Message `MACSettings`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_b_timeout` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Maximum delay for the device to answer a MAC request or a confirmed downlink frame. If unset, the default value from Network Server configuration will be used. |
| `ping_slot_periodicity` | [`PingSlotPeriodValue`](#ttn.lorawan.v3.PingSlotPeriodValue) |  | Periodicity of the class B ping slot. If unset, the default value from Network Server configuration will be used. |
| `ping_slot_data_rate_index` | [`DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue) |  | Data rate index of the class B ping slot. If unset, the default value from Network Server configuration will be used. |
| `ping_slot_frequency` | [`ZeroableFrequencyValue`](#ttn.lorawan.v3.ZeroableFrequencyValue) |  | Frequency of the class B ping slot (Hz). If unset, the default value from Network Server configuration will be used. |
| `beacon_frequency` | [`ZeroableFrequencyValue`](#ttn.lorawan.v3.ZeroableFrequencyValue) |  | Frequency of the class B beacon (Hz). If unset, the default value from Network Server configuration will be used. |
| `class_c_timeout` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Maximum delay for the device to answer a MAC request or a confirmed downlink frame. If unset, the default value from Network Server configuration will be used. |
| `rx1_delay` | [`RxDelayValue`](#ttn.lorawan.v3.RxDelayValue) |  | Class A Rx1 delay. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `rx1_data_rate_offset` | [`DataRateOffsetValue`](#ttn.lorawan.v3.DataRateOffsetValue) |  | Rx1 data rate offset. If unset, the default value from Network Server configuration will be used. |
| `rx2_data_rate_index` | [`DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue) |  | Data rate index for Rx2. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `rx2_frequency` | [`FrequencyValue`](#ttn.lorawan.v3.FrequencyValue) |  | Frequency for Rx2 (Hz). If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `factory_preset_frequencies` | [`uint64`](#uint64) | repeated | List of factory-preset frequencies. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `max_duty_cycle` | [`AggregatedDutyCycleValue`](#ttn.lorawan.v3.AggregatedDutyCycleValue) |  | Maximum uplink duty cycle (of all channels). |
| `supports_32_bit_f_cnt` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether the device supports 32-bit frame counters. If unset, the default value from Network Server configuration will be used. |
| `use_adr` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether the Network Server should use ADR for the device. This field is deprecated, use adr_settings instead. |
| `adr_margin` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | The ADR margin (dB) tells the network server how much margin it should add in ADR requests. A bigger margin is less efficient, but gives a better chance of successful reception. This field is deprecated, use adr_settings.dynamic.margin instead. |
| `resets_f_cnt` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether the device resets the frame counters (not LoRaWAN compliant). If unset, the default value from Network Server configuration will be used. |
| `status_time_periodicity` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | The interval after which a DevStatusReq MACCommand shall be sent. If unset, the default value from Network Server configuration will be used. |
| `status_count_periodicity` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | Number of uplink messages after which a DevStatusReq MACCommand shall be sent. If unset, the default value from Network Server configuration will be used. |
| `desired_rx1_delay` | [`RxDelayValue`](#ttn.lorawan.v3.RxDelayValue) |  | The Rx1 delay Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `desired_rx1_data_rate_offset` | [`DataRateOffsetValue`](#ttn.lorawan.v3.DataRateOffsetValue) |  | The Rx1 data rate offset Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from Network Server configuration will be used. |
| `desired_rx2_data_rate_index` | [`DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue) |  | The Rx2 data rate index Network Server should configure device to use via MAC commands or Join-Accept. If unset, the default value from frequency plan, Network Server configuration or regional parameters specification will be used. |
| `desired_rx2_frequency` | [`FrequencyValue`](#ttn.lorawan.v3.FrequencyValue) |  | The Rx2 frequency index Network Server should configure device to use via MAC commands. If unset, the default value from frequency plan, Network Server configuration or regional parameters specification will be used. |
| `desired_max_duty_cycle` | [`AggregatedDutyCycleValue`](#ttn.lorawan.v3.AggregatedDutyCycleValue) |  | The maximum uplink duty cycle (of all channels) Network Server should configure device to use via MAC commands. If unset, the default value from Network Server configuration will be used. |
| `desired_adr_ack_limit_exponent` | [`ADRAckLimitExponentValue`](#ttn.lorawan.v3.ADRAckLimitExponentValue) |  | The ADR ACK limit Network Server should configure device to use via MAC commands. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `desired_adr_ack_delay_exponent` | [`ADRAckDelayExponentValue`](#ttn.lorawan.v3.ADRAckDelayExponentValue) |  | The ADR ACK delay Network Server should configure device to use via MAC commands. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `desired_ping_slot_data_rate_index` | [`DataRateIndexValue`](#ttn.lorawan.v3.DataRateIndexValue) |  | The data rate index of the class B ping slot Network Server should configure device to use via MAC commands. If unset, the default value from Network Server configuration will be used. |
| `desired_ping_slot_frequency` | [`ZeroableFrequencyValue`](#ttn.lorawan.v3.ZeroableFrequencyValue) |  | The frequency of the class B ping slot (Hz) Network Server should configure device to use via MAC commands. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `desired_beacon_frequency` | [`ZeroableFrequencyValue`](#ttn.lorawan.v3.ZeroableFrequencyValue) |  | The frequency of the class B beacon (Hz) Network Server should configure device to use via MAC commands. If unset, the default value from Network Server configuration will be used. |
| `desired_max_eirp` | [`DeviceEIRPValue`](#ttn.lorawan.v3.DeviceEIRPValue) |  | Maximum EIRP (dBm). If unset, the default value from regional parameters specification will be used. |
| `class_b_c_downlink_interval` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | The minimum duration passed before a network-initiated(e.g. Class B or C) downlink following an arbitrary downlink. |
| `uplink_dwell_time` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether uplink dwell time is set (400ms). If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `downlink_dwell_time` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether downlink dwell time is set (400ms). If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `adr` | [`ADRSettings`](#ttn.lorawan.v3.ADRSettings) |  | Adaptive Data Rate settings. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `schedule_downlinks` | [`BoolValue`](#ttn.lorawan.v3.BoolValue) |  | Whether or not downlink messages should be scheduled. This option can be used in order to disable any downlink interaction with the end device. It will affect all types of downlink messages: data and MAC downlinks, and join accepts. |
| `relay` | [`RelayParameters`](#ttn.lorawan.v3.RelayParameters) |  | The relay parameters the end device is using. If unset, the default value from Network Server configuration or regional parameters specification will be used. |
| `desired_relay` | [`RelayParameters`](#ttn.lorawan.v3.RelayParameters) |  | The relay parameters the Network Server should configure device to use via MAC commands. If unset, the default value from Network Server configuration or regional parameters specification will be used. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `factory_preset_frequencies` | <p>`repeated.max_items`: `96`</p> |

### <a name="ttn.lorawan.v3.MACState">Message `MACState`</a>

MACState represents the state of MAC layer of the device.
MACState is reset on each join for OTAA or ResetInd for ABP devices.
This is used internally by the Network Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `current_parameters` | [`MACParameters`](#ttn.lorawan.v3.MACParameters) |  | Current LoRaWAN MAC parameters. |
| `desired_parameters` | [`MACParameters`](#ttn.lorawan.v3.MACParameters) |  | Desired LoRaWAN MAC parameters. |
| `device_class` | [`Class`](#ttn.lorawan.v3.Class) |  | Currently active LoRaWAN device class - Device class is A by default - If device sets ClassB bit in uplink, this will be set to B - If device sent DeviceModeInd MAC message, this will be set to that value |
| `lorawan_version` | [`MACVersion`](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN MAC version. |
| `last_confirmed_downlink_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time when the last confirmed downlink message or MAC command was scheduled. |
| `last_dev_status_f_cnt_up` | [`uint32`](#uint32) |  | Frame counter value of last uplink containing DevStatusAns. |
| `ping_slot_periodicity` | [`PingSlotPeriodValue`](#ttn.lorawan.v3.PingSlotPeriodValue) |  | Periodicity of the class B ping slot. |
| `pending_application_downlink` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  | A confirmed application downlink, for which an acknowledgment is expected to arrive. |
| `queued_responses` | [`MACCommand`](#ttn.lorawan.v3.MACCommand) | repeated | Queued MAC responses. Regenerated on each uplink. |
| `pending_requests` | [`MACCommand`](#ttn.lorawan.v3.MACCommand) | repeated | Pending MAC requests(i.e. sent requests, for which no response has been received yet). Regenerated on each downlink. |
| `queued_join_accept` | [`MACState.JoinAccept`](#ttn.lorawan.v3.MACState.JoinAccept) |  | Queued join-accept. Set each time a (re-)join request accept is received from Join Server and removed each time a downlink is scheduled. |
| `pending_join_request` | [`MACState.JoinRequest`](#ttn.lorawan.v3.MACState.JoinRequest) |  | Pending join request. Set each time a join-accept is scheduled and removed each time an uplink is received from the device. |
| `rx_windows_available` | [`bool`](#bool) |  | Whether or not Rx windows are expected to be open. Set to true every time an uplink is received. Set to false every time a successful downlink scheduling attempt is made. |
| `recent_uplinks` | [`MACState.UplinkMessage`](#ttn.lorawan.v3.MACState.UplinkMessage) | repeated | Recent data uplink messages sorted by time. The number of messages stored may depend on configuration. |
| `recent_downlinks` | [`MACState.DownlinkMessage`](#ttn.lorawan.v3.MACState.DownlinkMessage) | repeated | Recent data downlink messages sorted by time. The number of messages stored may depend on configuration. |
| `last_network_initiated_downlink_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time when the last network-initiated downlink message was scheduled. |
| `rejected_adr_data_rate_indexes` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) | repeated | ADR Data rate index values rejected by the device. Reset each time `current_parameters.channels` change. Elements are sorted in ascending order. |
| `rejected_adr_tx_power_indexes` | [`uint32`](#uint32) | repeated | ADR TX output power index values rejected by the device. Elements are sorted in ascending order. |
| `rejected_frequencies` | [`uint64`](#uint64) | repeated | Frequencies rejected by the device. |
| `last_downlink_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time when the last downlink message was scheduled. |
| `rejected_data_rate_ranges` | [`MACState.RejectedDataRateRangesEntry`](#ttn.lorawan.v3.MACState.RejectedDataRateRangesEntry) | repeated | Data rate ranges rejected by the device per frequency. |
| `last_adr_change_f_cnt_up` | [`uint32`](#uint32) |  | Frame counter of uplink, which confirmed the last ADR parameter change. |
| `recent_mac_command_identifiers` | [`MACCommandIdentifier`](#ttn.lorawan.v3.MACCommandIdentifier) | repeated | MAC command identifiers sent by the end device in the last received uplink. The Network Server may choose to store only certain types of MAC command identifiers in the underlying implementation. |
| `pending_relay_downlink` | [`RelayForwardDownlinkReq`](#ttn.lorawan.v3.RelayForwardDownlinkReq) |  | Pending relay downlink contents. The pending downlink will be scheduled to the relay in either Rx1 or Rx2. The pending downlink will be cleared after the scheduling attempt. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `current_parameters` | <p>`message.required`: `true`</p> |
| `desired_parameters` | <p>`message.required`: `true`</p> |
| `device_class` | <p>`enum.defined_only`: `true`</p> |
| `lorawan_version` | <p>`enum.defined_only`: `true`</p> |
| `rejected_adr_data_rate_indexes` | <p>`repeated.max_items`: `15`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `rejected_adr_tx_power_indexes` | <p>`repeated.max_items`: `15`</p><p>`repeated.items.uint32.lte`: `15`</p> |
| `rejected_frequencies` | <p>`repeated.items.uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.MACState.DataRateRange">Message `MACState.DataRateRange`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `max_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `min_data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `max_data_rate_index` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACState.DataRateRanges">Message `MACState.DataRateRanges`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ranges` | [`MACState.DataRateRange`](#ttn.lorawan.v3.MACState.DataRateRange) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ranges` | <p>`repeated.min_items`: `1`</p> |

### <a name="ttn.lorawan.v3.MACState.DownlinkMessage">Message `MACState.DownlinkMessage`</a>

A minimal DownlinkMessage definition which is binary compatible with the top level DownlinkMessage message.
Used for type safe recent downlink storage.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payload` | [`MACState.DownlinkMessage.Message`](#ttn.lorawan.v3.MACState.DownlinkMessage.Message) |  |  |
| `correlation_ids` | [`string`](#string) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.MACState.DownlinkMessage.Message">Message `MACState.DownlinkMessage.Message`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `m_hdr` | [`MACState.DownlinkMessage.Message.MHDR`](#ttn.lorawan.v3.MACState.DownlinkMessage.Message.MHDR) |  |  |
| `mac_payload` | [`MACState.DownlinkMessage.Message.MACPayload`](#ttn.lorawan.v3.MACState.DownlinkMessage.Message.MACPayload) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `m_hdr` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.MACState.DownlinkMessage.Message.MACPayload">Message `MACState.DownlinkMessage.Message.MACPayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `f_port` | [`uint32`](#uint32) |  |  |
| `full_f_cnt` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `f_port` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.MACState.DownlinkMessage.Message.MHDR">Message `MACState.DownlinkMessage.Message.MHDR`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `m_type` | [`MType`](#ttn.lorawan.v3.MType) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `m_type` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACState.JoinAccept">Message `MACState.JoinAccept`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payload` | [`bytes`](#bytes) |  | Payload of the join-accept received from Join Server. |
| `request` | [`MACState.JoinRequest`](#ttn.lorawan.v3.MACState.JoinRequest) |  |  |
| `keys` | [`SessionKeys`](#ttn.lorawan.v3.SessionKeys) |  | Network session keys associated with the join. |
| `correlation_ids` | [`string`](#string) | repeated |  |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `net_id` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `payload` | <p>`bytes.min_len`: `17`</p><p>`bytes.max_len`: `33`</p> |
| `request` | <p>`message.required`: `true`</p> |
| `keys` | <p>`message.required`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `net_id` | <p>`bytes.len`: `3`</p> |

### <a name="ttn.lorawan.v3.MACState.JoinRequest">Message `MACState.JoinRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `downlink_settings` | [`DLSettings`](#ttn.lorawan.v3.DLSettings) |  |  |
| `rx_delay` | [`RxDelay`](#ttn.lorawan.v3.RxDelay) |  |  |
| `cf_list` | [`CFList`](#ttn.lorawan.v3.CFList) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `downlink_settings` | <p>`message.required`: `true`</p> |
| `rx_delay` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACState.RejectedDataRateRangesEntry">Message `MACState.RejectedDataRateRangesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`uint64`](#uint64) |  |  |
| `value` | [`MACState.DataRateRanges`](#ttn.lorawan.v3.MACState.DataRateRanges) |  |  |

### <a name="ttn.lorawan.v3.MACState.UplinkMessage">Message `MACState.UplinkMessage`</a>

A minimal UplinkMessage definition which is binary compatible with the top level UplinkMessage message.
Used for type safe recent uplink storage.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payload` | [`Message`](#ttn.lorawan.v3.Message) |  |  |
| `settings` | [`MACState.UplinkMessage.TxSettings`](#ttn.lorawan.v3.MACState.UplinkMessage.TxSettings) |  |  |
| `rx_metadata` | [`MACState.UplinkMessage.RxMetadata`](#ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata) | repeated |  |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `correlation_ids` | [`string`](#string) | repeated |  |
| `device_channel_index` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `payload` | <p>`message.required`: `true`</p> |
| `settings` | <p>`message.required`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |
| `device_channel_index` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata">Message `MACState.UplinkMessage.RxMetadata`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `channel_rssi` | [`float`](#float) |  |  |
| `snr` | [`float`](#float) |  |  |
| `downlink_path_constraint` | [`DownlinkPathConstraint`](#ttn.lorawan.v3.DownlinkPathConstraint) |  |  |
| `uplink_token` | [`bytes`](#bytes) |  |  |
| `packet_broker` | [`MACState.UplinkMessage.RxMetadata.PacketBrokerMetadata`](#ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata.PacketBrokerMetadata) |  |  |
| `relay` | [`MACState.UplinkMessage.RxMetadata.RelayMetadata`](#ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata.RelayMetadata) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `downlink_path_constraint` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata.PacketBrokerMetadata">Message `MACState.UplinkMessage.RxMetadata.PacketBrokerMetadata`</a>

### <a name="ttn.lorawan.v3.MACState.UplinkMessage.RxMetadata.RelayMetadata">Message `MACState.UplinkMessage.RxMetadata.RelayMetadata`</a>

### <a name="ttn.lorawan.v3.MACState.UplinkMessage.TxSettings">Message `MACState.UplinkMessage.TxSettings`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data_rate` | [`DataRate`](#ttn.lorawan.v3.DataRate) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `data_rate` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.RelayParameters">Message `RelayParameters`</a>

RelayParameters represent the parameters of a relay.
This is used internally by the Network Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `serving` | [`ServingRelayParameters`](#ttn.lorawan.v3.ServingRelayParameters) |  | Parameters related to a relay which is serving end devices. |
| `served` | [`ServedRelayParameters`](#ttn.lorawan.v3.ServedRelayParameters) |  | Parameters related to an end device served by a relay. |

### <a name="ttn.lorawan.v3.RelayUplinkForwardingRule">Message `RelayUplinkForwardingRule`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limits` | [`RelayUplinkForwardLimits`](#ttn.lorawan.v3.RelayUplinkForwardLimits) |  | Bucket configuration for the served end device. If unset, no individual limits will apply to the end device, but the relay global limitations will apply. |
| `last_w_f_cnt` | [`uint32`](#uint32) |  | Last wake on radio frame counter used by the served end device. |
| `device_id` | [`string`](#string) |  | End device identifier of the served end device. |
| `session_key_id` | [`bytes`](#bytes) |  | Session key ID of the session keys used to derive the root relay session key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `device_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ResetAndGetEndDeviceRequest">Message `ResetAndGetEndDeviceRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the end device fields that should be returned. See the API reference for which fields can be returned by the different services. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ServedRelayParameters">Message `ServedRelayParameters`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `always` | [`RelayEndDeviceAlwaysMode`](#ttn.lorawan.v3.RelayEndDeviceAlwaysMode) |  | The end device will always attempt to use the relay mode in order to send uplink messages. |
| `dynamic` | [`RelayEndDeviceDynamicMode`](#ttn.lorawan.v3.RelayEndDeviceDynamicMode) |  | The end device will attempt to use relay mode only after a number of uplink messages have been sent without receiving a valid a downlink message. |
| `end_device_controlled` | [`RelayEndDeviceControlledMode`](#ttn.lorawan.v3.RelayEndDeviceControlledMode) |  | The end device will control when it uses the relay mode. This is the default mode. |
| `backoff` | [`uint32`](#uint32) |  | Number of uplinks to be sent without a wake on radio frame. |
| `second_channel` | [`RelaySecondChannel`](#ttn.lorawan.v3.RelaySecondChannel) |  | Second wake on radio channel configuration. |
| `serving_device_id` | [`string`](#string) |  | End device identifier of the serving end device. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `backoff` | <p>`uint32.lte`: `63`</p> |
| `serving_device_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ServingRelayForwardingLimits">Message `ServingRelayForwardingLimits`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reset_behavior` | [`RelayResetLimitCounter`](#ttn.lorawan.v3.RelayResetLimitCounter) |  | Reset behavior of the buckets upon limit update. |
| `join_requests` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  | Bucket configuration for join requests. If unset, no individual limits will apply to join requests, but the relay overall limitations will apply. |
| `notifications` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  | Bucket configuration for unknown device notifications. If unset, no individual limits will apply to unknown end device notifications, but the relay overall limitations will still apply. |
| `uplink_messages` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  | Bucket configuration for uplink messages across all served end devices. If unset, no individual limits will apply to uplink messages across all served end devices, but the relay overall limitations will still apply. |
| `overall` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  | Bucket configuration for all relay messages. If unset, no overall limits will apply to the relay, but individual limitations will still apply. |

### <a name="ttn.lorawan.v3.ServingRelayParameters">Message `ServingRelayParameters`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `second_channel` | [`RelaySecondChannel`](#ttn.lorawan.v3.RelaySecondChannel) |  | Second wake on radio channel configuration. |
| `default_channel_index` | [`uint32`](#uint32) |  | Index of the default wake on radio channel. |
| `cad_periodicity` | [`RelayCADPeriodicity`](#ttn.lorawan.v3.RelayCADPeriodicity) |  | Channel activity detection periodicity. |
| `uplink_forwarding_rules` | [`RelayUplinkForwardingRule`](#ttn.lorawan.v3.RelayUplinkForwardingRule) | repeated | Configured uplink forwarding rules. |
| `limits` | [`ServingRelayForwardingLimits`](#ttn.lorawan.v3.ServingRelayForwardingLimits) |  | Configured forwarding limits. If unset, the default value from Network Server configuration will be used. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `default_channel_index` | <p>`uint32.lte`: `255`</p> |
| `cad_periodicity` | <p>`enum.defined_only`: `true`</p> |
| `uplink_forwarding_rules` | <p>`repeated.max_items`: `16`</p> |

### <a name="ttn.lorawan.v3.Session">Message `Session`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr` | [`bytes`](#bytes) |  | Device Address, issued by the Network Server or chosen by device manufacturer in case of testing range (beginning with 00-03). Known by Network Server, Application Server and Join Server. Owned by Network Server. |
| `keys` | [`SessionKeys`](#ttn.lorawan.v3.SessionKeys) |  |  |
| `last_f_cnt_up` | [`uint32`](#uint32) |  | Last uplink frame counter value used. Network Server only. Application Server assumes the Network Server checked it. |
| `last_n_f_cnt_down` | [`uint32`](#uint32) |  | Last network downlink frame counter value used. Network Server only. |
| `last_a_f_cnt_down` | [`uint32`](#uint32) |  | Last application downlink frame counter value used. Application Server only. |
| `last_conf_f_cnt_down` | [`uint32`](#uint32) |  | Frame counter of the last confirmed downlink message sent. Network Server only. |
| `started_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time when the session started. Network Server only. |
| `queued_application_downlinks` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) | repeated | Queued Application downlink messages. Stored in Application Server and Network Server. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `keys` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.SetEndDeviceRequest">Message `SetEndDeviceRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device` | [`EndDevice`](#ttn.lorawan.v3.EndDevice) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the end device fields that should be updated. See the API reference for which fields can be set on the different services. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateEndDeviceRequest">Message `UpdateEndDeviceRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device` | [`EndDevice`](#ttn.lorawan.v3.EndDevice) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the end device fields that should be updated. See the API reference for which fields can be set on the different services. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.PowerState">Enum `PowerState`</a>

Power state of the device.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `POWER_UNKNOWN` | 0 |  |
| `POWER_BATTERY` | 1 |  |
| `POWER_EXTERNAL` | 2 |  |

## <a name="ttn/lorawan/v3/end_device_services.proto">File `ttn/lorawan/v3/end_device_services.proto`</a>

### <a name="ttn.lorawan.v3.EndDeviceBatchRegistry">Service `EndDeviceBatchRegistry`</a>

The EndDeviceBatchRegistry service, exposed by the Identity Server, is used to manage
end device registrations in batches.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Get` | [`BatchGetEndDevicesRequest`](#ttn.lorawan.v3.BatchGetEndDevicesRequest) | [`EndDevices`](#ttn.lorawan.v3.EndDevices) | Get a batch of end devices with the given identifiers, selecting the fields specified in the field mask. More or less fields may be returned, depending on the rights of the caller. Devices not found are skipped and no error is returned. |
| `Delete` | [`BatchDeleteEndDevicesRequest`](#ttn.lorawan.v3.BatchDeleteEndDevicesRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a batch of end devices with the given IDs. This operation is atomic; either all devices are deleted or none. Devices not found are skipped and no error is returned. Before calling this RPC, use the corresponding BatchDelete RPCs of NsEndDeviceRegistry, AsEndDeviceRegistry and optionally the JsEndDeviceRegistry to delete the end devices. If the devices were claimed on a Join Server, use the BatchUnclaim RPC of the DeviceClaimingServer. This is NOT done automatically. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Get` | `GET` | `/api/v3/applications/{application_ids.application_id}/devices/batch` |  |
| `Delete` | `DELETE` | `/api/v3/applications/{application_ids.application_id}/devices/batch` |  |

### <a name="ttn.lorawan.v3.EndDeviceRegistry">Service `EndDeviceRegistry`</a>

The EndDeviceRegistry service, exposed by the Identity Server, is used to manage
end device registrations.

After registering an end device, it also needs to be registered in
the NsEndDeviceRegistry that is exposed by the Network Server,
the AsEndDeviceRegistry that is exposed by the Application Server,
and the JsEndDeviceRegistry that is exposed by the Join Server.

Before deleting an end device it first needs to be deleted from the
NsEndDeviceRegistry, the AsEndDeviceRegistry and the JsEndDeviceRegistry.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateEndDeviceRequest`](#ttn.lorawan.v3.CreateEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Create a new end device within an application. After registering an end device, it also needs to be registered in the NsEndDeviceRegistry that is exposed by the Network Server, the AsEndDeviceRegistry that is exposed by the Application Server, and the JsEndDeviceRegistry that is exposed by the Join Server. |
| `Get` | [`GetEndDeviceRequest`](#ttn.lorawan.v3.GetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Get the end device with the given identifiers, selecting the fields specified in the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `GetIdentifiersForEUIs` | [`GetEndDeviceIdentifiersForEUIsRequest`](#ttn.lorawan.v3.GetEndDeviceIdentifiersForEUIsRequest) | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | Get the identifiers of the end device that has the given EUIs registered. |
| `List` | [`ListEndDevicesRequest`](#ttn.lorawan.v3.ListEndDevicesRequest) | [`EndDevices`](#ttn.lorawan.v3.EndDevices) | List end devices in the given application. Similar to Get, this selects the fields given by the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `Update` | [`UpdateEndDeviceRequest`](#ttn.lorawan.v3.UpdateEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Update the end device, changing the fields specified by the field mask to the provided values. |
| `BatchUpdateLastSeen` | [`BatchUpdateEndDeviceLastSeenRequest`](#ttn.lorawan.v3.BatchUpdateEndDeviceLastSeenRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Update the last seen timestamp for a batch of end devices. |
| `Delete` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the end device with the given IDs. Before deleting an end device it first needs to be deleted from the NsEndDeviceRegistry, the AsEndDeviceRegistry and the JsEndDeviceRegistry. In addition, if the device claimed on a Join Server, it also needs to be unclaimed via the DeviceClaimingServer so it can be claimed in the future. This is NOT done automatically. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Create` | `POST` | `/api/v3/applications/{end_device.ids.application_ids.application_id}/devices` | `*` |
| `Get` | `GET` | `/api/v3/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |
| `List` | `GET` | `/api/v3/applications/{application_ids.application_id}/devices` |  |
| `Update` | `PUT` | `/api/v3/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/applications/{application_ids.application_id}/devices/{device_id}` |  |

### <a name="ttn.lorawan.v3.EndDeviceTemplateConverter">Service `EndDeviceTemplateConverter`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListFormats` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`EndDeviceTemplateFormats`](#ttn.lorawan.v3.EndDeviceTemplateFormats) | Returns the configured formats to convert from. |
| `Convert` | [`ConvertEndDeviceTemplateRequest`](#ttn.lorawan.v3.ConvertEndDeviceTemplateRequest) | [`EndDeviceTemplate`](#ttn.lorawan.v3.EndDeviceTemplate) _stream_ | Converts the binary data to a stream of end device templates. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListFormats` | `GET` | `/api/v3/edtc/formats` |  |
| `Convert` | `POST` | `/api/v3/edtc/convert` | `*` |

## <a name="ttn/lorawan/v3/enums.proto">File `ttn/lorawan/v3/enums.proto`</a>

### <a name="ttn.lorawan.v3.ClusterRole">Enum `ClusterRole`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `NONE` | 0 |  |
| `ENTITY_REGISTRY` | 1 |  |
| `ACCESS` | 2 |  |
| `GATEWAY_SERVER` | 3 |  |
| `NETWORK_SERVER` | 4 |  |
| `APPLICATION_SERVER` | 5 |  |
| `JOIN_SERVER` | 6 |  |
| `CRYPTO_SERVER` | 7 |  |
| `DEVICE_TEMPLATE_CONVERTER` | 8 |  |
| `DEVICE_CLAIMING_SERVER` | 9 |  |
| `GATEWAY_CONFIGURATION_SERVER` | 10 |  |
| `QR_CODE_GENERATOR` | 11 |  |
| `PACKET_BROKER_AGENT` | 12 |  |
| `DEVICE_REPOSITORY` | 13 |  |

### <a name="ttn.lorawan.v3.DownlinkPathConstraint">Enum `DownlinkPathConstraint`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DOWNLINK_PATH_CONSTRAINT_NONE` | 0 | Indicates that the gateway can be selected for downlink without constraints by the Network Server. |
| `DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER` | 1 | Indicates that the gateway can be selected for downlink only if no other or better gateway can be selected. |
| `DOWNLINK_PATH_CONSTRAINT_NEVER` | 2 | Indicates that this gateway will never be selected for downlink, even if that results in no available downlink path. |

### <a name="ttn.lorawan.v3.State">Enum `State`</a>

State enum defines states that an entity can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `STATE_REQUESTED` | 0 | Denotes that the entity has been requested and is pending review by an admin. |
| `STATE_APPROVED` | 1 | Denotes that the entity has been reviewed and approved by an admin. |
| `STATE_REJECTED` | 2 | Denotes that the entity has been reviewed and rejected by an admin. |
| `STATE_FLAGGED` | 3 | Denotes that the entity has been flagged and is pending review by an admin. |
| `STATE_SUSPENDED` | 4 | Denotes that the entity has been reviewed and suspended by an admin. |

## <a name="ttn/lorawan/v3/error.proto">File `ttn/lorawan/v3/error.proto`</a>

### <a name="ttn.lorawan.v3.ErrorDetails">Message `ErrorDetails`</a>

Error details that are communicated over gRPC (and HTTP) APIs.
The messages (for translation) are stored as "error:<namespace>:<name>".

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `namespace` | [`string`](#string) |  | Namespace of the error (typically the package name in The Things Stack). |
| `name` | [`string`](#string) |  | Name of the error. |
| `message_format` | [`string`](#string) |  | The default (fallback) message format that should be used for the error. This is also used if the client does not have a translation for the error. |
| `attributes` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Attributes that should be filled into the message format. Any extra attributes can be displayed as error details. |
| `correlation_id` | [`string`](#string) |  | The correlation ID of the error can be used to correlate the error to stack traces the network may (or may not) store about recent errors. |
| `cause` | [`ErrorDetails`](#ttn.lorawan.v3.ErrorDetails) |  | The error that caused this error. |
| `code` | [`uint32`](#uint32) |  | The status code of the error. |
| `details` | [`google.protobuf.Any`](#google.protobuf.Any) | repeated | The details of the error. |

## <a name="ttn/lorawan/v3/events.proto">File `ttn/lorawan/v3/events.proto`</a>

### <a name="ttn.lorawan.v3.Event">Message `Event`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [`string`](#string) |  | Name of the event. This can be used to find the (localized) event description. |
| `time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time at which the event was triggered. |
| `identifiers` | [`EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers) | repeated | Identifiers of the entity (or entities) involved. |
| `data` | [`google.protobuf.Any`](#google.protobuf.Any) |  | Optional data attached to the event. |
| `correlation_ids` | [`string`](#string) | repeated | Correlation IDs can be used to find related events and actions such as API calls. |
| `origin` | [`string`](#string) |  | The origin of the event. Typically the hostname of the server that created it. |
| `context` | [`Event.ContextEntry`](#ttn.lorawan.v3.Event.ContextEntry) | repeated | Event context, internal use only. |
| `visibility` | [`Rights`](#ttn.lorawan.v3.Rights) |  | The event will be visible to a caller that has any of these rights. |
| `authentication` | [`Event.Authentication`](#ttn.lorawan.v3.Event.Authentication) |  | Details on the authentication provided by the caller that triggered this event. |
| `remote_ip` | [`string`](#string) |  | The IP address of the caller that triggered this event. |
| `user_agent` | [`string`](#string) |  | The IP address of the caller that triggered this event. |
| `unique_id` | [`string`](#string) |  | The unique identifier of the event, assigned on creation. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `time` | <p>`timestamp.required`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.Event.Authentication">Message `Event.Authentication`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [`string`](#string) |  | The type of authentication that was used. This is typically a bearer token. |
| `token_type` | [`string`](#string) |  | The type of token that was used. Common types are APIKey, AccessToken and SessionToken. |
| `token_id` | [`string`](#string) |  | The ID of the token that was used. |

### <a name="ttn.lorawan.v3.Event.ContextEntry">Message `Event.ContextEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`bytes`](#bytes) |  |  |

### <a name="ttn.lorawan.v3.FindRelatedEventsRequest">Message `FindRelatedEventsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `correlation_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `correlation_id` | <p>`string.min_len`: `1`</p><p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.FindRelatedEventsResponse">Message `FindRelatedEventsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `events` | [`Event`](#ttn.lorawan.v3.Event) | repeated |  |

### <a name="ttn.lorawan.v3.StreamEventsRequest">Message `StreamEventsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `identifiers` | [`EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers) | repeated |  |
| `tail` | [`uint32`](#uint32) |  | If greater than zero, this will return historical events, up to this maximum when the stream starts. If used in combination with "after", the limit that is reached first, is used. The availability of historical events depends on server support and retention policy. |
| `after` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | If not empty, this will return historical events after the given time when the stream starts. If used in combination with "tail", the limit that is reached first, is used. The availability of historical events depends on server support and retention policy. |
| `names` | [`string`](#string) | repeated | If provided, this will filter events, so that only events with the given names are returned. Names can be provided as either exact event names (e.g. 'gs.up.receive'), or as regular expressions (e.g. '/^gs\..+/'). |

### <a name="ttn.lorawan.v3.Events">Service `Events`</a>

The Events service serves events from the cluster.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Stream` | [`StreamEventsRequest`](#ttn.lorawan.v3.StreamEventsRequest) | [`Event`](#ttn.lorawan.v3.Event) _stream_ | Stream live events, optionally with a tail of historical events (depending on server support and retention policy). Events may arrive out-of-order. |
| `FindRelated` | [`FindRelatedEventsRequest`](#ttn.lorawan.v3.FindRelatedEventsRequest) | [`FindRelatedEventsResponse`](#ttn.lorawan.v3.FindRelatedEventsResponse) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Stream` | `POST` | `/api/v3/events` | `*` |
| `FindRelated` | `GET` | `/api/v3/events/related` |  |

## <a name="ttn/lorawan/v3/gateway.proto">File `ttn/lorawan/v3/gateway.proto`</a>

### <a name="ttn.lorawan.v3.CreateGatewayAPIKeyRequest">Message `CreateGatewayAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `name` | [`string`](#string) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `rights` | <p>`repeated.min_items`: `1`</p><p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `expires_at` | <p>`timestamp.gt_now`: `true`</p> |

### <a name="ttn.lorawan.v3.CreateGatewayRequest">Message `CreateGatewayRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway` | [`Gateway`](#ttn.lorawan.v3.Gateway) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created gateway. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteGatewayAPIKeyRequest">Message `DeleteGatewayAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteGatewayCollaboratorRequest">Message `DeleteGatewayCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `collaborator_ids` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `collaborator_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.Gateway">Message `Gateway`</a>

Gateway is the message that defines a gateway on the network.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  | The identifiers of the gateway. These are public and can be seen by any authenticated user in the network. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the gateway was created. This information is public and can be seen by any authenticated user in the network. |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the gateway was last updated. This information is public and can be seen by any authenticated user in the network. |
| `deleted_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the gateway was deleted. This information is public and can be seen by any authenticated user in the network. |
| `name` | [`string`](#string) |  | The name of the gateway. This information is public and can be seen by any authenticated user in the network. |
| `description` | [`string`](#string) |  | A description for the gateway. This information is public and can be seen by any authenticated user in the network. |
| `attributes` | [`Gateway.AttributesEntry`](#ttn.lorawan.v3.Gateway.AttributesEntry) | repeated | Key-value attributes for this gateway. Typically used for organizing gateways or for storing integration-specific data. |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated | Contact information for this gateway. Typically used to indicate who to contact with technical/security questions about the gateway. This field is deprecated. Use administrative_contact and technical_contact instead. |
| `administrative_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `technical_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `version_ids` | [`GatewayVersionIdentifiers`](#ttn.lorawan.v3.GatewayVersionIdentifiers) |  |  |
| `gateway_server_address` | [`string`](#string) |  | The address of the Gateway Server to connect to. This information is public and can be seen by any authenticated user in the network if status_public is true. The typical format of the address is "scheme://host:port". The scheme is optional. If the port is omitted, the normal port inference (with DNS lookup, otherwise defaults) is used. The connection shall be established with transport layer security (TLS). Custom certificate authorities may be configured out-of-band. |
| `auto_update` | [`bool`](#bool) |  |  |
| `update_channel` | [`string`](#string) |  |  |
| `frequency_plan_id` | [`string`](#string) |  | Frequency plan ID of the gateway. This information is public and can be seen by any authenticated user in the network. DEPRECATED: use frequency_plan_ids. This equals the first element of the frequency_plan_ids field. |
| `frequency_plan_ids` | [`string`](#string) | repeated | Frequency plan IDs of the gateway. This information is public and can be seen by any authenticated user in the network. The first element equals the frequency_plan_id field. |
| `antennas` | [`GatewayAntenna`](#ttn.lorawan.v3.GatewayAntenna) | repeated | Antennas of the gateway. Location information of the antennas is public and can be seen by any authenticated user in the network if location_public=true. |
| `status_public` | [`bool`](#bool) |  | The status of this gateway may be publicly displayed. |
| `location_public` | [`bool`](#bool) |  | The location of this gateway may be publicly displayed. |
| `schedule_downlink_late` | [`bool`](#bool) |  | Enable server-side buffering of downlink messages. This is recommended for gateways using the Semtech UDP Packet Forwarder v2.x or older, as it does not feature a just-in-time queue. If enabled, the Gateway Server schedules the downlink message late to the gateway so that it does not overwrite previously scheduled downlink messages that have not been transmitted yet. |
| `enforce_duty_cycle` | [`bool`](#bool) |  | Enforcing gateway duty cycle is recommended for all gateways to respect spectrum regulations. Disable enforcing the duty cycle only in controlled research and development environments. |
| `downlink_path_constraint` | [`DownlinkPathConstraint`](#ttn.lorawan.v3.DownlinkPathConstraint) |  |  |
| `schedule_anytime_delay` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Adjust the time that GS schedules class C messages in advance. This is useful for gateways that have a known high latency backhaul, like 3G and satellite. |
| `update_location_from_status` | [`bool`](#bool) |  | Update the location of this gateway from status messages. This only works for gateways connecting with authentication; gateways connected over UDP are not supported. |
| `lbs_lns_secret` | [`Secret`](#ttn.lorawan.v3.Secret) |  | The LoRa Basics Station LNS secret. This is either an auth token (such as an API Key) or a TLS private certificate. Requires the RIGHT_GATEWAY_READ_SECRETS for reading and RIGHT_GATEWAY_WRITE_SECRETS for updating this value. |
| `claim_authentication_code` | [`GatewayClaimAuthenticationCode`](#ttn.lorawan.v3.GatewayClaimAuthenticationCode) |  | The authentication code for gateway claiming. Requires the RIGHT_GATEWAY_READ_SECRETS for reading and RIGHT_GATEWAY_WRITE_SECRETS for updating this value. The entire field must be used in RPCs since sub-fields are validated wrt to each other. Direct selection/update of sub-fields only are not allowed. Use the top level field mask `claim_authentication_code` even when updating single fields. |
| `target_cups_uri` | [`string`](#string) |  | CUPS URI for LoRa Basics Station CUPS redirection. The CUPS Trust field will be automatically fetched from the cert chain presented by the target server. |
| `target_cups_key` | [`Secret`](#ttn.lorawan.v3.Secret) |  | CUPS Key for LoRa Basics Station CUPS redirection. If redirecting to another instance of TTS, use the CUPS API Key for the gateway on the target instance. Requires the RIGHT_GATEWAY_READ_SECRETS for reading and RIGHT_GATEWAY_WRITE_SECRETS for updating this value. |
| `require_authenticated_connection` | [`bool`](#bool) |  | Require an authenticated gateway connection. This prevents the gateway from using the UDP protocol and requires authentication when using other protocols. |
| `lrfhss` | [`Gateway.LRFHSS`](#ttn.lorawan.v3.Gateway.LRFHSS) |  |  |
| `disable_packet_broker_forwarding` | [`bool`](#bool) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `description` | <p>`string.max_len`: `2000`</p> |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |
| `contact_info` | <p>`repeated.max_items`: `10`</p> |
| `gateway_server_address` | <p>`string.pattern`: `^([a-z]{2,5}://)?(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `update_channel` | <p>`string.max_len`: `128`</p> |
| `frequency_plan_id` | <p>`string.max_len`: `64`</p> |
| `frequency_plan_ids` | <p>`repeated.max_items`: `8`</p><p>`repeated.items.string.max_len`: `64`</p> |
| `antennas` | <p>`repeated.max_items`: `8`</p> |
| `downlink_path_constraint` | <p>`enum.defined_only`: `true`</p> |
| `target_cups_uri` | <p>`string.uri`: `true`</p> |

### <a name="ttn.lorawan.v3.Gateway.AttributesEntry">Message `Gateway.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.Gateway.LRFHSS">Message `Gateway.LRFHSS`</a>

LR-FHSS gateway capabilities.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `supported` | [`bool`](#bool) |  | The gateway supports the LR-FHSS uplink channels. |

### <a name="ttn.lorawan.v3.GatewayAntenna">Message `GatewayAntenna`</a>

GatewayAntenna is the message that defines a gateway antenna.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gain` | [`float`](#float) |  | Antenna gain relative to the gateway, in dBi. |
| `location` | [`Location`](#ttn.lorawan.v3.Location) |  | location is the antenna's location. |
| `attributes` | [`GatewayAntenna.AttributesEntry`](#ttn.lorawan.v3.GatewayAntenna.AttributesEntry) | repeated |  |
| `placement` | [`GatewayAntennaPlacement`](#ttn.lorawan.v3.GatewayAntennaPlacement) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |

### <a name="ttn.lorawan.v3.GatewayAntenna.AttributesEntry">Message `GatewayAntenna.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.GatewayBrand">Message `GatewayBrand`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  |  |
| `name` | [`string`](#string) |  |  |
| `url` | [`string`](#string) |  |  |
| `logos` | [`string`](#string) | repeated | Logos contains file names of brand logos. |

### <a name="ttn.lorawan.v3.GatewayClaimAuthenticationCode">Message `GatewayClaimAuthenticationCode`</a>

Authentication code for claiming gateways.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `secret` | [`Secret`](#ttn.lorawan.v3.Secret) |  |  |
| `valid_from` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `valid_to` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

### <a name="ttn.lorawan.v3.GatewayConnectionStats">Message `GatewayConnectionStats`</a>

Connection stats as monitored by the Gateway Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `connected_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `disconnected_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `protocol` | [`string`](#string) |  | Protocol used to connect (for example, udp, mqtt, grpc) |
| `last_status_received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `last_status` | [`GatewayStatus`](#ttn.lorawan.v3.GatewayStatus) |  |  |
| `last_uplink_received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `uplink_count` | [`uint64`](#uint64) |  |  |
| `last_downlink_received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `downlink_count` | [`uint64`](#uint64) |  |  |
| `last_tx_acknowledgment_received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `tx_acknowledgment_count` | [`uint64`](#uint64) |  |  |
| `round_trip_times` | [`GatewayConnectionStats.RoundTripTimes`](#ttn.lorawan.v3.GatewayConnectionStats.RoundTripTimes) |  |  |
| `sub_bands` | [`GatewayConnectionStats.SubBand`](#ttn.lorawan.v3.GatewayConnectionStats.SubBand) | repeated | Statistics for each sub band. |
| `gateway_remote_address` | [`GatewayRemoteAddress`](#ttn.lorawan.v3.GatewayRemoteAddress) |  | Gateway Remote Address. |

### <a name="ttn.lorawan.v3.GatewayConnectionStats.RoundTripTimes">Message `GatewayConnectionStats.RoundTripTimes`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `max` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `median` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `count` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `min` | <p>`duration.required`: `true`</p> |
| `max` | <p>`duration.required`: `true`</p> |
| `median` | <p>`duration.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GatewayConnectionStats.SubBand">Message `GatewayConnectionStats.SubBand`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_frequency` | [`uint64`](#uint64) |  |  |
| `max_frequency` | [`uint64`](#uint64) |  |  |
| `downlink_utilization_limit` | [`float`](#float) |  | Duty-cycle limit of the sub-band as a fraction of time. |
| `downlink_utilization` | [`float`](#float) |  | Utilization rate of the available duty-cycle. This value should not exceed downlink_utilization_limit. |

### <a name="ttn.lorawan.v3.GatewayModel">Message `GatewayModel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `brand_id` | [`string`](#string) |  |  |
| `id` | [`string`](#string) |  |  |
| `name` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.GatewayRadio">Message `GatewayRadio`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable` | [`bool`](#bool) |  |  |
| `chip_type` | [`string`](#string) |  |  |
| `frequency` | [`uint64`](#uint64) |  |  |
| `rssi_offset` | [`float`](#float) |  |  |
| `tx_configuration` | [`GatewayRadio.TxConfiguration`](#ttn.lorawan.v3.GatewayRadio.TxConfiguration) |  |  |

### <a name="ttn.lorawan.v3.GatewayRadio.TxConfiguration">Message `GatewayRadio.TxConfiguration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_frequency` | [`uint64`](#uint64) |  |  |
| `max_frequency` | [`uint64`](#uint64) |  |  |
| `notch_frequency` | [`uint64`](#uint64) |  |  |

### <a name="ttn.lorawan.v3.GatewayRemoteAddress">Message `GatewayRemoteAddress`</a>

Remote Address of the Gateway, as seen by the Gateway Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ip` | [`string`](#string) |  | IPv4 or IPv6 address. |

### <a name="ttn.lorawan.v3.GatewayStatus">Message `GatewayStatus`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Current time of the gateway |
| `boot_time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Boot time of the gateway - can be left out to save bandwidth; old value will be kept |
| `versions` | [`GatewayStatus.VersionsEntry`](#ttn.lorawan.v3.GatewayStatus.VersionsEntry) | repeated | Versions of gateway subsystems - each field can be left out to save bandwidth; old value will be kept - map keys are written in snake_case - for example: firmware: "2.0.4" forwarder: "v2-3.3.1" fpga: "48" dsp: "27" hal: "v2-3.5.0" |
| `antenna_locations` | [`Location`](#ttn.lorawan.v3.Location) | repeated | Location of each gateway's antenna - if left out, server uses registry-set location as fallback |
| `ip` | [`string`](#string) | repeated | IP addresses of this gateway. Repeated addresses can be used to communicate addresses of multiple interfaces (LAN, Public IP, ...). |
| `metrics` | [`GatewayStatus.MetricsEntry`](#ttn.lorawan.v3.GatewayStatus.MetricsEntry) | repeated | Metrics - can be used for forwarding gateway metrics such as temperatures or performance metrics - map keys are written in snake_case |
| `advanced` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `versions` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[_-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `128`</p> |
| `antenna_locations` | <p>`repeated.max_items`: `8`</p> |
| `ip` | <p>`repeated.max_items`: `10`</p><p>`repeated.items.string.ip`: `true`</p> |
| `metrics` | <p>`map.max_pairs`: `32`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[_-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.GatewayStatus.MetricsEntry">Message `GatewayStatus.MetricsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`float`](#float) |  |  |

### <a name="ttn.lorawan.v3.GatewayStatus.VersionsEntry">Message `GatewayStatus.VersionsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.GatewayVersionIdentifiers">Message `GatewayVersionIdentifiers`</a>

Identifies an end device model with version information.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `brand_id` | [`string`](#string) |  |  |
| `model_id` | [`string`](#string) |  |  |
| `hardware_version` | [`string`](#string) |  |  |
| `firmware_version` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `model_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `hardware_version` | <p>`string.max_len`: `32`</p> |
| `firmware_version` | <p>`string.max_len`: `32`</p> |

### <a name="ttn.lorawan.v3.Gateways">Message `Gateways`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateways` | [`Gateway`](#ttn.lorawan.v3.Gateway) | repeated |  |

### <a name="ttn.lorawan.v3.GetGatewayAPIKeyRequest">Message `GetGatewayAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  | Unique public identifier for the API key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetGatewayCollaboratorRequest">Message `GetGatewayCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest">Message `GetGatewayIdentifiersForEUIRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetGatewayRequest">Message `GetGatewayRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the gateway fields that should be returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ListGatewayAPIKeysRequest">Message `ListGatewayAPIKeysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path. Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `order` | <p>`string.in`: `[ api_key_id -api_key_id name -name created_at -created_at expires_at -expires_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListGatewayCollaboratorsRequest">Message `ListGatewayCollaboratorsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `order` | <p>`string.in`: `[ id -id -rights rights]`</p> |

### <a name="ttn.lorawan.v3.ListGatewaysRequest">Message `ListGatewaysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | By default we list all gateways the caller has rights on. Set the user or the organization (not both) to instead list the gateways where the user or organization is collaborator on. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the gateway fields that should be returned. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted gateways. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `order` | <p>`string.in`: `[ gateway_id -gateway_id gateway_eui -gateway_eui name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SetGatewayCollaboratorRequest">Message `SetGatewayCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `collaborator` | [`Collaborator`](#ttn.lorawan.v3.Collaborator) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateGatewayAPIKeyRequest">Message `UpdateGatewayAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `api_key` | [`APIKey`](#ttn.lorawan.v3.APIKey) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the api key fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `api_key` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateGatewayRequest">Message `UpdateGatewayRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway` | [`Gateway`](#ttn.lorawan.v3.Gateway) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the gateway fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GatewayAntennaPlacement">Enum `GatewayAntennaPlacement`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `PLACEMENT_UNKNOWN` | 0 |  |
| `INDOOR` | 1 |  |
| `OUTDOOR` | 2 |  |

## <a name="ttn/lorawan/v3/gateway_configuration.proto">File `ttn/lorawan/v3/gateway_configuration.proto`</a>

### <a name="ttn.lorawan.v3.GetGatewayConfigurationRequest">Message `GetGatewayConfigurationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `format` | [`string`](#string) |  |  |
| `type` | [`string`](#string) |  |  |
| `filename` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `format` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |
| `type` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |
| `filename` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-._]?[a-z0-9]){2,}$|^$`</p> |

### <a name="ttn.lorawan.v3.GetGatewayConfigurationResponse">Message `GetGatewayConfigurationResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contents` | [`bytes`](#bytes) |  |  |

### <a name="ttn.lorawan.v3.GatewayConfigurationService">Service `GatewayConfigurationService`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetGatewayConfiguration` | [`GetGatewayConfigurationRequest`](#ttn.lorawan.v3.GetGatewayConfigurationRequest) | [`GetGatewayConfigurationResponse`](#ttn.lorawan.v3.GetGatewayConfigurationResponse) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetGatewayConfiguration` | `` | `/api/v3` |  |
| `GetGatewayConfiguration` | `GET` | `/api/v3/gcs/gateways/configuration/{gateway_ids.gateway_id}/{format}/{filename}` |  |
| `GetGatewayConfiguration` | `GET` | `/api/v3/gcs/gateways/configuration/{gateway_ids.gateway_id}/{format}/{type}/{filename}` |  |

## <a name="ttn/lorawan/v3/gateway_services.proto">File `ttn/lorawan/v3/gateway_services.proto`</a>

### <a name="ttn.lorawan.v3.AssertGatewayRightsRequest">Message `AssertGatewayRightsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | repeated |  |
| `required` | [`Rights`](#ttn.lorawan.v3.Rights) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`repeated.min_items`: `1`</p><p>`repeated.max_items`: `100`</p> |
| `required` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.BatchDeleteGatewaysRequest">Message `BatchDeleteGatewaysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`repeated.min_items`: `1`</p><p>`repeated.max_items`: `20`</p> |

### <a name="ttn.lorawan.v3.PullGatewayConfigurationRequest">Message `PullGatewayConfigurationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

### <a name="ttn.lorawan.v3.GatewayAccess">Service `GatewayAccess`</a>

The GatewayAccess service, exposed by the Identity Server, is used to manage
API keys and collaborators of gateways.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListRights` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`Rights`](#ttn.lorawan.v3.Rights) | List the rights the caller has on this gateway. |
| `CreateAPIKey` | [`CreateGatewayAPIKeyRequest`](#ttn.lorawan.v3.CreateGatewayAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Create an API key scoped to this gateway. |
| `ListAPIKeys` | [`ListGatewayAPIKeysRequest`](#ttn.lorawan.v3.ListGatewayAPIKeysRequest) | [`APIKeys`](#ttn.lorawan.v3.APIKeys) | List the API keys for this gateway. |
| `GetAPIKey` | [`GetGatewayAPIKeyRequest`](#ttn.lorawan.v3.GetGatewayAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Get a single API key of this gateway. |
| `UpdateAPIKey` | [`UpdateGatewayAPIKeyRequest`](#ttn.lorawan.v3.UpdateGatewayAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Update the rights of an API key of the gateway. This method can also be used to delete the API key, by giving it no rights. The caller is required to have all assigned or/and removed rights. |
| `DeleteAPIKey` | [`DeleteGatewayAPIKeyRequest`](#ttn.lorawan.v3.DeleteGatewayAPIKeyRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a single API key of this gateway. |
| `GetCollaborator` | [`GetGatewayCollaboratorRequest`](#ttn.lorawan.v3.GetGatewayCollaboratorRequest) | [`GetCollaboratorResponse`](#ttn.lorawan.v3.GetCollaboratorResponse) | Get the rights of a collaborator (member) of the gateway. Pseudo-rights in the response (such as the "_ALL" right) are not expanded. |
| `SetCollaborator` | [`SetGatewayCollaboratorRequest`](#ttn.lorawan.v3.SetGatewayCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Set the rights of a collaborator (member) on the gateway. This method can also be used to delete the collaborator, by giving them no rights. The caller is required to have all assigned or/and removed rights. |
| `ListCollaborators` | [`ListGatewayCollaboratorsRequest`](#ttn.lorawan.v3.ListGatewayCollaboratorsRequest) | [`Collaborators`](#ttn.lorawan.v3.Collaborators) | List the collaborators on this gateway. |
| `DeleteCollaborator` | [`DeleteGatewayCollaboratorRequest`](#ttn.lorawan.v3.DeleteGatewayCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | DeleteCollaborator removes a collaborator from a gateway. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListRights` | `GET` | `/api/v3/gateways/{gateway_id}/rights` |  |
| `CreateAPIKey` | `POST` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys` | `*` |
| `ListAPIKeys` | `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys` |  |
| `GetAPIKey` | `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys/{key_id}` |  |
| `UpdateAPIKey` | `PUT` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys/{api_key.id}` | `*` |
| `DeleteAPIKey` | `DELETE` | `/api/v3/gateways/{gateway_ids.gateway_id}/api-keys/{key_id}` |  |
| `GetCollaborator` | `` | `/api/v3` |  |
| `GetCollaborator` | `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
| `GetCollaborator` | `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborator/organization/{collaborator.organization_ids.organization_id}` |  |
| `SetCollaborator` | `PUT` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborators` | `*` |
| `ListCollaborators` | `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborators` |  |
| `DeleteCollaborator` | `` | `/api/v3` |  |
| `DeleteCollaborator` | `DELETE` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborators/user/{collaborator_ids.user_ids.user_id}` |  |
| `DeleteCollaborator` | `DELETE` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborators/organization/{collaborator_ids.organization_ids.organization_id}` |  |

### <a name="ttn.lorawan.v3.GatewayBatchAccess">Service `GatewayBatchAccess`</a>

The GatewayBatchAccess service, exposed by the Identity Server, is used to
check the rights of the caller on multiple gateways at once.
EXPERIMENTAL: This service is subject to change.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `AssertRights` | [`AssertGatewayRightsRequest`](#ttn.lorawan.v3.AssertGatewayRightsRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Assert that the caller has the requested rights on all the requested gateways. The check is successful if there are no errors. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `AssertRights` | `GET` | `/api/v3/gateways/rights/batch` |  |

### <a name="ttn.lorawan.v3.GatewayBatchRegistry">Service `GatewayBatchRegistry`</a>

The GatewayBatchRegistry service, exposed by the Identity Server, is used to manage
gateway registrations in batches.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Delete` | [`BatchDeleteGatewaysRequest`](#ttn.lorawan.v3.BatchDeleteGatewaysRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a batch of gateways. This operation is atomic; either all gateways are deleted or none. The caller must have delete rights on all requested gateways. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Delete` | `DELETE` | `/api/v3/gateways/batch` |  |

### <a name="ttn.lorawan.v3.GatewayConfigurator">Service `GatewayConfigurator`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `PullConfiguration` | [`PullGatewayConfigurationRequest`](#ttn.lorawan.v3.PullGatewayConfigurationRequest) | [`Gateway`](#ttn.lorawan.v3.Gateway) _stream_ |  |

### <a name="ttn.lorawan.v3.GatewayRegistry">Service `GatewayRegistry`</a>

The GatewayRegistry service, exposed by the Identity Server, is used to manage
gateway registrations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateGatewayRequest`](#ttn.lorawan.v3.CreateGatewayRequest) | [`Gateway`](#ttn.lorawan.v3.Gateway) | Create a new gateway. This also sets the given organization or user as first collaborator with all possible rights. |
| `Get` | [`GetGatewayRequest`](#ttn.lorawan.v3.GetGatewayRequest) | [`Gateway`](#ttn.lorawan.v3.Gateway) | Get the gateway with the given identifiers, selecting the fields specified in the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `GetIdentifiersForEUI` | [`GetGatewayIdentifiersForEUIRequest`](#ttn.lorawan.v3.GetGatewayIdentifiersForEUIRequest) | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | Get the identifiers of the gateway that has the given EUI registered. |
| `List` | [`ListGatewaysRequest`](#ttn.lorawan.v3.ListGatewaysRequest) | [`Gateways`](#ttn.lorawan.v3.Gateways) | List gateways where the given user or organization is a direct collaborator. If no user or organization is given, this returns the gateways the caller has access to. Similar to Get, this selects the fields given by the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `Update` | [`UpdateGatewayRequest`](#ttn.lorawan.v3.UpdateGatewayRequest) | [`Gateway`](#ttn.lorawan.v3.Gateway) | Update the gateway, changing the fields specified by the field mask to the provided values. |
| `Delete` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the gateway. This may not release the gateway ID for reuse, but it does release the EUI. |
| `Restore` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Restore a recently deleted gateway. This does not restore the EUI, as that was released when deleting the gateway. Deployment configuration may specify if, and for how long after deletion, entities can be restored. |
| `Purge` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Purge the gateway. This will release both gateway ID and EUI for reuse. The gateway owner is responsible for clearing data from any (external) integrations that may store and expose data by gateway ID. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Create` | `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/gateways` | `*` |
| `Create` | `POST` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/gateways` | `*` |
| `Get` | `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}` |  |
| `List` | `GET` | `/api/v3/gateways` |  |
| `List` | `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/gateways` |  |
| `List` | `GET` | `/api/v3/organizations/{collaborator.organization_ids.organization_id}/gateways` |  |
| `Update` | `PUT` | `/api/v3/gateways/{gateway.ids.gateway_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/gateways/{gateway_id}` |  |
| `Restore` | `POST` | `/api/v3/gateways/{gateway_id}/restore` |  |
| `Purge` | `DELETE` | `/api/v3/gateways/{gateway_id}/purge` |  |

## <a name="ttn/lorawan/v3/gatewayserver.proto">File `ttn/lorawan/v3/gatewayserver.proto`</a>

### <a name="ttn.lorawan.v3.BatchGetGatewayConnectionStatsRequest">Message `BatchGetGatewayConnectionStatsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | repeated |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the gateway stats fields that should be returned. This mask will be applied on each entry returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`repeated.min_items`: `1`</p><p>`repeated.max_items`: `100`</p> |

### <a name="ttn.lorawan.v3.BatchGetGatewayConnectionStatsResponse">Message `BatchGetGatewayConnectionStatsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entries` | [`BatchGetGatewayConnectionStatsResponse.EntriesEntry`](#ttn.lorawan.v3.BatchGetGatewayConnectionStatsResponse.EntriesEntry) | repeated | The map key is the gateway identifier. |

### <a name="ttn.lorawan.v3.BatchGetGatewayConnectionStatsResponse.EntriesEntry">Message `BatchGetGatewayConnectionStatsResponse.EntriesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`GatewayConnectionStats`](#ttn.lorawan.v3.GatewayConnectionStats) |  |  |

### <a name="ttn.lorawan.v3.GatewayDown">Message `GatewayDown`</a>

GatewayDown contains downlink messages for the gateway.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `downlink_message` | [`DownlinkMessage`](#ttn.lorawan.v3.DownlinkMessage) |  | DownlinkMessage for the gateway. |

### <a name="ttn.lorawan.v3.GatewayUp">Message `GatewayUp`</a>

GatewayUp may contain zero or more uplink messages and/or a status message for the gateway.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uplink_messages` | [`UplinkMessage`](#ttn.lorawan.v3.UplinkMessage) | repeated | Uplink messages received by the gateway. |
| `gateway_status` | [`GatewayStatus`](#ttn.lorawan.v3.GatewayStatus) |  | Gateway status produced by the gateway. |
| `tx_acknowledgment` | [`TxAcknowledgment`](#ttn.lorawan.v3.TxAcknowledgment) |  | A Tx acknowledgment or error. |

### <a name="ttn.lorawan.v3.ScheduleDownlinkErrorDetails">Message `ScheduleDownlinkErrorDetails`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `path_errors` | [`ErrorDetails`](#ttn.lorawan.v3.ErrorDetails) | repeated | Errors per path when downlink scheduling failed. |

### <a name="ttn.lorawan.v3.ScheduleDownlinkResponse">Message `ScheduleDownlinkResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delay` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | The amount of time between the message has been scheduled and it will be transmitted by the gateway. |
| `downlink_path` | [`DownlinkPath`](#ttn.lorawan.v3.DownlinkPath) |  | Downlink path chosen by the Gateway Server. |
| `rx1` | [`bool`](#bool) |  | Whether RX1 has been chosen for the downlink message. Both RX1 and RX2 can be used for transmitting the same message by the same gateway. |
| `rx2` | [`bool`](#bool) |  | Whether RX2 has been chosen for the downlink message. Both RX1 and RX2 can be used for transmitting the same message by the same gateway. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `delay` | <p>`duration.required`: `true`</p> |

### <a name="ttn.lorawan.v3.Gs">Service `Gs`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetGatewayConnectionStats` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`GatewayConnectionStats`](#ttn.lorawan.v3.GatewayConnectionStats) | Get statistics about the current gateway connection to the Gateway Server. This is not persisted between reconnects. |
| `BatchGetGatewayConnectionStats` | [`BatchGetGatewayConnectionStatsRequest`](#ttn.lorawan.v3.BatchGetGatewayConnectionStatsRequest) | [`BatchGetGatewayConnectionStatsResponse`](#ttn.lorawan.v3.BatchGetGatewayConnectionStatsResponse) | Get statistics about gateway connections to the Gateway Server of a batch of gateways. - Statistics are not persisted between reconnects. - Gateways that are not connected or are part of a different cluster are ignored. - The client should ensure that the requested gateways are in the requested cluster. - The client should have the right to get the gateway connection stats on all requested gateways. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetGatewayConnectionStats` | `GET` | `/api/v3/gs/gateways/{gateway_id}/connection/stats` |  |
| `BatchGetGatewayConnectionStats` | `POST` | `/api/v3/gs/gateways/connection/stats` | `*` |

### <a name="ttn.lorawan.v3.GtwGs">Service `GtwGs`</a>

The GtwGs service connects a gateway to a Gateway Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `LinkGateway` | [`GatewayUp`](#ttn.lorawan.v3.GatewayUp) _stream_ | [`GatewayDown`](#ttn.lorawan.v3.GatewayDown) _stream_ | Link a gateway to the Gateway Server for streaming upstream messages and downstream messages. |
| `GetConcentratorConfig` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`ConcentratorConfig`](#ttn.lorawan.v3.ConcentratorConfig) | Get configuration for the concentrator. |
| `GetMQTTConnectionInfo` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`MQTTConnectionInfo`](#ttn.lorawan.v3.MQTTConnectionInfo) | Get connection information to connect an MQTT gateway. |
| `GetMQTTV2ConnectionInfo` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | [`MQTTConnectionInfo`](#ttn.lorawan.v3.MQTTConnectionInfo) | Get legacy connection information to connect a The Things Network Stack V2 MQTT gateway. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetMQTTConnectionInfo` | `GET` | `/api/v3/gs/gateways/{gateway_id}/mqtt-connection-info` |  |
| `GetMQTTV2ConnectionInfo` | `GET` | `/api/v3/gs/gateways/{gateway_id}/mqttv2-connection-info` |  |

### <a name="ttn.lorawan.v3.NsGs">Service `NsGs`</a>

The NsGs service connects a Network Server to a Gateway Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ScheduleDownlink` | [`DownlinkMessage`](#ttn.lorawan.v3.DownlinkMessage) | [`ScheduleDownlinkResponse`](#ttn.lorawan.v3.ScheduleDownlinkResponse) | Instructs the Gateway Server to schedule a downlink message. The Gateway Server may refuse if there are any conflicts in the schedule or if a duty cycle prevents the gateway from transmitting. |

## <a name="ttn/lorawan/v3/identifiers.proto">File `ttn/lorawan/v3/identifiers.proto`</a>

### <a name="ttn.lorawan.v3.ApplicationIdentifiers">Message `ApplicationIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ClientIdentifiers">Message `ClientIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `client_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `client_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.EndDeviceIdentifiers">Message `EndDeviceIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `device_id` | [`string`](#string) |  |  |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `dev_eui` | [`bytes`](#bytes) |  | The LoRaWAN DevEUI. |
| `join_eui` | [`bytes`](#bytes) |  | The LoRaWAN JoinEUI (AppEUI until LoRaWAN 1.0.3 end devices). |
| `dev_addr` | [`bytes`](#bytes) |  | The LoRaWAN DevAddr. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `device_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `application_ids` | <p>`message.required`: `true`</p> |
| `dev_eui` | <p>`bytes.len`: `8`</p> |
| `join_eui` | <p>`bytes.len`: `8`</p> |
| `dev_addr` | <p>`bytes.len`: `4`</p> |

### <a name="ttn.lorawan.v3.EndDeviceIdentifiersList">Message `EndDeviceIdentifiersList`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | repeated |  |

### <a name="ttn.lorawan.v3.EndDeviceVersionIdentifiers">Message `EndDeviceVersionIdentifiers`</a>

Identifies an end device model with version information.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `brand_id` | [`string`](#string) |  |  |
| `model_id` | [`string`](#string) |  |  |
| `hardware_version` | [`string`](#string) |  |  |
| `firmware_version` | [`string`](#string) |  |  |
| `band_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `brand_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `model_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `hardware_version` | <p>`string.max_len`: `32`</p> |
| `firmware_version` | <p>`string.max_len`: `32`</p> |
| `band_id` | <p>`string.max_len`: `32`</p> |

### <a name="ttn.lorawan.v3.EntityIdentifiers">Message `EntityIdentifiers`</a>

EntityIdentifiers contains one of the possible entity identifiers.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |

### <a name="ttn.lorawan.v3.GatewayIdentifiers">Message `GatewayIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_id` | [`string`](#string) |  |  |
| `eui` | [`bytes`](#bytes) |  | Secondary identifier, which can only be used in specific requests. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GatewayIdentifiersList">Message `GatewayIdentifiersList`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) | repeated |  |

### <a name="ttn.lorawan.v3.LoRaAllianceProfileIdentifiers">Message `LoRaAllianceProfileIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vendor_id` | [`uint32`](#uint32) |  | VendorID managed by the LoRa Alliance, as defined in TR005. |
| `vendor_profile_id` | [`uint32`](#uint32) |  | ID of the LoRaWAN end device profile assigned by the vendor. |

### <a name="ttn.lorawan.v3.NetworkIdentifiers">Message `NetworkIdentifiers`</a>

Identifies a Network Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `net_id` | [`bytes`](#bytes) |  | LoRa Alliance NetID. |
| `ns_id` | [`bytes`](#bytes) |  | LoRaWAN NSID (EUI-64) that uniquely identifies the Network Server instance. |
| `tenant_id` | [`string`](#string) |  | Optional tenant identifier for multi-tenant deployments. |
| `cluster_id` | [`string`](#string) |  | Cluster identifier of the Network Server. |
| `cluster_address` | [`string`](#string) |  | Cluster address of the Network Server. |
| `tenant_address` | [`string`](#string) |  | Optional tenant address for multi-tenant deployments. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `net_id` | <p>`bytes.len`: `3`</p> |
| `ns_id` | <p>`bytes.len`: `8`</p> |
| `tenant_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |
| `cluster_id` | <p>`string.max_len`: `64`</p> |
| `cluster_address` | <p>`string.max_len`: `256`</p> |
| `tenant_address` | <p>`string.max_len`: `256`</p> |

### <a name="ttn.lorawan.v3.OrganizationIdentifiers">Message `OrganizationIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_id` | [`string`](#string) |  | This ID shares namespace with user IDs. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.OrganizationOrUserIdentifiers">Message `OrganizationOrUserIdentifiers`</a>

OrganizationOrUserIdentifiers contains either organization or user identifiers.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |

### <a name="ttn.lorawan.v3.UserIdentifiers">Message `UserIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_id` | [`string`](#string) |  | This ID shares namespace with organization IDs. |
| `email` | [`string`](#string) |  | Secondary identifier, which can only be used in specific requests. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){1,}$`</p> |

## <a name="ttn/lorawan/v3/identityserver.proto">File `ttn/lorawan/v3/identityserver.proto`</a>

### <a name="ttn.lorawan.v3.AuthInfoResponse">Message `AuthInfoResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `api_key` | [`AuthInfoResponse.APIKeyAccess`](#ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess) |  |  |
| `oauth_access_token` | [`OAuthAccessToken`](#ttn.lorawan.v3.OAuthAccessToken) |  |  |
| `user_session` | [`UserSession`](#ttn.lorawan.v3.UserSession) |  | Warning: A user authorized by session cookie will be granted all current and future rights. When using this auth type, the respective handlers need to ensure thorough CSRF and CORS protection using appropriate middleware. |
| `gateway_token` | [`AuthInfoResponse.GatewayToken`](#ttn.lorawan.v3.AuthInfoResponse.GatewayToken) |  |  |
| `universal_rights` | [`Rights`](#ttn.lorawan.v3.Rights) |  |  |
| `is_admin` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.AuthInfoResponse.APIKeyAccess">Message `AuthInfoResponse.APIKeyAccess`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `api_key` | [`APIKey`](#ttn.lorawan.v3.APIKey) |  |  |
| `entity_ids` | [`EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `api_key` | <p>`message.required`: `true`</p> |
| `entity_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.AuthInfoResponse.GatewayToken">Message `AuthInfoResponse.GatewayToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetIsConfigurationRequest">Message `GetIsConfigurationRequest`</a>

### <a name="ttn.lorawan.v3.GetIsConfigurationResponse">Message `GetIsConfigurationResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `configuration` | [`IsConfiguration`](#ttn.lorawan.v3.IsConfiguration) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration">Message `IsConfiguration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_registration` | [`IsConfiguration.UserRegistration`](#ttn.lorawan.v3.IsConfiguration.UserRegistration) |  |  |
| `profile_picture` | [`IsConfiguration.ProfilePicture`](#ttn.lorawan.v3.IsConfiguration.ProfilePicture) |  |  |
| `end_device_picture` | [`IsConfiguration.EndDevicePicture`](#ttn.lorawan.v3.IsConfiguration.EndDevicePicture) |  |  |
| `user_rights` | [`IsConfiguration.UserRights`](#ttn.lorawan.v3.IsConfiguration.UserRights) |  |  |
| `user_login` | [`IsConfiguration.UserLogin`](#ttn.lorawan.v3.IsConfiguration.UserLogin) |  |  |
| `admin_rights` | [`IsConfiguration.AdminRights`](#ttn.lorawan.v3.IsConfiguration.AdminRights) |  |  |
| `collaborator_rights` | [`IsConfiguration.CollaboratorRights`](#ttn.lorawan.v3.IsConfiguration.CollaboratorRights) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.AdminRights">Message `IsConfiguration.AdminRights`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `all` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.CollaboratorRights">Message `IsConfiguration.CollaboratorRights`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `set_others_as_contacts` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.EndDevicePicture">Message `IsConfiguration.EndDevicePicture`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `disable_upload` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.ProfilePicture">Message `IsConfiguration.ProfilePicture`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `disable_upload` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `use_gravatar` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.UserLogin">Message `IsConfiguration.UserLogin`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `disable_credentials_login` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.UserRegistration">Message `IsConfiguration.UserRegistration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `invitation` | [`IsConfiguration.UserRegistration.Invitation`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.Invitation) |  |  |
| `contact_info_validation` | [`IsConfiguration.UserRegistration.ContactInfoValidation`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.ContactInfoValidation) |  |  |
| `admin_approval` | [`IsConfiguration.UserRegistration.AdminApproval`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.AdminApproval) |  |  |
| `password_requirements` | [`IsConfiguration.UserRegistration.PasswordRequirements`](#ttn.lorawan.v3.IsConfiguration.UserRegistration.PasswordRequirements) |  |  |
| `enabled` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.UserRegistration.AdminApproval">Message `IsConfiguration.UserRegistration.AdminApproval`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `required` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.UserRegistration.ContactInfoValidation">Message `IsConfiguration.UserRegistration.ContactInfoValidation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `required` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `token_ttl` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `retry_interval` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | The minimum interval between validation emails. |

### <a name="ttn.lorawan.v3.IsConfiguration.UserRegistration.Invitation">Message `IsConfiguration.UserRegistration.Invitation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `required` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `token_ttl` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.UserRegistration.PasswordRequirements">Message `IsConfiguration.UserRegistration.PasswordRequirements`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_length` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `max_length` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `min_uppercase` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `min_digits` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |
| `min_special` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  |  |

### <a name="ttn.lorawan.v3.IsConfiguration.UserRights">Message `IsConfiguration.UserRights`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `create_applications` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `create_clients` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `create_gateways` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |
| `create_organizations` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  |  |

### <a name="ttn.lorawan.v3.EntityAccess">Service `EntityAccess`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `AuthInfo` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`AuthInfoResponse`](#ttn.lorawan.v3.AuthInfoResponse) | AuthInfo returns information about the authentication that is used on the request. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `AuthInfo` | `GET` | `/api/v3/auth_info` |  |

### <a name="ttn.lorawan.v3.Is">Service `Is`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetConfiguration` | [`GetIsConfigurationRequest`](#ttn.lorawan.v3.GetIsConfigurationRequest) | [`GetIsConfigurationResponse`](#ttn.lorawan.v3.GetIsConfigurationResponse) | Get the configuration of the Identity Server. The response is typically used to enable or disable features in a user interface. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetConfiguration` | `GET` | `/api/v3/is/configuration` |  |

## <a name="ttn/lorawan/v3/join.proto">File `ttn/lorawan/v3/join.proto`</a>

### <a name="ttn.lorawan.v3.JoinRequest">Message `JoinRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_payload` | [`bytes`](#bytes) |  |  |
| `payload` | [`Message`](#ttn.lorawan.v3.Message) |  |  |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `selected_mac_version` | [`MACVersion`](#ttn.lorawan.v3.MACVersion) |  |  |
| `net_id` | [`bytes`](#bytes) |  |  |
| `downlink_settings` | [`DLSettings`](#ttn.lorawan.v3.DLSettings) |  |  |
| `rx_delay` | [`RxDelay`](#ttn.lorawan.v3.RxDelay) |  |  |
| `cf_list` | [`CFList`](#ttn.lorawan.v3.CFList) |  | Optional CFList. |
| `correlation_ids` | [`string`](#string) | repeated |  |
| `consumed_airtime` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Consumed airtime for the transmission of the join request. Calculated by Network Server using the RawPayload size and the transmission settings. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `raw_payload` | <p>`bytes.len`: `23`</p> |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `net_id` | <p>`bytes.len`: `3`</p> |
| `downlink_settings` | <p>`message.required`: `true`</p> |
| `rx_delay` | <p>`enum.defined_only`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.JoinResponse">Message `JoinResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_payload` | [`bytes`](#bytes) |  |  |
| `session_keys` | [`SessionKeys`](#ttn.lorawan.v3.SessionKeys) |  |  |
| `lifetime` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |
| `correlation_ids` | [`string`](#string) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `raw_payload` | <p>`bytes.min_len`: `17`</p><p>`bytes.max_len`: `33`</p> |
| `session_keys` | <p>`message.required`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |

## <a name="ttn/lorawan/v3/joinserver.proto">File `ttn/lorawan/v3/joinserver.proto`</a>

### <a name="ttn.lorawan.v3.AppSKeyResponse">Message `AppSKeyResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `app_s_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Session Key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `app_s_key` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationActivationSettings">Message `ApplicationActivationSettings`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `kek_label` | [`string`](#string) |  | The KEK label to use for wrapping application keys. |
| `kek` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Key Encryption Key. |
| `home_net_id` | [`bytes`](#bytes) |  | Home NetID. |
| `application_server_id` | [`string`](#string) |  | The AS-ID of the Application Server to use. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `kek_label` | <p>`string.max_len`: `2048`</p> |
| `home_net_id` | <p>`bytes.len`: `3`</p> |
| `application_server_id` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.CryptoServicePayloadRequest">Message `CryptoServicePayloadRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  | End device identifiers for the cryptographic operation. |
| `lorawan_version` | [`MACVersion`](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN version to use for the cryptographic operation. |
| `payload` | [`bytes`](#bytes) |  | Raw input payload. |
| `provisioner_id` | [`string`](#string) |  | Provisioner that provisioned the end device. |
| `provisioning_data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Provisioning data for the provisioner. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `lorawan_version` | <p>`enum.defined_only`: `true`</p> |
| `payload` | <p>`bytes.max_len`: `256`</p> |
| `provisioner_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |

### <a name="ttn.lorawan.v3.CryptoServicePayloadResponse">Message `CryptoServicePayloadResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payload` | [`bytes`](#bytes) |  | Raw output payload. |

### <a name="ttn.lorawan.v3.DeleteApplicationActivationSettingsRequest">Message `DeleteApplicationActivationSettingsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeriveSessionKeysRequest">Message `DeriveSessionKeysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  | End device identifiers to use for key derivation. The DevAddr must be set in this request. The DevEUI may need to be set, depending on the provisioner. |
| `lorawan_version` | [`MACVersion`](#ttn.lorawan.v3.MACVersion) |  | LoRaWAN key derivation scheme. |
| `join_nonce` | [`bytes`](#bytes) |  | LoRaWAN JoinNonce (or AppNonce). |
| `dev_nonce` | [`bytes`](#bytes) |  | LoRaWAN DevNonce. |
| `net_id` | [`bytes`](#bytes) |  | LoRaWAN NetID. |
| `provisioner_id` | [`string`](#string) |  | Provisioner that provisioned the end device. |
| `provisioning_data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Provisioning data for the provisioner. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `lorawan_version` | <p>`enum.defined_only`: `true`</p> |
| `join_nonce` | <p>`bytes.len`: `3`</p> |
| `dev_nonce` | <p>`bytes.len`: `2`</p> |
| `net_id` | <p>`bytes.len`: `3`</p> |
| `provisioner_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |

### <a name="ttn.lorawan.v3.GetApplicationActivationSettingsRequest">Message `GetApplicationActivationSettingsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetDefaultJoinEUIResponse">Message `GetDefaultJoinEUIResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.GetRootKeysRequest">Message `GetRootKeysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  | End device identifiers to request the root keys for. |
| `provisioner_id` | [`string`](#string) |  | Provisioner that provisioned the end device. |
| `provisioning_data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Provisioning data for the provisioner. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `provisioner_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |

### <a name="ttn.lorawan.v3.JoinAcceptMICRequest">Message `JoinAcceptMICRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payload_request` | [`CryptoServicePayloadRequest`](#ttn.lorawan.v3.CryptoServicePayloadRequest) |  | Request data for the cryptographic operation. |
| `join_request_type` | [`JoinRequestType`](#ttn.lorawan.v3.JoinRequestType) |  | LoRaWAN join-request type. |
| `dev_nonce` | [`bytes`](#bytes) |  | LoRaWAN DevNonce. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `payload_request` | <p>`message.required`: `true`</p> |
| `join_request_type` | <p>`enum.defined_only`: `true`</p> |
| `dev_nonce` | <p>`bytes.len`: `2`</p> |

### <a name="ttn.lorawan.v3.JoinEUIPrefix">Message `JoinEUIPrefix`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `length` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.JoinEUIPrefixes">Message `JoinEUIPrefixes`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `prefixes` | [`JoinEUIPrefix`](#ttn.lorawan.v3.JoinEUIPrefix) | repeated |  |

### <a name="ttn.lorawan.v3.NwkSKeysResponse">Message `NwkSKeysResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `f_nwk_s_int_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Forwarding Network Session Integrity Key (or Network Session Key in 1.0 compatibility mode). |
| `s_nwk_s_int_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Serving Network Session Integrity Key. |
| `nwk_s_enc_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Session Encryption Key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `f_nwk_s_int_key` | <p>`message.required`: `true`</p> |
| `s_nwk_s_int_key` | <p>`message.required`: `true`</p> |
| `nwk_s_enc_key` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest">Message `ProvisionEndDevicesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `provisioner_id` | [`string`](#string) |  | ID of the provisioner service as configured in the Join Server. |
| `provisioning_data` | [`bytes`](#bytes) |  | Vendor-specific provisioning data. |
| `list` | [`ProvisionEndDevicesRequest.IdentifiersList`](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList) |  | List of device identifiers that will be provisioned. The device identifiers must contain device_id and dev_eui. If set, the application_ids must equal the provision request's application_ids. The number of entries in data must match the number of given identifiers. |
| `range` | [`ProvisionEndDevicesRequest.IdentifiersRange`](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange) |  | Provision devices in a range. The device_id will be generated by the provisioner from the vendor-specific data. The dev_eui will be issued from the given start_dev_eui. |
| `from_data` | [`ProvisionEndDevicesRequest.IdentifiersFromData`](#ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData) |  | Provision devices with identifiers from the given data. The device_id and dev_eui will be generated by the provisioner from the vendor-specific data. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `provisioner_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersFromData">Message `ProvisionEndDevicesRequest.IdentifiersFromData`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersList">Message `ProvisionEndDevicesRequest.IdentifiersList`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.ProvisionEndDevicesRequest.IdentifiersRange">Message `ProvisionEndDevicesRequest.IdentifiersRange`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `start_dev_eui` | [`bytes`](#bytes) |  | DevEUI to start issuing from. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |
| `start_dev_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.SessionKeyRequest">Message `SessionKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_key_id` | [`bytes`](#bytes) |  | Join Server issued identifier for the session keys. |
| `dev_eui` | [`bytes`](#bytes) |  | LoRaWAN DevEUI. |
| `join_eui` | [`bytes`](#bytes) |  | The LoRaWAN JoinEUI (AppEUI until LoRaWAN 1.0.3 end devices). |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |
| `dev_eui` | <p>`bytes.len`: `8`</p> |
| `join_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.SetApplicationActivationSettingsRequest">Message `SetApplicationActivationSettingsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `settings` | [`ApplicationActivationSettings`](#ttn.lorawan.v3.ApplicationActivationSettings) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `settings` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.AppJs">Service `AppJs`</a>

The AppJs service connects an Application to a Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetAppSKey` | [`SessionKeyRequest`](#ttn.lorawan.v3.SessionKeyRequest) | [`AppSKeyResponse`](#ttn.lorawan.v3.AppSKeyResponse) | Request the application session key for a particular session. |

### <a name="ttn.lorawan.v3.ApplicationActivationSettingRegistry">Service `ApplicationActivationSettingRegistry`</a>

The ApplicationActivationSettingRegistry service allows clients to manage their application activation settings.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Get` | [`GetApplicationActivationSettingsRequest`](#ttn.lorawan.v3.GetApplicationActivationSettingsRequest) | [`ApplicationActivationSettings`](#ttn.lorawan.v3.ApplicationActivationSettings) | Get returns application activation settings. |
| `Set` | [`SetApplicationActivationSettingsRequest`](#ttn.lorawan.v3.SetApplicationActivationSettingsRequest) | [`ApplicationActivationSettings`](#ttn.lorawan.v3.ApplicationActivationSettings) | Set creates or updates application activation settings. |
| `Delete` | [`DeleteApplicationActivationSettingsRequest`](#ttn.lorawan.v3.DeleteApplicationActivationSettingsRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete deletes application activation settings. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Get` | `GET` | `/api/v3/js/applications/{application_ids.application_id}/settings` |  |
| `Set` | `POST` | `/api/v3/js/applications/{application_ids.application_id}/settings` | `*` |
| `Delete` | `DELETE` | `/api/v3/js/applications/{application_ids.application_id}/settings` |  |

### <a name="ttn.lorawan.v3.ApplicationCryptoService">Service `ApplicationCryptoService`</a>

Service for application layer cryptographic operations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `DeriveAppSKey` | [`DeriveSessionKeysRequest`](#ttn.lorawan.v3.DeriveSessionKeysRequest) | [`AppSKeyResponse`](#ttn.lorawan.v3.AppSKeyResponse) | Derive the application session key (AppSKey). |
| `GetAppKey` | [`GetRootKeysRequest`](#ttn.lorawan.v3.GetRootKeysRequest) | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) | Get the AppKey. Crypto Servers may return status code FAILED_PRECONDITION when root keys are not exposed. |

### <a name="ttn.lorawan.v3.AsJs">Service `AsJs`</a>

The AsJs service connects an Application Server to a Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetAppSKey` | [`SessionKeyRequest`](#ttn.lorawan.v3.SessionKeyRequest) | [`AppSKeyResponse`](#ttn.lorawan.v3.AppSKeyResponse) | Request the application session key for a particular session. |

### <a name="ttn.lorawan.v3.Js">Service `Js`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetJoinEUIPrefixes` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`JoinEUIPrefixes`](#ttn.lorawan.v3.JoinEUIPrefixes) | Request the JoinEUI prefixes that are configured for this Join Server. |
| `GetDefaultJoinEUI` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`GetDefaultJoinEUIResponse`](#ttn.lorawan.v3.GetDefaultJoinEUIResponse) | Request the default JoinEUI that is configured for this Join Server. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetJoinEUIPrefixes` | `GET` | `/api/v3/js/join_eui_prefixes` |  |
| `GetDefaultJoinEUI` | `GET` | `/api/v3/js/default_join_eui` |  |

### <a name="ttn.lorawan.v3.JsEndDeviceBatchRegistry">Service `JsEndDeviceBatchRegistry`</a>

JsEndDeviceBatchRegistry service allows clients to manage batches of end devices on the Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Delete` | [`BatchDeleteEndDevicesRequest`](#ttn.lorawan.v3.BatchDeleteEndDevicesRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a list of devices within the same application. This operation is atomic; either all devices are deleted or none. Devices not found are skipped and no error is returned. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Delete` | `DELETE` | `/api/v3/js/applications/{application_ids.application_id}/devices/batch` |  |

### <a name="ttn.lorawan.v3.JsEndDeviceRegistry">Service `JsEndDeviceRegistry`</a>

The JsEndDeviceRegistry service allows clients to manage their end devices on the Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Get` | [`GetEndDeviceRequest`](#ttn.lorawan.v3.GetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Get returns the device that matches the given identifiers. If there are multiple matches, an error will be returned. |
| `Set` | [`SetEndDeviceRequest`](#ttn.lorawan.v3.SetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Set creates or updates the device. |
| `Provision` | [`ProvisionEndDevicesRequest`](#ttn.lorawan.v3.ProvisionEndDevicesRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) _stream_ | This rpc is deprecated; use EndDeviceTemplateConverter service instead. TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/999) |
| `Delete` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete deletes the device that matches the given identifiers. If there are multiple matches, an error will be returned. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Get` | `GET` | `/api/v3/js/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |
| `Set` | `PUT` | `/api/v3/js/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | `*` |
| `Set` | `POST` | `/api/v3/js/applications/{end_device.ids.application_ids.application_id}/devices` | `*` |
| `Provision` | `PUT` | `/api/v3/js/applications/{application_ids.application_id}/provision-devices` | `*` |
| `Delete` | `DELETE` | `/api/v3/js/applications/{application_ids.application_id}/devices/{device_id}` |  |

### <a name="ttn.lorawan.v3.NetworkCryptoService">Service `NetworkCryptoService`</a>

Service for network layer cryptographic operations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `JoinRequestMIC` | [`CryptoServicePayloadRequest`](#ttn.lorawan.v3.CryptoServicePayloadRequest) | [`CryptoServicePayloadResponse`](#ttn.lorawan.v3.CryptoServicePayloadResponse) | Calculate the join-request message MIC. |
| `JoinAcceptMIC` | [`JoinAcceptMICRequest`](#ttn.lorawan.v3.JoinAcceptMICRequest) | [`CryptoServicePayloadResponse`](#ttn.lorawan.v3.CryptoServicePayloadResponse) | Calculate the join-accept message MIC. |
| `EncryptJoinAccept` | [`CryptoServicePayloadRequest`](#ttn.lorawan.v3.CryptoServicePayloadRequest) | [`CryptoServicePayloadResponse`](#ttn.lorawan.v3.CryptoServicePayloadResponse) | Encrypt the join-accept payload. |
| `EncryptRejoinAccept` | [`CryptoServicePayloadRequest`](#ttn.lorawan.v3.CryptoServicePayloadRequest) | [`CryptoServicePayloadResponse`](#ttn.lorawan.v3.CryptoServicePayloadResponse) | Encrypt the rejoin-accept payload. |
| `DeriveNwkSKeys` | [`DeriveSessionKeysRequest`](#ttn.lorawan.v3.DeriveSessionKeysRequest) | [`NwkSKeysResponse`](#ttn.lorawan.v3.NwkSKeysResponse) | Derive network session keys (NwkSKey, or FNwkSKey, SNwkSKey and NwkSEncKey) |
| `GetNwkKey` | [`GetRootKeysRequest`](#ttn.lorawan.v3.GetRootKeysRequest) | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) | Get the NwkKey. Crypto Servers may return status code FAILED_PRECONDITION when root keys are not exposed. |

### <a name="ttn.lorawan.v3.NsJs">Service `NsJs`</a>

The NsJs service connects a Network Server to a Join Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `HandleJoin` | [`JoinRequest`](#ttn.lorawan.v3.JoinRequest) | [`JoinResponse`](#ttn.lorawan.v3.JoinResponse) | Handle a join-request message. |
| `GetNwkSKeys` | [`SessionKeyRequest`](#ttn.lorawan.v3.SessionKeyRequest) | [`NwkSKeysResponse`](#ttn.lorawan.v3.NwkSKeysResponse) | Request the network session keys for a particular session. |

## <a name="ttn/lorawan/v3/keys.proto">File `ttn/lorawan/v3/keys.proto`</a>

### <a name="ttn.lorawan.v3.KeyEnvelope">Message `KeyEnvelope`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`bytes`](#bytes) |  | The unencrypted AES key. |
| `kek_label` | [`string`](#string) |  | The label of the RFC 3394 key-encryption-key (KEK) that was used to encrypt the key. |
| `encrypted_key` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `key` | <p>`bytes.len`: `16`</p> |
| `kek_label` | <p>`string.max_len`: `2048`</p> |
| `encrypted_key` | <p>`bytes.max_len`: `1024`</p> |

### <a name="ttn.lorawan.v3.RootKeys">Message `RootKeys`</a>

Root keys for a LoRaWAN device.
These are stored on the Join Server.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `root_key_id` | [`string`](#string) |  | Join Server issued identifier for the root keys. |
| `app_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Key. |
| `nwk_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `root_key_id` | <p>`string.max_len`: `2048`</p> |

### <a name="ttn.lorawan.v3.SessionKeys">Message `SessionKeys`</a>

Session keys for a LoRaWAN session.
Only the components for which the keys were meant, will have the key-encryption-key (KEK) to decrypt the individual keys.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_key_id` | [`bytes`](#bytes) |  | Join Server issued identifier for the session keys. This ID can be used to request the keys from the Join Server in case the are lost. |
| `f_nwk_s_int_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Forwarding Network Session Integrity Key (or Network Session Key in 1.0 compatibility mode). This key is stored by the (forwarding) Network Server. |
| `s_nwk_s_int_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Serving Network Session Integrity Key. This key is stored by the (serving) Network Server. |
| `nwk_s_enc_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Network Session Encryption Key. This key is stored by the (serving) Network Server. |
| `app_s_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The (encrypted) Application Session Key. This key is stored by the Application Server. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |

## <a name="ttn/lorawan/v3/lorawan.proto">File `ttn/lorawan/v3/lorawan.proto`</a>

### <a name="ttn.lorawan.v3.ADRAckDelayExponentValue">Message `ADRAckDelayExponentValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`ADRAckDelayExponent`](#ttn.lorawan.v3.ADRAckDelayExponent) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.ADRAckLimitExponentValue">Message `ADRAckLimitExponentValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`ADRAckLimitExponent`](#ttn.lorawan.v3.ADRAckLimitExponent) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.AggregatedDutyCycleValue">Message `AggregatedDutyCycleValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`AggregatedDutyCycle`](#ttn.lorawan.v3.AggregatedDutyCycle) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.CFList">Message `CFList`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `type` | [`CFListType`](#ttn.lorawan.v3.CFListType) |  |  |
| `freq` | [`uint32`](#uint32) | repeated | Frequencies to be broadcasted, in hecto-Hz. These values are broadcasted as 24 bits unsigned integers. This field should not contain default values. |
| `ch_masks` | [`bool`](#bool) | repeated | ChMasks controlling the channels to be used. Length of this field must be equal to the amount of uplink channels defined by the selected frequency plan. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `type` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.ClassBCGatewayIdentifiers">Message `ClassBCGatewayIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `antenna_index` | [`uint32`](#uint32) |  |  |
| `group_index` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DLSettings">Message `DLSettings`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rx1_dr_offset` | [`DataRateOffset`](#ttn.lorawan.v3.DataRateOffset) |  |  |
| `rx2_dr` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `opt_neg` | [`bool`](#bool) |  | OptNeg is set if Network Server implements LoRaWAN 1.1 or greater. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rx1_dr_offset` | <p>`enum.defined_only`: `true`</p> |
| `rx2_dr` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.DataRate">Message `DataRate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lora` | [`LoRaDataRate`](#ttn.lorawan.v3.LoRaDataRate) |  |  |
| `fsk` | [`FSKDataRate`](#ttn.lorawan.v3.FSKDataRate) |  |  |
| `lrfhss` | [`LRFHSSDataRate`](#ttn.lorawan.v3.LRFHSSDataRate) |  |  |

### <a name="ttn.lorawan.v3.DataRateIndexValue">Message `DataRateIndexValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.DataRateOffsetValue">Message `DataRateOffsetValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`DataRateOffset`](#ttn.lorawan.v3.DataRateOffset) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.DeviceEIRPValue">Message `DeviceEIRPValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`DeviceEIRP`](#ttn.lorawan.v3.DeviceEIRP) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.DownlinkPath">Message `DownlinkPath`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uplink_token` | [`bytes`](#bytes) |  |  |
| `fixed` | [`GatewayAntennaIdentifiers`](#ttn.lorawan.v3.GatewayAntennaIdentifiers) |  |  |

### <a name="ttn.lorawan.v3.FCtrl">Message `FCtrl`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `adr` | [`bool`](#bool) |  |  |
| `adr_ack_req` | [`bool`](#bool) |  | Only on uplink. |
| `ack` | [`bool`](#bool) |  |  |
| `f_pending` | [`bool`](#bool) |  | Only on downlink. |
| `class_b` | [`bool`](#bool) |  | Only on uplink. |

### <a name="ttn.lorawan.v3.FHDR">Message `FHDR`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `f_ctrl` | [`FCtrl`](#ttn.lorawan.v3.FCtrl) |  |  |
| `f_cnt` | [`uint32`](#uint32) |  |  |
| `f_opts` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `f_ctrl` | <p>`message.required`: `true`</p> |
| `f_cnt` | <p>`uint32.lte`: `65535`</p> |
| `f_opts` | <p>`bytes.max_len`: `15`</p> |

### <a name="ttn.lorawan.v3.FSKDataRate">Message `FSKDataRate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bit_rate` | [`uint32`](#uint32) |  | Bit rate (bps). |

### <a name="ttn.lorawan.v3.FrequencyValue">Message `FrequencyValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`uint64`](#uint64) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.GatewayAntennaIdentifiers">Message `GatewayAntennaIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `antenna_index` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.JoinAcceptPayload">Message `JoinAcceptPayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `encrypted` | [`bytes`](#bytes) |  |  |
| `join_nonce` | [`bytes`](#bytes) |  |  |
| `net_id` | [`bytes`](#bytes) |  |  |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `dl_settings` | [`DLSettings`](#ttn.lorawan.v3.DLSettings) |  |  |
| `rx_delay` | [`RxDelay`](#ttn.lorawan.v3.RxDelay) |  |  |
| `cf_list` | [`CFList`](#ttn.lorawan.v3.CFList) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_nonce` | <p>`bytes.len`: `3`</p> |
| `net_id` | <p>`bytes.len`: `3`</p> |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `dl_settings` | <p>`message.required`: `true`</p> |
| `rx_delay` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.JoinRequestPayload">Message `JoinRequestPayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `dev_eui` | [`bytes`](#bytes) |  |  |
| `dev_nonce` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |
| `dev_eui` | <p>`bytes.len`: `8`</p> |
| `dev_nonce` | <p>`bytes.len`: `2`</p> |

### <a name="ttn.lorawan.v3.LRFHSSDataRate">Message `LRFHSSDataRate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `modulation_type` | [`uint32`](#uint32) |  |  |
| `operating_channel_width` | [`uint32`](#uint32) |  | Operating Channel Width (Hz). |
| `coding_rate` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.LoRaDataRate">Message `LoRaDataRate`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bandwidth` | [`uint32`](#uint32) |  | Bandwidth (Hz). |
| `spreading_factor` | [`uint32`](#uint32) |  |  |
| `coding_rate` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.MACCommand">Message `MACCommand`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cid` | [`MACCommandIdentifier`](#ttn.lorawan.v3.MACCommandIdentifier) |  |  |
| `raw_payload` | [`bytes`](#bytes) |  |  |
| `reset_ind` | [`MACCommand.ResetInd`](#ttn.lorawan.v3.MACCommand.ResetInd) |  |  |
| `reset_conf` | [`MACCommand.ResetConf`](#ttn.lorawan.v3.MACCommand.ResetConf) |  |  |
| `link_check_ans` | [`MACCommand.LinkCheckAns`](#ttn.lorawan.v3.MACCommand.LinkCheckAns) |  |  |
| `link_adr_req` | [`MACCommand.LinkADRReq`](#ttn.lorawan.v3.MACCommand.LinkADRReq) |  |  |
| `link_adr_ans` | [`MACCommand.LinkADRAns`](#ttn.lorawan.v3.MACCommand.LinkADRAns) |  |  |
| `duty_cycle_req` | [`MACCommand.DutyCycleReq`](#ttn.lorawan.v3.MACCommand.DutyCycleReq) |  |  |
| `rx_param_setup_req` | [`MACCommand.RxParamSetupReq`](#ttn.lorawan.v3.MACCommand.RxParamSetupReq) |  |  |
| `rx_param_setup_ans` | [`MACCommand.RxParamSetupAns`](#ttn.lorawan.v3.MACCommand.RxParamSetupAns) |  |  |
| `dev_status_ans` | [`MACCommand.DevStatusAns`](#ttn.lorawan.v3.MACCommand.DevStatusAns) |  |  |
| `new_channel_req` | [`MACCommand.NewChannelReq`](#ttn.lorawan.v3.MACCommand.NewChannelReq) |  |  |
| `new_channel_ans` | [`MACCommand.NewChannelAns`](#ttn.lorawan.v3.MACCommand.NewChannelAns) |  |  |
| `dl_channel_req` | [`MACCommand.DLChannelReq`](#ttn.lorawan.v3.MACCommand.DLChannelReq) |  |  |
| `dl_channel_ans` | [`MACCommand.DLChannelAns`](#ttn.lorawan.v3.MACCommand.DLChannelAns) |  |  |
| `rx_timing_setup_req` | [`MACCommand.RxTimingSetupReq`](#ttn.lorawan.v3.MACCommand.RxTimingSetupReq) |  |  |
| `tx_param_setup_req` | [`MACCommand.TxParamSetupReq`](#ttn.lorawan.v3.MACCommand.TxParamSetupReq) |  |  |
| `rekey_ind` | [`MACCommand.RekeyInd`](#ttn.lorawan.v3.MACCommand.RekeyInd) |  |  |
| `rekey_conf` | [`MACCommand.RekeyConf`](#ttn.lorawan.v3.MACCommand.RekeyConf) |  |  |
| `adr_param_setup_req` | [`MACCommand.ADRParamSetupReq`](#ttn.lorawan.v3.MACCommand.ADRParamSetupReq) |  |  |
| `device_time_ans` | [`MACCommand.DeviceTimeAns`](#ttn.lorawan.v3.MACCommand.DeviceTimeAns) |  |  |
| `force_rejoin_req` | [`MACCommand.ForceRejoinReq`](#ttn.lorawan.v3.MACCommand.ForceRejoinReq) |  |  |
| `rejoin_param_setup_req` | [`MACCommand.RejoinParamSetupReq`](#ttn.lorawan.v3.MACCommand.RejoinParamSetupReq) |  |  |
| `rejoin_param_setup_ans` | [`MACCommand.RejoinParamSetupAns`](#ttn.lorawan.v3.MACCommand.RejoinParamSetupAns) |  |  |
| `ping_slot_info_req` | [`MACCommand.PingSlotInfoReq`](#ttn.lorawan.v3.MACCommand.PingSlotInfoReq) |  |  |
| `ping_slot_channel_req` | [`MACCommand.PingSlotChannelReq`](#ttn.lorawan.v3.MACCommand.PingSlotChannelReq) |  |  |
| `ping_slot_channel_ans` | [`MACCommand.PingSlotChannelAns`](#ttn.lorawan.v3.MACCommand.PingSlotChannelAns) |  |  |
| `beacon_timing_ans` | [`MACCommand.BeaconTimingAns`](#ttn.lorawan.v3.MACCommand.BeaconTimingAns) |  |  |
| `beacon_freq_req` | [`MACCommand.BeaconFreqReq`](#ttn.lorawan.v3.MACCommand.BeaconFreqReq) |  |  |
| `beacon_freq_ans` | [`MACCommand.BeaconFreqAns`](#ttn.lorawan.v3.MACCommand.BeaconFreqAns) |  |  |
| `device_mode_ind` | [`MACCommand.DeviceModeInd`](#ttn.lorawan.v3.MACCommand.DeviceModeInd) |  |  |
| `device_mode_conf` | [`MACCommand.DeviceModeConf`](#ttn.lorawan.v3.MACCommand.DeviceModeConf) |  |  |
| `relay_conf_req` | [`MACCommand.RelayConfReq`](#ttn.lorawan.v3.MACCommand.RelayConfReq) |  |  |
| `relay_conf_ans` | [`MACCommand.RelayConfAns`](#ttn.lorawan.v3.MACCommand.RelayConfAns) |  |  |
| `relay_end_device_conf_req` | [`MACCommand.RelayEndDeviceConfReq`](#ttn.lorawan.v3.MACCommand.RelayEndDeviceConfReq) |  |  |
| `relay_end_device_conf_ans` | [`MACCommand.RelayEndDeviceConfAns`](#ttn.lorawan.v3.MACCommand.RelayEndDeviceConfAns) |  |  |
| `relay_update_uplink_list_req` | [`MACCommand.RelayUpdateUplinkListReq`](#ttn.lorawan.v3.MACCommand.RelayUpdateUplinkListReq) |  |  |
| `relay_update_uplink_list_ans` | [`MACCommand.RelayUpdateUplinkListAns`](#ttn.lorawan.v3.MACCommand.RelayUpdateUplinkListAns) |  |  |
| `relay_ctrl_uplink_list_req` | [`MACCommand.RelayCtrlUplinkListReq`](#ttn.lorawan.v3.MACCommand.RelayCtrlUplinkListReq) |  |  |
| `relay_ctrl_uplink_list_ans` | [`MACCommand.RelayCtrlUplinkListAns`](#ttn.lorawan.v3.MACCommand.RelayCtrlUplinkListAns) |  |  |
| `relay_configure_fwd_limit_req` | [`MACCommand.RelayConfigureFwdLimitReq`](#ttn.lorawan.v3.MACCommand.RelayConfigureFwdLimitReq) |  |  |
| `relay_configure_fwd_limit_ans` | [`MACCommand.RelayConfigureFwdLimitAns`](#ttn.lorawan.v3.MACCommand.RelayConfigureFwdLimitAns) |  |  |
| `relay_notify_new_end_device_req` | [`MACCommand.RelayNotifyNewEndDeviceReq`](#ttn.lorawan.v3.MACCommand.RelayNotifyNewEndDeviceReq) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `cid` | <p>`enum.defined_only`: `true`</p><p>`enum.not_in`: `[0]`</p> |

### <a name="ttn.lorawan.v3.MACCommand.ADRParamSetupReq">Message `MACCommand.ADRParamSetupReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `adr_ack_limit_exponent` | [`ADRAckLimitExponent`](#ttn.lorawan.v3.ADRAckLimitExponent) |  | Exponent e that configures the ADR_ACK_LIMIT = 2^e messages. |
| `adr_ack_delay_exponent` | [`ADRAckDelayExponent`](#ttn.lorawan.v3.ADRAckDelayExponent) |  | Exponent e that configures the ADR_ACK_DELAY = 2^e messages. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `adr_ack_limit_exponent` | <p>`enum.defined_only`: `true`</p> |
| `adr_ack_delay_exponent` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.BeaconFreqAns">Message `MACCommand.BeaconFreqAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.BeaconFreqReq">Message `MACCommand.BeaconFreqReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency` | [`uint64`](#uint64) |  | Frequency of the Class B beacons (Hz). |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `frequency` | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.MACCommand.BeaconTimingAns">Message `MACCommand.BeaconTimingAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delay` | [`uint32`](#uint32) |  | (uint16) See LoRaWAN specification. |
| `channel_index` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `delay` | <p>`uint32.lte`: `65535`</p> |
| `channel_index` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.MACCommand.DLChannelAns">Message `MACCommand.DLChannelAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel_index_ack` | [`bool`](#bool) |  |  |
| `frequency_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.DLChannelReq">Message `MACCommand.DLChannelReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel_index` | [`uint32`](#uint32) |  |  |
| `frequency` | [`uint64`](#uint64) |  | Downlink channel frequency (Hz). |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `channel_index` | <p>`uint32.lte`: `255`</p> |
| `frequency` | <p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.MACCommand.DevStatusAns">Message `MACCommand.DevStatusAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `battery` | [`uint32`](#uint32) |  | Device battery status. 0 indicates that the device is connected to an external power source. 1..254 indicates a battery level. 255 indicates that the device was not able to measure the battery level. |
| `margin` | [`int32`](#int32) |  | SNR of the last downlink (dB; [-32, +31]). |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `battery` | <p>`uint32.lte`: `255`</p> |
| `margin` | <p>`int32.lte`: `31`</p><p>`int32.gte`: `-32`</p> |

### <a name="ttn.lorawan.v3.MACCommand.DeviceModeConf">Message `MACCommand.DeviceModeConf`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class` | [`Class`](#ttn.lorawan.v3.Class) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `class` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.DeviceModeInd">Message `MACCommand.DeviceModeInd`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class` | [`Class`](#ttn.lorawan.v3.Class) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `class` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.DeviceTimeAns">Message `MACCommand.DeviceTimeAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `time` | <p>`timestamp.required`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.DutyCycleReq">Message `MACCommand.DutyCycleReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_duty_cycle` | [`AggregatedDutyCycle`](#ttn.lorawan.v3.AggregatedDutyCycle) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `max_duty_cycle` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.ForceRejoinReq">Message `MACCommand.ForceRejoinReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rejoin_type` | [`RejoinRequestType`](#ttn.lorawan.v3.RejoinRequestType) |  |  |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `max_retries` | [`uint32`](#uint32) |  |  |
| `period_exponent` | [`RejoinPeriodExponent`](#ttn.lorawan.v3.RejoinPeriodExponent) |  | Exponent e that configures the rejoin period = 32 * 2^e + rand(0,32) seconds. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rejoin_type` | <p>`enum.defined_only`: `true`</p> |
| `data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `max_retries` | <p>`uint32.lte`: `7`</p> |
| `period_exponent` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.LinkADRAns">Message `MACCommand.LinkADRAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel_mask_ack` | [`bool`](#bool) |  |  |
| `data_rate_index_ack` | [`bool`](#bool) |  |  |
| `tx_power_index_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.LinkADRReq">Message `MACCommand.LinkADRReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `tx_power_index` | [`uint32`](#uint32) |  |  |
| `channel_mask` | [`bool`](#bool) | repeated |  |
| `channel_mask_control` | [`uint32`](#uint32) |  |  |
| `nb_trans` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `tx_power_index` | <p>`uint32.lte`: `15`</p> |
| `channel_mask` | <p>`repeated.max_items`: `16`</p> |
| `channel_mask_control` | <p>`uint32.lte`: `7`</p> |
| `nb_trans` | <p>`uint32.lte`: `15`</p> |

### <a name="ttn.lorawan.v3.MACCommand.LinkCheckAns">Message `MACCommand.LinkCheckAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `margin` | [`uint32`](#uint32) |  | Indicates the link margin in dB of the received LinkCheckReq, relative to the demodulation floor. |
| `gateway_count` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `margin` | <p>`uint32.lte`: `254`</p> |
| `gateway_count` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.MACCommand.NewChannelAns">Message `MACCommand.NewChannelAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency_ack` | [`bool`](#bool) |  |  |
| `data_rate_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.NewChannelReq">Message `MACCommand.NewChannelReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel_index` | [`uint32`](#uint32) |  |  |
| `frequency` | [`uint64`](#uint64) |  | Channel frequency (Hz). |
| `min_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `max_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `channel_index` | <p>`uint32.lte`: `255`</p> |
| `frequency` | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p> |
| `min_data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `max_data_rate_index` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.PingSlotChannelAns">Message `MACCommand.PingSlotChannelAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency_ack` | [`bool`](#bool) |  |  |
| `data_rate_index_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.PingSlotChannelReq">Message `MACCommand.PingSlotChannelReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency` | [`uint64`](#uint64) |  | Ping slot channel frequency (Hz). |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `frequency` | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p> |
| `data_rate_index` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.PingSlotInfoReq">Message `MACCommand.PingSlotInfoReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `period` | [`PingSlotPeriod`](#ttn.lorawan.v3.PingSlotPeriod) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `period` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RejoinParamSetupAns">Message `MACCommand.RejoinParamSetupAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_time_exponent_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RejoinParamSetupReq">Message `MACCommand.RejoinParamSetupReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_count_exponent` | [`RejoinCountExponent`](#ttn.lorawan.v3.RejoinCountExponent) |  | Exponent e that configures the rejoin counter = 2^(e+4) messages. |
| `max_time_exponent` | [`RejoinTimeExponent`](#ttn.lorawan.v3.RejoinTimeExponent) |  | Exponent e that configures the rejoin timer = 2^(e+10) seconds. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `max_count_exponent` | <p>`enum.defined_only`: `true`</p> |
| `max_time_exponent` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RekeyConf">Message `MACCommand.RekeyConf`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minor_version` | [`Minor`](#ttn.lorawan.v3.Minor) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `minor_version` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RekeyInd">Message `MACCommand.RekeyInd`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minor_version` | [`Minor`](#ttn.lorawan.v3.Minor) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `minor_version` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RelayConfAns">Message `MACCommand.RelayConfAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `second_channel_frequency_ack` | [`bool`](#bool) |  |  |
| `second_channel_ack_offset_ack` | [`bool`](#bool) |  |  |
| `second_channel_data_rate_index_ack` | [`bool`](#bool) |  |  |
| `second_channel_index_ack` | [`bool`](#bool) |  |  |
| `default_channel_index_ack` | [`bool`](#bool) |  |  |
| `cad_periodicity_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RelayConfReq">Message `MACCommand.RelayConfReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `configuration` | [`MACCommand.RelayConfReq.Configuration`](#ttn.lorawan.v3.MACCommand.RelayConfReq.Configuration) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RelayConfReq.Configuration">Message `MACCommand.RelayConfReq.Configuration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `second_channel` | [`RelaySecondChannel`](#ttn.lorawan.v3.RelaySecondChannel) |  |  |
| `default_channel_index` | [`uint32`](#uint32) |  |  |
| `cad_periodicity` | [`RelayCADPeriodicity`](#ttn.lorawan.v3.RelayCADPeriodicity) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `default_channel_index` | <p>`uint32.lte`: `255`</p> |
| `cad_periodicity` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RelayConfigureFwdLimitAns">Message `MACCommand.RelayConfigureFwdLimitAns`</a>

### <a name="ttn.lorawan.v3.MACCommand.RelayConfigureFwdLimitReq">Message `MACCommand.RelayConfigureFwdLimitReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reset_limit_counter` | [`RelayResetLimitCounter`](#ttn.lorawan.v3.RelayResetLimitCounter) |  |  |
| `join_request_limits` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  |  |
| `notify_limits` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  |  |
| `global_uplink_limits` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  |  |
| `overall_limits` | [`RelayForwardLimits`](#ttn.lorawan.v3.RelayForwardLimits) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RelayCtrlUplinkListAns">Message `MACCommand.RelayCtrlUplinkListAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rule_index_ack` | [`bool`](#bool) |  |  |
| `w_f_cnt` | [`uint32`](#uint32) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RelayCtrlUplinkListReq">Message `MACCommand.RelayCtrlUplinkListReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rule_index` | [`uint32`](#uint32) |  |  |
| `action` | [`RelayCtrlUplinkListAction`](#ttn.lorawan.v3.RelayCtrlUplinkListAction) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rule_index` | <p>`uint32.lte`: `15`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RelayEndDeviceConfAns">Message `MACCommand.RelayEndDeviceConfAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `second_channel_frequency_ack` | [`bool`](#bool) |  |  |
| `second_channel_data_rate_index_ack` | [`bool`](#bool) |  |  |
| `second_channel_index_ack` | [`bool`](#bool) |  |  |
| `backoff_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RelayEndDeviceConfReq">Message `MACCommand.RelayEndDeviceConfReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `configuration` | [`MACCommand.RelayEndDeviceConfReq.Configuration`](#ttn.lorawan.v3.MACCommand.RelayEndDeviceConfReq.Configuration) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RelayEndDeviceConfReq.Configuration">Message `MACCommand.RelayEndDeviceConfReq.Configuration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `always` | [`RelayEndDeviceAlwaysMode`](#ttn.lorawan.v3.RelayEndDeviceAlwaysMode) |  |  |
| `dynamic` | [`RelayEndDeviceDynamicMode`](#ttn.lorawan.v3.RelayEndDeviceDynamicMode) |  |  |
| `end_device_controlled` | [`RelayEndDeviceControlledMode`](#ttn.lorawan.v3.RelayEndDeviceControlledMode) |  |  |
| `backoff` | [`uint32`](#uint32) |  |  |
| `second_channel` | [`RelaySecondChannel`](#ttn.lorawan.v3.RelaySecondChannel) |  |  |
| `serving_device_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `backoff` | <p>`uint32.lte`: `63`</p> |
| `serving_device_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RelayNotifyNewEndDeviceReq">Message `MACCommand.RelayNotifyNewEndDeviceReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `snr` | [`int32`](#int32) |  |  |
| `rssi` | [`int32`](#int32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `snr` | <p>`int32.lte`: `11`</p><p>`int32.gte`: `-20`</p> |
| `rssi` | <p>`int32.lte`: `-15`</p><p>`int32.gte`: `-142`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RelayUpdateUplinkListAns">Message `MACCommand.RelayUpdateUplinkListAns`</a>

### <a name="ttn.lorawan.v3.MACCommand.RelayUpdateUplinkListReq">Message `MACCommand.RelayUpdateUplinkListReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rule_index` | [`uint32`](#uint32) |  |  |
| `forward_limits` | [`RelayUplinkForwardLimits`](#ttn.lorawan.v3.RelayUplinkForwardLimits) |  |  |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `w_f_cnt` | [`uint32`](#uint32) |  |  |
| `root_wor_s_key` | [`bytes`](#bytes) |  |  |
| `device_id` | [`string`](#string) |  |  |
| `session_key_id` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rule_index` | <p>`uint32.lte`: `15`</p> |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `root_wor_s_key` | <p>`bytes.len`: `16`</p> |
| `device_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.MACCommand.ResetConf">Message `MACCommand.ResetConf`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minor_version` | [`Minor`](#ttn.lorawan.v3.Minor) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `minor_version` | <p>`enum.defined_only`: `true`</p><p>`enum.in`: `[1]`</p> |

### <a name="ttn.lorawan.v3.MACCommand.ResetInd">Message `MACCommand.ResetInd`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minor_version` | [`Minor`](#ttn.lorawan.v3.Minor) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `minor_version` | <p>`enum.defined_only`: `true`</p><p>`enum.in`: `[1]`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RxParamSetupAns">Message `MACCommand.RxParamSetupAns`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rx2_data_rate_index_ack` | [`bool`](#bool) |  |  |
| `rx1_data_rate_offset_ack` | [`bool`](#bool) |  |  |
| `rx2_frequency_ack` | [`bool`](#bool) |  |  |

### <a name="ttn.lorawan.v3.MACCommand.RxParamSetupReq">Message `MACCommand.RxParamSetupReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rx2_data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  |  |
| `rx1_data_rate_offset` | [`DataRateOffset`](#ttn.lorawan.v3.DataRateOffset) |  |  |
| `rx2_frequency` | [`uint64`](#uint64) |  | Rx2 frequency (Hz). |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rx2_data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `rx1_data_rate_offset` | <p>`enum.defined_only`: `true`</p> |
| `rx2_frequency` | <p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.MACCommand.RxTimingSetupReq">Message `MACCommand.RxTimingSetupReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delay` | [`RxDelay`](#ttn.lorawan.v3.RxDelay) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `delay` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommand.TxParamSetupReq">Message `MACCommand.TxParamSetupReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_eirp_index` | [`DeviceEIRP`](#ttn.lorawan.v3.DeviceEIRP) |  | Indicates the maximum EIRP value in dBm, indexed by the following vector: [ 8 10 12 13 14 16 18 20 21 24 26 27 29 30 33 36 ] |
| `uplink_dwell_time` | [`bool`](#bool) |  |  |
| `downlink_dwell_time` | [`bool`](#bool) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `max_eirp_index` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.MACCommands">Message `MACCommands`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commands` | [`MACCommand`](#ttn.lorawan.v3.MACCommand) | repeated |  |

### <a name="ttn.lorawan.v3.MACPayload">Message `MACPayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `f_hdr` | [`FHDR`](#ttn.lorawan.v3.FHDR) |  |  |
| `f_port` | [`uint32`](#uint32) |  |  |
| `frm_payload` | [`bytes`](#bytes) |  |  |
| `decoded_payload` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  |  |
| `full_f_cnt` | [`uint32`](#uint32) |  | Full 32-bit FCnt value. Used internally by Network Server. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `f_hdr` | <p>`message.required`: `true`</p> |
| `f_port` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.MHDR">Message `MHDR`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `m_type` | [`MType`](#ttn.lorawan.v3.MType) |  |  |
| `major` | [`Major`](#ttn.lorawan.v3.Major) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `m_type` | <p>`enum.defined_only`: `true`</p> |
| `major` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.Message">Message `Message`</a>

Message represents a LoRaWAN message

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `m_hdr` | [`MHDR`](#ttn.lorawan.v3.MHDR) |  |  |
| `mic` | [`bytes`](#bytes) |  |  |
| `mac_payload` | [`MACPayload`](#ttn.lorawan.v3.MACPayload) |  |  |
| `join_request_payload` | [`JoinRequestPayload`](#ttn.lorawan.v3.JoinRequestPayload) |  |  |
| `join_accept_payload` | [`JoinAcceptPayload`](#ttn.lorawan.v3.JoinAcceptPayload) |  |  |
| `rejoin_request_payload` | [`RejoinRequestPayload`](#ttn.lorawan.v3.RejoinRequestPayload) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `m_hdr` | <p>`message.required`: `true`</p> |
| `mic` | <p>`bytes.min_len`: `0`</p><p>`bytes.max_len`: `4`</p> |

### <a name="ttn.lorawan.v3.PingSlotPeriodValue">Message `PingSlotPeriodValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`PingSlotPeriod`](#ttn.lorawan.v3.PingSlotPeriod) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.RejoinRequestPayload">Message `RejoinRequestPayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rejoin_type` | [`RejoinRequestType`](#ttn.lorawan.v3.RejoinRequestType) |  |  |
| `net_id` | [`bytes`](#bytes) |  |  |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `dev_eui` | [`bytes`](#bytes) |  |  |
| `rejoin_cnt` | [`uint32`](#uint32) |  | Contains RJCount0 or RJCount1 depending on rejoin_type. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rejoin_type` | <p>`enum.defined_only`: `true`</p> |
| `net_id` | <p>`bytes.len`: `3`</p> |
| `join_eui` | <p>`bytes.len`: `8`</p> |
| `dev_eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.RelayEndDeviceAlwaysMode">Message `RelayEndDeviceAlwaysMode`</a>

### <a name="ttn.lorawan.v3.RelayEndDeviceControlledMode">Message `RelayEndDeviceControlledMode`</a>

### <a name="ttn.lorawan.v3.RelayEndDeviceDynamicMode">Message `RelayEndDeviceDynamicMode`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `smart_enable_level` | [`RelaySmartEnableLevel`](#ttn.lorawan.v3.RelaySmartEnableLevel) |  | The number of consecutive uplinks without a valid downlink before the end device attempts to use the relay mode to transmit messages. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `smart_enable_level` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.RelayForwardDownlinkReq">Message `RelayForwardDownlinkReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_payload` | [`bytes`](#bytes) |  |  |

### <a name="ttn.lorawan.v3.RelayForwardLimits">Message `RelayForwardLimits`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bucket_size` | [`RelayLimitBucketSize`](#ttn.lorawan.v3.RelayLimitBucketSize) |  | The multiplier used to compute the total bucket size for the limits. The multiplier is multiplied by the reload rate in order to compute the total bucket size. |
| `reload_rate` | [`uint32`](#uint32) |  | The number of tokens which are replenished in the bucket every hour. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `bucket_size` | <p>`enum.defined_only`: `true`</p> |
| `reload_rate` | <p>`uint32.lte`: `126`</p> |

### <a name="ttn.lorawan.v3.RelayForwardUplinkReq">Message `RelayForwardUplinkReq`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data_rate` | [`DataRate`](#ttn.lorawan.v3.DataRate) |  |  |
| `snr` | [`int32`](#int32) |  |  |
| `rssi` | [`int32`](#int32) |  |  |
| `wor_channel` | [`RelayWORChannel`](#ttn.lorawan.v3.RelayWORChannel) |  |  |
| `frequency` | [`uint64`](#uint64) |  | Uplink channel frequency (Hz). |
| `raw_payload` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `data_rate` | <p>`message.required`: `true`</p> |
| `snr` | <p>`int32.lte`: `11`</p><p>`int32.gte`: `-20`</p> |
| `rssi` | <p>`int32.lte`: `-15`</p><p>`int32.gte`: `-142`</p> |
| `wor_channel` | <p>`enum.defined_only`: `true`</p> |
| `frequency` | <p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.RelaySecondChannel">Message `RelaySecondChannel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ack_offset` | [`RelaySecondChAckOffset`](#ttn.lorawan.v3.RelaySecondChAckOffset) |  | The frequency (Hz) offset used for the WOR acknowledgement. |
| `data_rate_index` | [`DataRateIndex`](#ttn.lorawan.v3.DataRateIndex) |  | The data rate index used by the WOR and WOR acknowledgement. |
| `frequency` | [`uint64`](#uint64) |  | The frequency (Hz) used by the wake on radio message. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ack_offset` | <p>`enum.defined_only`: `true`</p> |
| `data_rate_index` | <p>`enum.defined_only`: `true`</p> |
| `frequency` | <p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.RelayUplinkForwardLimits">Message `RelayUplinkForwardLimits`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bucket_size` | [`RelayLimitBucketSize`](#ttn.lorawan.v3.RelayLimitBucketSize) |  | The multiplier used to compute the total bucket size for the limits. The multiplier is multiplied by the reload rate in order to compute the total bucket size. |
| `reload_rate` | [`uint32`](#uint32) |  | The number of tokens which are replenished in the bucket every hour. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `bucket_size` | <p>`enum.defined_only`: `true`</p> |
| `reload_rate` | <p>`uint32.lte`: `62`</p> |

### <a name="ttn.lorawan.v3.RelayUplinkToken">Message `RelayUplinkToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `session_key_id` | [`bytes`](#bytes) |  |  |
| `full_f_cnt` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.RxDelayValue">Message `RxDelayValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`RxDelay`](#ttn.lorawan.v3.RxDelay) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.TxRequest">Message `TxRequest`</a>

TxRequest is a request for transmission.
If sent to a roaming partner, this request is used to generate the DLMetadata Object (see Backend Interfaces 1.0, Table 22).
If the gateway has a scheduler, this request is sent to the gateway, in the order of gateway_ids.
Otherwise, the Gateway Server attempts to schedule the request and creates the TxSettings.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class` | [`Class`](#ttn.lorawan.v3.Class) |  |  |
| `downlink_paths` | [`DownlinkPath`](#ttn.lorawan.v3.DownlinkPath) | repeated | Downlink paths used to select a gateway for downlink. In class A, the downlink paths are required to only contain uplink tokens. In class B and C, the downlink paths may contain uplink tokens and fixed gateways antenna identifiers. |
| `rx1_delay` | [`RxDelay`](#ttn.lorawan.v3.RxDelay) |  | Rx1 delay (Rx2 delay is Rx1 delay + 1 second). |
| `rx1_data_rate` | [`DataRate`](#ttn.lorawan.v3.DataRate) |  | LoRaWAN data rate for Rx1. |
| `rx1_frequency` | [`uint64`](#uint64) |  | Frequency (Hz) for Rx1. |
| `rx2_data_rate` | [`DataRate`](#ttn.lorawan.v3.DataRate) |  | LoRaWAN data rate for Rx2. |
| `rx2_frequency` | [`uint64`](#uint64) |  | Frequency (Hz) for Rx2. |
| `priority` | [`TxSchedulePriority`](#ttn.lorawan.v3.TxSchedulePriority) |  | Priority for scheduling. Requests with a higher priority are allocated more channel time than messages with a lower priority, in duty-cycle limited regions. A priority of HIGH or higher sets the HiPriorityFlag in the DLMetadata Object. |
| `absolute_time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time when the downlink message should be transmitted. This value is only valid for class C downlink; class A downlink uses uplink tokens and class B downlink is scheduled on ping slots. This requires the gateway to have GPS time sychronization. If the absolute time is not set, the first available time will be used that does not conflict or violate regional limitations. |
| `frequency_plan_id` | [`string`](#string) |  | Frequency plan ID from which the frequencies in this message are retrieved. |
| `advanced` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rx1_delay` | <p>`enum.defined_only`: `true`</p> |
| `priority` | <p>`enum.defined_only`: `true`</p> |
| `frequency_plan_id` | <p>`string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.TxSettings">Message `TxSettings`</a>

TxSettings contains the settings for a transmission.
This message is used on both uplink and downlink.
On downlink, this is a scheduled transmission.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data_rate` | [`DataRate`](#ttn.lorawan.v3.DataRate) |  | Data rate. |
| `frequency` | [`uint64`](#uint64) |  | Frequency (Hz). |
| `enable_crc` | [`bool`](#bool) |  | Send a CRC in the packet; only on uplink; on downlink, CRC should not be enabled. |
| `timestamp` | [`uint32`](#uint32) |  | Timestamp of the gateway concentrator when the uplink message was received, or when the downlink message should be transmitted (microseconds). On downlink, set timestamp to 0 and time to null to use immediate scheduling. |
| `time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time of the gateway when the uplink message was received, or when the downlink message should be transmitted. For downlink, this requires the gateway to have GPS time synchronization. |
| `downlink` | [`TxSettings.Downlink`](#ttn.lorawan.v3.TxSettings.Downlink) |  | Transmission settings for downlink. |
| `concentrator_timestamp` | [`int64`](#int64) |  | Concentrator timestamp for the downlink as calculated by the Gateway Server scheduler. This value takes into account necessary offsets such as the RTT (Round Trip Time) and TOA (Time Of Arrival). This field is set and used only by the Gateway Server. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `data_rate` | <p>`message.required`: `true`</p> |
| `frequency` | <p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.TxSettings.Downlink">Message `TxSettings.Downlink`</a>

Transmission settings for downlink.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `antenna_index` | [`uint32`](#uint32) |  | Index of the antenna on which the uplink was received and/or downlink must be sent. |
| `tx_power` | [`float`](#float) |  | Transmission power (dBm). Only on downlink. |
| `invert_polarization` | [`bool`](#bool) |  | Invert LoRa polarization; false for LoRaWAN uplink, true for downlink. |

### <a name="ttn.lorawan.v3.UplinkToken">Message `UplinkToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`GatewayAntennaIdentifiers`](#ttn.lorawan.v3.GatewayAntennaIdentifiers) |  |  |
| `timestamp` | [`uint32`](#uint32) |  |  |
| `server_time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Absolute time observed by the server when the uplink message has been received. |
| `concentrator_time` | [`int64`](#int64) |  | Absolute concentrator time as observed by the Gateway Server, accounting for rollovers. |
| `gateway_time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Absolute time observed by the gateway when the uplink has been received. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ZeroableFrequencyValue">Message `ZeroableFrequencyValue`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [`uint64`](#uint64) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`uint64.lte`: `0`</p><p>`uint64.gte`: `100000`</p> |

### <a name="ttn.lorawan.v3.ADRAckDelayExponent">Enum `ADRAckDelayExponent`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `ADR_ACK_DELAY_1` | 0 |  |
| `ADR_ACK_DELAY_2` | 1 |  |
| `ADR_ACK_DELAY_4` | 2 |  |
| `ADR_ACK_DELAY_8` | 3 |  |
| `ADR_ACK_DELAY_16` | 4 |  |
| `ADR_ACK_DELAY_32` | 5 |  |
| `ADR_ACK_DELAY_64` | 6 |  |
| `ADR_ACK_DELAY_128` | 7 |  |
| `ADR_ACK_DELAY_256` | 8 |  |
| `ADR_ACK_DELAY_512` | 9 |  |
| `ADR_ACK_DELAY_1024` | 10 |  |
| `ADR_ACK_DELAY_2048` | 11 |  |
| `ADR_ACK_DELAY_4096` | 12 |  |
| `ADR_ACK_DELAY_8192` | 13 |  |
| `ADR_ACK_DELAY_16384` | 14 |  |
| `ADR_ACK_DELAY_32768` | 15 |  |

### <a name="ttn.lorawan.v3.ADRAckLimitExponent">Enum `ADRAckLimitExponent`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `ADR_ACK_LIMIT_1` | 0 |  |
| `ADR_ACK_LIMIT_2` | 1 |  |
| `ADR_ACK_LIMIT_4` | 2 |  |
| `ADR_ACK_LIMIT_8` | 3 |  |
| `ADR_ACK_LIMIT_16` | 4 |  |
| `ADR_ACK_LIMIT_32` | 5 |  |
| `ADR_ACK_LIMIT_64` | 6 |  |
| `ADR_ACK_LIMIT_128` | 7 |  |
| `ADR_ACK_LIMIT_256` | 8 |  |
| `ADR_ACK_LIMIT_512` | 9 |  |
| `ADR_ACK_LIMIT_1024` | 10 |  |
| `ADR_ACK_LIMIT_2048` | 11 |  |
| `ADR_ACK_LIMIT_4096` | 12 |  |
| `ADR_ACK_LIMIT_8192` | 13 |  |
| `ADR_ACK_LIMIT_16384` | 14 |  |
| `ADR_ACK_LIMIT_32768` | 15 |  |

### <a name="ttn.lorawan.v3.AggregatedDutyCycle">Enum `AggregatedDutyCycle`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DUTY_CYCLE_1` | 0 | 100%. |
| `DUTY_CYCLE_2` | 1 | 50%. |
| `DUTY_CYCLE_4` | 2 | 25%. |
| `DUTY_CYCLE_8` | 3 | 12.5%. |
| `DUTY_CYCLE_16` | 4 | 6.25%. |
| `DUTY_CYCLE_32` | 5 | 3.125%. |
| `DUTY_CYCLE_64` | 6 | 1.5625%. |
| `DUTY_CYCLE_128` | 7 | Roughly 0.781%. |
| `DUTY_CYCLE_256` | 8 | Roughly 0.390%. |
| `DUTY_CYCLE_512` | 9 | Roughly 0.195%. |
| `DUTY_CYCLE_1024` | 10 | Roughly 0.098%. |
| `DUTY_CYCLE_2048` | 11 | Roughly 0.049%. |
| `DUTY_CYCLE_4096` | 12 | Roughly 0.024%. |
| `DUTY_CYCLE_8192` | 13 | Roughly 0.012%. |
| `DUTY_CYCLE_16384` | 14 | Roughly 0.006%. |
| `DUTY_CYCLE_32768` | 15 | Roughly 0.003%. |

### <a name="ttn.lorawan.v3.CFListType">Enum `CFListType`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `FREQUENCIES` | 0 |  |
| `CHANNEL_MASKS` | 1 |  |

### <a name="ttn.lorawan.v3.Class">Enum `Class`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `CLASS_A` | 0 |  |
| `CLASS_B` | 1 |  |
| `CLASS_C` | 2 |  |

### <a name="ttn.lorawan.v3.DataRateIndex">Enum `DataRateIndex`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DATA_RATE_0` | 0 |  |
| `DATA_RATE_1` | 1 |  |
| `DATA_RATE_2` | 2 |  |
| `DATA_RATE_3` | 3 |  |
| `DATA_RATE_4` | 4 |  |
| `DATA_RATE_5` | 5 |  |
| `DATA_RATE_6` | 6 |  |
| `DATA_RATE_7` | 7 |  |
| `DATA_RATE_8` | 8 |  |
| `DATA_RATE_9` | 9 |  |
| `DATA_RATE_10` | 10 |  |
| `DATA_RATE_11` | 11 |  |
| `DATA_RATE_12` | 12 |  |
| `DATA_RATE_13` | 13 |  |
| `DATA_RATE_14` | 14 |  |
| `DATA_RATE_15` | 15 |  |

### <a name="ttn.lorawan.v3.DataRateOffset">Enum `DataRateOffset`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DATA_RATE_OFFSET_0` | 0 |  |
| `DATA_RATE_OFFSET_1` | 1 |  |
| `DATA_RATE_OFFSET_2` | 2 |  |
| `DATA_RATE_OFFSET_3` | 3 |  |
| `DATA_RATE_OFFSET_4` | 4 |  |
| `DATA_RATE_OFFSET_5` | 5 |  |
| `DATA_RATE_OFFSET_6` | 6 |  |
| `DATA_RATE_OFFSET_7` | 7 |  |

### <a name="ttn.lorawan.v3.DeviceEIRP">Enum `DeviceEIRP`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DEVICE_EIRP_8` | 0 | 8 dBm. |
| `DEVICE_EIRP_10` | 1 | 10 dBm. |
| `DEVICE_EIRP_12` | 2 | 12 dBm. |
| `DEVICE_EIRP_13` | 3 | 13 dBm. |
| `DEVICE_EIRP_14` | 4 | 14 dBm. |
| `DEVICE_EIRP_16` | 5 | 16 dBm. |
| `DEVICE_EIRP_18` | 6 | 18 dBm. |
| `DEVICE_EIRP_20` | 7 | 20 dBm. |
| `DEVICE_EIRP_21` | 8 | 21 dBm. |
| `DEVICE_EIRP_24` | 9 | 24 dBm. |
| `DEVICE_EIRP_26` | 10 | 26 dBm. |
| `DEVICE_EIRP_27` | 11 | 27 dBm. |
| `DEVICE_EIRP_29` | 12 | 29 dBm. |
| `DEVICE_EIRP_30` | 13 | 30 dBm. |
| `DEVICE_EIRP_33` | 14 | 33 dBm. |
| `DEVICE_EIRP_36` | 15 | 36 dBm. |

### <a name="ttn.lorawan.v3.JoinRequestType">Enum `JoinRequestType`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `REJOIN_CONTEXT` | 0 | Resets DevAddr, Session Keys, Frame Counters, Radio Parameters. |
| `REJOIN_SESSION` | 1 | Equivalent to the initial JoinRequest. |
| `REJOIN_KEYS` | 2 | Resets DevAddr, Session Keys, Frame Counters, while keeping the Radio Parameters. |
| `JOIN` | 255 | Normal join-request. |

### <a name="ttn.lorawan.v3.MACCommandIdentifier">Enum `MACCommandIdentifier`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `CID_RFU_0` | 0 |  |
| `CID_RESET` | 1 |  |
| `CID_LINK_CHECK` | 2 |  |
| `CID_LINK_ADR` | 3 |  |
| `CID_DUTY_CYCLE` | 4 |  |
| `CID_RX_PARAM_SETUP` | 5 |  |
| `CID_DEV_STATUS` | 6 |  |
| `CID_NEW_CHANNEL` | 7 |  |
| `CID_RX_TIMING_SETUP` | 8 |  |
| `CID_TX_PARAM_SETUP` | 9 |  |
| `CID_DL_CHANNEL` | 10 |  |
| `CID_REKEY` | 11 |  |
| `CID_ADR_PARAM_SETUP` | 12 |  |
| `CID_DEVICE_TIME` | 13 |  |
| `CID_FORCE_REJOIN` | 14 |  |
| `CID_REJOIN_PARAM_SETUP` | 15 |  |
| `CID_PING_SLOT_INFO` | 16 |  |
| `CID_PING_SLOT_CHANNEL` | 17 |  |
| `CID_BEACON_TIMING` | 18 | Deprecated |
| `CID_BEACON_FREQ` | 19 |  |
| `CID_DEVICE_MODE` | 32 |  |
| `CID_RELAY_CONF` | 64 |  |
| `CID_RELAY_END_DEVICE_CONF` | 65 |  |
| `CID_RELAY_FILTER_LIST` | 66 |  |
| `CID_RELAY_UPDATE_UPLINK_LIST` | 67 |  |
| `CID_RELAY_CTRL_UPLINK_LIST` | 68 |  |
| `CID_RELAY_CONFIGURE_FWD_LIMIT` | 69 |  |
| `CID_RELAY_NOTIFY_NEW_END_DEVICE` | 70 |  |

### <a name="ttn.lorawan.v3.MACVersion">Enum `MACVersion`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `MAC_UNKNOWN` | 0 |  |
| `MAC_V1_0` | 1 |  |
| `MAC_V1_0_1` | 2 |  |
| `MAC_V1_0_2` | 3 |  |
| `MAC_V1_1` | 4 |  |
| `MAC_V1_0_3` | 5 |  |
| `MAC_V1_0_4` | 6 |  |

### <a name="ttn.lorawan.v3.MType">Enum `MType`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `JOIN_REQUEST` | 0 |  |
| `JOIN_ACCEPT` | 1 |  |
| `UNCONFIRMED_UP` | 2 |  |
| `UNCONFIRMED_DOWN` | 3 |  |
| `CONFIRMED_UP` | 4 |  |
| `CONFIRMED_DOWN` | 5 |  |
| `REJOIN_REQUEST` | 6 |  |
| `PROPRIETARY` | 7 |  |

### <a name="ttn.lorawan.v3.Major">Enum `Major`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `LORAWAN_R1` | 0 |  |

### <a name="ttn.lorawan.v3.Minor">Enum `Minor`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `MINOR_RFU_0` | 0 |  |
| `MINOR_1` | 1 |  |
| `MINOR_RFU_2` | 2 |  |
| `MINOR_RFU_3` | 3 |  |
| `MINOR_RFU_4` | 4 |  |
| `MINOR_RFU_5` | 5 |  |
| `MINOR_RFU_6` | 6 |  |
| `MINOR_RFU_7` | 7 |  |
| `MINOR_RFU_8` | 8 |  |
| `MINOR_RFU_9` | 9 |  |
| `MINOR_RFU_10` | 10 |  |
| `MINOR_RFU_11` | 11 |  |
| `MINOR_RFU_12` | 12 |  |
| `MINOR_RFU_13` | 13 |  |
| `MINOR_RFU_14` | 14 |  |
| `MINOR_RFU_15` | 15 |  |

### <a name="ttn.lorawan.v3.PHYVersion">Enum `PHYVersion`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `PHY_UNKNOWN` | 0 |  |
| `PHY_V1_0` | 1 |  |
| `TS001_V1_0` | 1 |  |
| `PHY_V1_0_1` | 2 |  |
| `TS001_V1_0_1` | 2 |  |
| `PHY_V1_0_2_REV_A` | 3 |  |
| `RP001_V1_0_2` | 3 |  |
| `PHY_V1_0_2_REV_B` | 4 |  |
| `RP001_V1_0_2_REV_B` | 4 |  |
| `PHY_V1_1_REV_A` | 5 |  |
| `RP001_V1_1_REV_A` | 5 |  |
| `PHY_V1_1_REV_B` | 6 |  |
| `RP001_V1_1_REV_B` | 6 |  |
| `PHY_V1_0_3_REV_A` | 7 |  |
| `RP001_V1_0_3_REV_A` | 7 |  |
| `RP002_V1_0_0` | 8 |  |
| `RP002_V1_0_1` | 9 |  |
| `RP002_V1_0_2` | 10 |  |
| `RP002_V1_0_3` | 11 |  |
| `RP002_V1_0_4` | 12 |  |

### <a name="ttn.lorawan.v3.PingSlotPeriod">Enum `PingSlotPeriod`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `PING_EVERY_1S` | 0 | Every second. |
| `PING_EVERY_2S` | 1 | Every 2 seconds. |
| `PING_EVERY_4S` | 2 | Every 4 seconds. |
| `PING_EVERY_8S` | 3 | Every 8 seconds. |
| `PING_EVERY_16S` | 4 | Every 16 seconds. |
| `PING_EVERY_32S` | 5 | Every 32 seconds. |
| `PING_EVERY_64S` | 6 | Every 64 seconds. |
| `PING_EVERY_128S` | 7 | Every 128 seconds. |

### <a name="ttn.lorawan.v3.RejoinCountExponent">Enum `RejoinCountExponent`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `REJOIN_COUNT_16` | 0 |  |
| `REJOIN_COUNT_32` | 1 |  |
| `REJOIN_COUNT_64` | 2 |  |
| `REJOIN_COUNT_128` | 3 |  |
| `REJOIN_COUNT_256` | 4 |  |
| `REJOIN_COUNT_512` | 5 |  |
| `REJOIN_COUNT_1024` | 6 |  |
| `REJOIN_COUNT_2048` | 7 |  |
| `REJOIN_COUNT_4096` | 8 |  |
| `REJOIN_COUNT_8192` | 9 |  |
| `REJOIN_COUNT_16384` | 10 |  |
| `REJOIN_COUNT_32768` | 11 |  |
| `REJOIN_COUNT_65536` | 12 |  |
| `REJOIN_COUNT_131072` | 13 |  |
| `REJOIN_COUNT_262144` | 14 |  |
| `REJOIN_COUNT_524288` | 15 |  |

### <a name="ttn.lorawan.v3.RejoinPeriodExponent">Enum `RejoinPeriodExponent`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `REJOIN_PERIOD_0` | 0 | Every 32 to 64 seconds. |
| `REJOIN_PERIOD_1` | 1 | Every 64 to 96 seconds. |
| `REJOIN_PERIOD_2` | 2 | Every 128 to 160 seconds. |
| `REJOIN_PERIOD_3` | 3 | Every 256 to 288 seconds. |
| `REJOIN_PERIOD_4` | 4 | Every 512 to 544 seconds. |
| `REJOIN_PERIOD_5` | 5 | Every 1024 to 1056 seconds. |
| `REJOIN_PERIOD_6` | 6 | Every 2048 to 2080 seconds. |
| `REJOIN_PERIOD_7` | 7 | Every 4096 to 4128 seconds. |

### <a name="ttn.lorawan.v3.RejoinRequestType">Enum `RejoinRequestType`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `CONTEXT` | 0 | Resets DevAddr, Session Keys, Frame Counters, Radio Parameters. |
| `SESSION` | 1 | Equivalent to the initial JoinRequest. |
| `KEYS` | 2 | Resets DevAddr, Session Keys, Frame Counters, while keeping the Radio Parameters. |

### <a name="ttn.lorawan.v3.RejoinTimeExponent">Enum `RejoinTimeExponent`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `REJOIN_TIME_0` | 0 | Every ~17.1 minutes. |
| `REJOIN_TIME_1` | 1 | Every ~34.1 minutes. |
| `REJOIN_TIME_2` | 2 | Every ~1.1 hours. |
| `REJOIN_TIME_3` | 3 | Every ~2.3 hours. |
| `REJOIN_TIME_4` | 4 | Every ~4.6 hours. |
| `REJOIN_TIME_5` | 5 | Every ~9.1 hours. |
| `REJOIN_TIME_6` | 6 | Every ~18.2 hours. |
| `REJOIN_TIME_7` | 7 | Every ~1.5 days. |
| `REJOIN_TIME_8` | 8 | Every ~3.0 days. |
| `REJOIN_TIME_9` | 9 | Every ~6.1 days. |
| `REJOIN_TIME_10` | 10 | Every ~12.1 days. |
| `REJOIN_TIME_11` | 11 | Every ~3.5 weeks. |
| `REJOIN_TIME_12` | 12 | Every ~1.6 months. |
| `REJOIN_TIME_13` | 13 | Every ~3.2 months. |
| `REJOIN_TIME_14` | 14 | Every ~6.4 months. |
| `REJOIN_TIME_15` | 15 | Every ~1.1 year. |

### <a name="ttn.lorawan.v3.RelayCADPeriodicity">Enum `RelayCADPeriodicity`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RELAY_CAD_PERIODICITY_1_SECOND` | 0 |  |
| `RELAY_CAD_PERIODICITY_500_MILLISECONDS` | 1 |  |
| `RELAY_CAD_PERIODICITY_250_MILLISECONDS` | 2 |  |
| `RELAY_CAD_PERIODICITY_100_MILLISECONDS` | 3 |  |
| `RELAY_CAD_PERIODICITY_50_MILLISECONDS` | 4 |  |
| `RELAY_CAD_PERIODICITY_20_MILLISECONDS` | 5 | sic |

### <a name="ttn.lorawan.v3.RelayCtrlUplinkListAction">Enum `RelayCtrlUplinkListAction`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RELAY_CTRL_UPLINK_LIST_ACTION_READ_W_F_CNT` | 0 |  |
| `RELAY_CTRL_UPLINK_LIST_ACTION_REMOVE_TRUSTED_END_DEVICE` | 1 |  |

### <a name="ttn.lorawan.v3.RelayLimitBucketSize">Enum `RelayLimitBucketSize`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RELAY_LIMIT_BUCKET_SIZE_1` | 0 |  |
| `RELAY_LIMIT_BUCKET_SIZE_2` | 1 |  |
| `RELAY_LIMIT_BUCKET_SIZE_4` | 2 |  |
| `RELAY_LIMIT_BUCKET_SIZE_12` | 3 | sic |

### <a name="ttn.lorawan.v3.RelayResetLimitCounter">Enum `RelayResetLimitCounter`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RELAY_RESET_LIMIT_COUNTER_ZERO` | 0 |  |
| `RELAY_RESET_LIMIT_COUNTER_RELOAD_RATE` | 1 |  |
| `RELAY_RESET_LIMIT_COUNTER_MAX_VALUE` | 2 |  |
| `RELAY_RESET_LIMIT_COUNTER_NO_RESET` | 3 |  |

### <a name="ttn.lorawan.v3.RelaySecondChAckOffset">Enum `RelaySecondChAckOffset`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RELAY_SECOND_CH_ACK_OFFSET_0` | 0 | 0 kHz |
| `RELAY_SECOND_CH_ACK_OFFSET_200` | 1 | 200 kHz |
| `RELAY_SECOND_CH_ACK_OFFSET_400` | 2 | 400 kHz |
| `RELAY_SECOND_CH_ACK_OFFSET_800` | 3 | 800 kHz |
| `RELAY_SECOND_CH_ACK_OFFSET_1600` | 4 | 1.6 MHz |
| `RELAY_SECOND_CH_ACK_OFFSET_3200` | 5 | 3.2 MHz |

### <a name="ttn.lorawan.v3.RelaySmartEnableLevel">Enum `RelaySmartEnableLevel`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RELAY_SMART_ENABLE_LEVEL_8` | 0 |  |
| `RELAY_SMART_ENABLE_LEVEL_16` | 1 |  |
| `RELAY_SMART_ENABLE_LEVEL_32` | 2 |  |
| `RELAY_SMART_ENABLE_LEVEL_64` | 3 |  |

### <a name="ttn.lorawan.v3.RelayWORChannel">Enum `RelayWORChannel`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RELAY_WOR_CHANNEL_DEFAULT` | 0 |  |
| `RELAY_WOR_CHANNEL_SECONDARY` | 1 |  |

### <a name="ttn.lorawan.v3.RxDelay">Enum `RxDelay`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RX_DELAY_0` | 0 | 1 second. |
| `RX_DELAY_1` | 1 | 1 second. |
| `RX_DELAY_2` | 2 | 2 seconds. |
| `RX_DELAY_3` | 3 | 3 seconds. |
| `RX_DELAY_4` | 4 | 4 seconds. |
| `RX_DELAY_5` | 5 | 5 seconds. |
| `RX_DELAY_6` | 6 | 6 seconds. |
| `RX_DELAY_7` | 7 | 7 seconds. |
| `RX_DELAY_8` | 8 | 8 seconds. |
| `RX_DELAY_9` | 9 | 9 seconds. |
| `RX_DELAY_10` | 10 | 10 seconds. |
| `RX_DELAY_11` | 11 | 11 seconds. |
| `RX_DELAY_12` | 12 | 12 seconds. |
| `RX_DELAY_13` | 13 | 13 seconds. |
| `RX_DELAY_14` | 14 | 14 seconds. |
| `RX_DELAY_15` | 15 | 15 seconds. |

### <a name="ttn.lorawan.v3.TxSchedulePriority">Enum `TxSchedulePriority`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `LOWEST` | 0 |  |
| `LOW` | 1 |  |
| `BELOW_NORMAL` | 2 |  |
| `NORMAL` | 3 |  |
| `ABOVE_NORMAL` | 4 |  |
| `HIGH` | 5 |  |
| `HIGHEST` | 6 |  |

## <a name="ttn/lorawan/v3/messages.proto">File `ttn/lorawan/v3/messages.proto`</a>

### <a name="ttn.lorawan.v3.ApplicationDownlink">Message `ApplicationDownlink`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_key_id` | [`bytes`](#bytes) |  | Join Server issued identifier for the session keys used by this downlink. |
| `f_port` | [`uint32`](#uint32) |  |  |
| `f_cnt` | [`uint32`](#uint32) |  |  |
| `frm_payload` | [`bytes`](#bytes) |  | The frame payload of the downlink message. The payload is encrypted if the skip_payload_crypto field of the EndDevice is true. |
| `decoded_payload` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | The decoded frame payload of the downlink message. When scheduling downlink with a message processor configured for the end device (see formatters) or application (see default_formatters), this fields acts as input for the downlink encoder, and the output is set to frm_payload. When reading downlink (listing the queue, downlink message events, etc), this fields acts as output of the downlink decoder, and the input is frm_payload. |
| `decoded_payload_warnings` | [`string`](#string) | repeated | Warnings generated by the message processor while encoding frm_payload (scheduling downlink) or decoding the frm_payload (reading downlink). |
| `confirmed` | [`bool`](#bool) |  |  |
| `class_b_c` | [`ApplicationDownlink.ClassBC`](#ttn.lorawan.v3.ApplicationDownlink.ClassBC) |  | Optional gateway and timing information for class B and C. If set, this downlink message will only be transmitted as class B or C downlink. If not set, this downlink message may be transmitted in class A, B and C. |
| `priority` | [`TxSchedulePriority`](#ttn.lorawan.v3.TxSchedulePriority) |  | Priority for scheduling the downlink message. |
| `correlation_ids` | [`string`](#string) | repeated |  |
| `confirmed_retry` | [`ApplicationDownlink.ConfirmedRetry`](#ttn.lorawan.v3.ApplicationDownlink.ConfirmedRetry) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |
| `f_port` | <p>`uint32.lte`: `255`</p><p>`uint32.not_in`: `[224]`</p> |
| `priority` | <p>`enum.defined_only`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ApplicationDownlink.ClassBC">Message `ApplicationDownlink.ClassBC`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateways` | [`ClassBCGatewayIdentifiers`](#ttn.lorawan.v3.ClassBCGatewayIdentifiers) | repeated | Possible gateway identifiers, antenna index, and group index to use for this downlink message. The Network Server selects one of these gateways for downlink, based on connectivity, signal quality, channel utilization and an available slot. If none of the gateways can be selected, the downlink message fails. If empty, a gateway and antenna is selected automatically from the gateways seen in recent uplinks. If group index is set, gateways will be grouped by the index for the Network Server to select one gateway per group. |
| `absolute_time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Absolute time when the downlink message should be transmitted. This requires the gateway to have GPS time synchronization. If the time is in the past or if there is a scheduling conflict, the downlink message fails. If null, the time is selected based on slot availability. This is recommended in class B mode. |

### <a name="ttn.lorawan.v3.ApplicationDownlink.ConfirmedRetry">Message `ApplicationDownlink.ConfirmedRetry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `attempt` | [`uint32`](#uint32) |  | The number of attempted confirmed downlink acknowledgements. |
| `max_attempts` | [`google.protobuf.UInt32Value`](#google.protobuf.UInt32Value) |  | The maximum number of confirmed downlink acknowledgement attempts. If null, the Application Server configuration is used instead. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `max_attempts` | <p>`uint32.lte`: `100`</p><p>`uint32.gt`: `0`</p> |

### <a name="ttn.lorawan.v3.ApplicationDownlinkFailed">Message `ApplicationDownlinkFailed`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `downlink` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| `error` | [`ErrorDetails`](#ttn.lorawan.v3.ErrorDetails) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `downlink` | <p>`message.required`: `true`</p> |
| `error` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationDownlinks">Message `ApplicationDownlinks`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `downlinks` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  |

### <a name="ttn.lorawan.v3.ApplicationInvalidatedDownlinks">Message `ApplicationInvalidatedDownlinks`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `downlinks` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  |
| `last_f_cnt_down` | [`uint32`](#uint32) |  |  |
| `session_key_id` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |

### <a name="ttn.lorawan.v3.ApplicationJoinAccept">Message `ApplicationJoinAccept`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_key_id` | [`bytes`](#bytes) |  | Join Server issued identifier for the session keys negotiated in this join. |
| `app_s_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | Encrypted Application Session Key (if Join Server sent it to Network Server). |
| `invalidated_downlinks` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) | repeated | Downlink messages in the queue that got invalidated because of the session change. |
| `pending_session` | [`bool`](#bool) |  | Indicates whether the security context refers to the pending session, i.e. when this join-accept is an answer to a rejoin-request. |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Server time when the Network Server received the message. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |
| `received_at` | <p>`timestamp.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationLocation">Message `ApplicationLocation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `service` | [`string`](#string) |  |  |
| `location` | [`Location`](#ttn.lorawan.v3.Location) |  |  |
| `attributes` | [`ApplicationLocation.AttributesEntry`](#ttn.lorawan.v3.ApplicationLocation.AttributesEntry) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `location` | <p>`message.required`: `true`</p> |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |

### <a name="ttn.lorawan.v3.ApplicationLocation.AttributesEntry">Message `ApplicationLocation.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.ApplicationServiceData">Message `ApplicationServiceData`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `service` | [`string`](#string) |  |  |
| `data` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  |  |

### <a name="ttn.lorawan.v3.ApplicationUp">Message `ApplicationUp`</a>

Application uplink message.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `correlation_ids` | [`string`](#string) | repeated |  |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Server time when the Application Server received the message. |
| `uplink_message` | [`ApplicationUplink`](#ttn.lorawan.v3.ApplicationUplink) |  |  |
| `uplink_normalized` | [`ApplicationUplinkNormalized`](#ttn.lorawan.v3.ApplicationUplinkNormalized) |  |  |
| `join_accept` | [`ApplicationJoinAccept`](#ttn.lorawan.v3.ApplicationJoinAccept) |  |  |
| `downlink_ack` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| `downlink_nack` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| `downlink_sent` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| `downlink_failed` | [`ApplicationDownlinkFailed`](#ttn.lorawan.v3.ApplicationDownlinkFailed) |  |  |
| `downlink_queued` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) |  |  |
| `downlink_queue_invalidated` | [`ApplicationInvalidatedDownlinks`](#ttn.lorawan.v3.ApplicationInvalidatedDownlinks) |  |  |
| `location_solved` | [`ApplicationLocation`](#ttn.lorawan.v3.ApplicationLocation) |  |  |
| `service_data` | [`ApplicationServiceData`](#ttn.lorawan.v3.ApplicationServiceData) |  |  |
| `simulated` | [`bool`](#bool) |  | Signals if the message is coming from the Network Server or is simulated. The Application Server automatically sets this field, and callers must not manually set it. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ApplicationUplink">Message `ApplicationUplink`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_key_id` | [`bytes`](#bytes) |  | Join Server issued identifier for the session keys used by this uplink. |
| `f_port` | [`uint32`](#uint32) |  | LoRaWAN FPort of the uplink message. |
| `f_cnt` | [`uint32`](#uint32) |  | LoRaWAN FCntUp of the uplink message. |
| `frm_payload` | [`bytes`](#bytes) |  | The frame payload of the uplink message. The payload is still encrypted if the skip_payload_crypto field of the EndDevice is true, which is indicated by the presence of the app_s_key field. |
| `decoded_payload` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | The decoded frame payload of the uplink message. This field is set by the message processor that is configured for the end device (see formatters) or application (see default_formatters). |
| `decoded_payload_warnings` | [`string`](#string) | repeated | Warnings generated by the message processor while decoding the frm_payload. |
| `normalized_payload` | [`google.protobuf.Struct`](#google.protobuf.Struct) | repeated | The normalized frame payload of the uplink message. This field is set by the message processor that is configured for the end device (see formatters) or application (see default_formatters). If the message processor is a custom script, there is no uplink normalizer script and the decoded output is valid normalized payload, this field contains the decoded payload. |
| `normalized_payload_warnings` | [`string`](#string) | repeated | Warnings generated by the message processor while normalizing the decoded payload. |
| `rx_metadata` | [`RxMetadata`](#ttn.lorawan.v3.RxMetadata) | repeated | A list of metadata for each antenna of each gateway that received this message. |
| `settings` | [`TxSettings`](#ttn.lorawan.v3.TxSettings) |  | Transmission settings used by the end device. |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Server time when the Network Server received the message. |
| `app_s_key` | [`KeyEnvelope`](#ttn.lorawan.v3.KeyEnvelope) |  | The AppSKey of the current session. This field is only present if the skip_payload_crypto field of the EndDevice is true. Can be used to decrypt uplink payloads and encrypt downlink payloads. |
| `last_a_f_cnt_down` | [`uint32`](#uint32) |  | The last AFCntDown of the current session. This field is only present if the skip_payload_crypto field of the EndDevice is true. Can be used with app_s_key to encrypt downlink payloads. |
| `confirmed` | [`bool`](#bool) |  | Indicates whether the end device used confirmed data uplink. |
| `consumed_airtime` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Consumed airtime for the transmission of the uplink message. Calculated by Network Server using the raw payload size and the transmission settings. |
| `locations` | [`ApplicationUplink.LocationsEntry`](#ttn.lorawan.v3.ApplicationUplink.LocationsEntry) | repeated | End device location metadata, set by the Application Server while handling the message. |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | End device version identifiers, set by the Application Server while handling the message. |
| `network_ids` | [`NetworkIdentifiers`](#ttn.lorawan.v3.NetworkIdentifiers) |  | Network identifiers, set by the Network Server that handles the message. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |
| `f_port` | <p>`uint32.lte`: `255`</p><p>`uint32.not_in`: `[224]`</p> |
| `settings` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationUplink.LocationsEntry">Message `ApplicationUplink.LocationsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`Location`](#ttn.lorawan.v3.Location) |  |  |

### <a name="ttn.lorawan.v3.ApplicationUplinkNormalized">Message `ApplicationUplinkNormalized`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_key_id` | [`bytes`](#bytes) |  | Join Server issued identifier for the session keys used by this uplink. |
| `f_port` | [`uint32`](#uint32) |  | LoRaWAN FPort of the uplink message. |
| `f_cnt` | [`uint32`](#uint32) |  | LoRaWAN FCntUp of the uplink message. |
| `frm_payload` | [`bytes`](#bytes) |  | The frame payload of the uplink message. This field is always decrypted with AppSKey. |
| `normalized_payload` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | The normalized frame payload of the uplink message. This field is set for each item in normalized_payload in the corresponding ApplicationUplink message. |
| `normalized_payload_warnings` | [`string`](#string) | repeated | This field is set to normalized_payload_warnings in the corresponding ApplicationUplink message. |
| `rx_metadata` | [`RxMetadata`](#ttn.lorawan.v3.RxMetadata) | repeated | A list of metadata for each antenna of each gateway that received this message. |
| `settings` | [`TxSettings`](#ttn.lorawan.v3.TxSettings) |  | Transmission settings used by the end device. |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Server time when the Network Server received the message. |
| `confirmed` | [`bool`](#bool) |  | Indicates whether the end device used confirmed data uplink. |
| `consumed_airtime` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Consumed airtime for the transmission of the uplink message. Calculated by Network Server using the raw payload size and the transmission settings. |
| `locations` | [`ApplicationUplinkNormalized.LocationsEntry`](#ttn.lorawan.v3.ApplicationUplinkNormalized.LocationsEntry) | repeated | End device location metadata, set by the Application Server while handling the message. |
| `version_ids` | [`EndDeviceVersionIdentifiers`](#ttn.lorawan.v3.EndDeviceVersionIdentifiers) |  | End device version identifiers, set by the Application Server while handling the message. |
| `network_ids` | [`NetworkIdentifiers`](#ttn.lorawan.v3.NetworkIdentifiers) |  | Network identifiers, set by the Network Server that handles the message. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |
| `f_port` | <p>`uint32.lte`: `255`</p><p>`uint32.gte`: `1`</p><p>`uint32.not_in`: `[224]`</p> |
| `normalized_payload` | <p>`message.required`: `true`</p> |
| `settings` | <p>`message.required`: `true`</p> |
| `received_at` | <p>`timestamp.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ApplicationUplinkNormalized.LocationsEntry">Message `ApplicationUplinkNormalized.LocationsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`Location`](#ttn.lorawan.v3.Location) |  |  |

### <a name="ttn.lorawan.v3.DownlinkMessage">Message `DownlinkMessage`</a>

Downlink message from the network to the end device

Mapping from UDP message:

imme: -
tmst: scheduled.timestamp
tmms: scheduled.time
freq: scheduled.frequency
rfch: (0)
powe: scheduled.tx_power
modu: scheduled.modulation
datr: scheduled.data_rate_index (derived)
codr: scheduled.coding_rate
fdev: (derived from bandwidth)
ipol: scheduled.invert_polarization
prea: [scheduled.advanced]
size: (derived from len(raw_payload))
data: raw_payload
ncrc: [scheduled.advanced]

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_payload` | [`bytes`](#bytes) |  |  |
| `payload` | [`Message`](#ttn.lorawan.v3.Message) |  |  |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `request` | [`TxRequest`](#ttn.lorawan.v3.TxRequest) |  |  |
| `scheduled` | [`TxSettings`](#ttn.lorawan.v3.TxSettings) |  |  |
| `correlation_ids` | [`string`](#string) | repeated |  |
| `session_key_id` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |

### <a name="ttn.lorawan.v3.DownlinkQueueOperationErrorDetails">Message `DownlinkQueueOperationErrorDetails`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `session_key_id` | [`bytes`](#bytes) |  |  |
| `min_f_cnt_down` | [`uint32`](#uint32) |  |  |
| `pending_dev_addr` | [`bytes`](#bytes) |  |  |
| `pending_session_key_id` | [`bytes`](#bytes) |  |  |
| `pending_min_f_cnt_down` | [`uint32`](#uint32) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `session_key_id` | <p>`bytes.max_len`: `2048`</p> |
| `pending_dev_addr` | <p>`bytes.len`: `4`</p> |
| `pending_session_key_id` | <p>`bytes.max_len`: `2048`</p> |

### <a name="ttn.lorawan.v3.DownlinkQueueRequest">Message `DownlinkQueueRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `end_device_ids` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) |  |  |
| `downlinks` | [`ApplicationDownlink`](#ttn.lorawan.v3.ApplicationDownlink) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `end_device_ids` | <p>`message.required`: `true`</p> |
| `downlinks` | <p>`repeated.max_items`: `100000`</p> |

### <a name="ttn.lorawan.v3.GatewayTxAcknowledgment">Message `GatewayTxAcknowledgment`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `tx_ack` | [`TxAcknowledgment`](#ttn.lorawan.v3.TxAcknowledgment) |  |  |

### <a name="ttn.lorawan.v3.GatewayUplinkMessage">Message `GatewayUplinkMessage`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `message` | [`UplinkMessage`](#ttn.lorawan.v3.UplinkMessage) |  |  |
| `band_id` | [`string`](#string) |  | LoRaWAN band ID of the gateway. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `message` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.MessagePayloadFormatters">Message `MessagePayloadFormatters`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `up_formatter` | [`PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter for uplink messages, must be set together with its parameter. |
| `up_formatter_parameter` | [`string`](#string) |  | Parameter for the up_formatter, must be set together. The API enforces a maximum length of 16KB, but the size may be restricted further by deployment configuration. |
| `down_formatter` | [`PayloadFormatter`](#ttn.lorawan.v3.PayloadFormatter) |  | Payload formatter for downlink messages, must be set together with its parameter. |
| `down_formatter_parameter` | [`string`](#string) |  | Parameter for the down_formatter, must be set together. The API enforces a maximum length of 16KB, but the size may be restricted further by deployment configuration. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `up_formatter` | <p>`enum.defined_only`: `true`</p> |
| `up_formatter_parameter` | <p>`string.max_len`: `40960`</p> |
| `down_formatter` | <p>`enum.defined_only`: `true`</p> |
| `down_formatter_parameter` | <p>`string.max_len`: `40960`</p> |

### <a name="ttn.lorawan.v3.TxAcknowledgment">Message `TxAcknowledgment`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `correlation_ids` | [`string`](#string) | repeated | Correlation IDs for the downlink message. Set automatically by the UDP and LBS frontends. For gRPC and the MQTT v3 frontends, the correlation IDs must match the ones of the downlink message the Tx acknowledgment message refers to. |
| `result` | [`TxAcknowledgment.Result`](#ttn.lorawan.v3.TxAcknowledgment.Result) |  |  |
| `downlink_message` | [`DownlinkMessage`](#ttn.lorawan.v3.DownlinkMessage) |  | The acknowledged downlink message. Set by the Gateway Server. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |
| `result` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.UplinkMessage">Message `UplinkMessage`</a>

Uplink message from the end device to the network

Mapping from UDP message (other fields can be set in "advanced"):

- time: rx_metadata.time
- tmst: rx_metadata.timestamp
- freq: settings.frequency
- modu: settings.modulation
- datr: settings.data_rate_index (and derived fields)
- codr: settings.coding_rate
- size: len(raw_payload)
- data: raw_payload (and derived payload)
- rsig: rx_metadata
 - ant: rx_metadata.antenna_index
 - chan: rx_metadata.channel_index
 - rssis: rx_metadata.rssi
 - lsnr: rx_metadata.snr

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_payload` | [`bytes`](#bytes) |  |  |
| `payload` | [`Message`](#ttn.lorawan.v3.Message) |  |  |
| `settings` | [`TxSettings`](#ttn.lorawan.v3.TxSettings) |  |  |
| `rx_metadata` | [`RxMetadata`](#ttn.lorawan.v3.RxMetadata) | repeated |  |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Server time when a component received the message. The Gateway Server and Network Server set this value to their local server time of reception. |
| `correlation_ids` | [`string`](#string) | repeated |  |
| `device_channel_index` | [`uint32`](#uint32) |  | Index of the device channel that received the message. Set by Network Server. |
| `consumed_airtime` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Consumed airtime for the transmission of the uplink message. Calculated by Network Server using the RawPayload size and the transmission settings. |
| `crc_status` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  | Cyclic Redundancy Check (CRC) status of demodulating the frame. If unset, the modulation does not support CRC or the gateway did not provide a CRC status. If set to false, this message should not be processed. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `settings` | <p>`message.required`: `true`</p> |
| `correlation_ids` | <p>`repeated.items.string.max_len`: `100`</p> |
| `device_channel_index` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.PayloadFormatter">Enum `PayloadFormatter`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `FORMATTER_NONE` | 0 | No payload formatter to work with raw payload only. |
| `FORMATTER_REPOSITORY` | 1 | Use payload formatter for the end device type from a repository. |
| `FORMATTER_GRPC_SERVICE` | 2 | gRPC service payload formatter. The parameter is the host:port of the service. |
| `FORMATTER_JAVASCRIPT` | 3 | Custom payload formatter that executes Javascript code. The parameter is a JavaScript filename. |
| `FORMATTER_CAYENNELPP` | 4 | CayenneLPP payload formatter. More payload formatters can be added. |

### <a name="ttn.lorawan.v3.TxAcknowledgment.Result">Enum `TxAcknowledgment.Result`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `SUCCESS` | 0 |  |
| `UNKNOWN_ERROR` | 1 |  |
| `TOO_LATE` | 2 |  |
| `TOO_EARLY` | 3 |  |
| `COLLISION_PACKET` | 4 |  |
| `COLLISION_BEACON` | 5 |  |
| `TX_FREQ` | 6 |  |
| `TX_POWER` | 7 |  |
| `GPS_UNLOCKED` | 8 |  |

## <a name="ttn/lorawan/v3/metadata.proto">File `ttn/lorawan/v3/metadata.proto`</a>

### <a name="ttn.lorawan.v3.Location">Message `Location`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `latitude` | [`double`](#double) |  | The North–South position (degrees; -90 to +90), where 0 is the equator, North pole is positive, South pole is negative. |
| `longitude` | [`double`](#double) |  | The East-West position (degrees; -180 to +180), where 0 is the Prime Meridian (Greenwich), East is positive , West is negative. |
| `altitude` | [`int32`](#int32) |  | The altitude (meters), where 0 is the mean sea level. |
| `accuracy` | [`int32`](#int32) |  | The accuracy of the location (meters). |
| `source` | [`LocationSource`](#ttn.lorawan.v3.LocationSource) |  | Source of the location information. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `latitude` | <p>`double.lte`: `90`</p><p>`double.gte`: `-90`</p> |
| `longitude` | <p>`double.lte`: `180`</p><p>`double.gte`: `-180`</p> |
| `source` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.PacketBrokerMetadata">Message `PacketBrokerMetadata`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `message_id` | [`string`](#string) |  | Message identifier generated by Packet Broker Router. |
| `forwarder_net_id` | [`bytes`](#bytes) |  | LoRa Alliance NetID of the Packet Broker Forwarder Member. |
| `forwarder_tenant_id` | [`string`](#string) |  | Tenant ID managed by the Packet Broker Forwarder Member. |
| `forwarder_cluster_id` | [`string`](#string) |  | Forwarder Cluster ID of the Packet Broker Forwarder. |
| `forwarder_gateway_eui` | [`bytes`](#bytes) |  | Forwarder gateway EUI. |
| `forwarder_gateway_id` | [`google.protobuf.StringValue`](#google.protobuf.StringValue) |  | Forwarder gateway ID. |
| `home_network_net_id` | [`bytes`](#bytes) |  | LoRa Alliance NetID of the Packet Broker Home Network Member. |
| `home_network_tenant_id` | [`string`](#string) |  | Tenant ID managed by the Packet Broker Home Network Member. This value is empty if it cannot be determined by the Packet Broker Router. |
| `home_network_cluster_id` | [`string`](#string) |  | Home Network Cluster ID of the Packet Broker Home Network. |
| `hops` | [`PacketBrokerRouteHop`](#ttn.lorawan.v3.PacketBrokerRouteHop) | repeated | Hops that the message passed. Each Packet Broker Router service appends an entry. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `forwarder_net_id` | <p>`bytes.len`: `3`</p> |
| `forwarder_gateway_eui` | <p>`bytes.len`: `8`</p> |
| `home_network_net_id` | <p>`bytes.len`: `3`</p> |

### <a name="ttn.lorawan.v3.PacketBrokerRouteHop">Message `PacketBrokerRouteHop`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Time when the service received the message. |
| `sender_name` | [`string`](#string) |  | Sender of the message, typically the authorized client identifier. |
| `sender_address` | [`string`](#string) |  | Sender IP address or host name. |
| `receiver_name` | [`string`](#string) |  | Receiver of the message. |
| `receiver_agent` | [`string`](#string) |  | Receiver agent. |

### <a name="ttn.lorawan.v3.RelayMetadata">Message `RelayMetadata`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `device_id` | [`string`](#string) |  | End device identifiers of the relay. |
| `wor_channel` | [`RelayWORChannel`](#ttn.lorawan.v3.RelayWORChannel) |  | Wake on radio channel. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `device_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `wor_channel` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.RxMetadata">Message `RxMetadata`</a>

Contains metadata for a received message. Each antenna that receives
a message corresponds to one RxMetadata.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `packet_broker` | [`PacketBrokerMetadata`](#ttn.lorawan.v3.PacketBrokerMetadata) |  |  |
| `relay` | [`RelayMetadata`](#ttn.lorawan.v3.RelayMetadata) |  |  |
| `antenna_index` | [`uint32`](#uint32) |  |  |
| `time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp at the end of the transmission, provided by the gateway. The accuracy is undefined. |
| `timestamp` | [`uint32`](#uint32) |  | Gateway concentrator timestamp when the Rx finished (microseconds). |
| `fine_timestamp` | [`uint64`](#uint64) |  | Gateway's internal fine timestamp when the Rx finished (nanoseconds). |
| `encrypted_fine_timestamp` | [`bytes`](#bytes) |  | Encrypted gateway's internal fine timestamp when the Rx finished (nanoseconds). |
| `encrypted_fine_timestamp_key_id` | [`string`](#string) |  |  |
| `rssi` | [`float`](#float) |  | Received signal strength indicator (dBm). This value equals `channel_rssi`. |
| `signal_rssi` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Received signal strength indicator of the signal (dBm). |
| `channel_rssi` | [`float`](#float) |  | Received signal strength indicator of the channel (dBm). |
| `rssi_standard_deviation` | [`float`](#float) |  | Standard deviation of the RSSI during preamble. |
| `snr` | [`float`](#float) |  | Signal-to-noise ratio (dB). |
| `frequency_offset` | [`int64`](#int64) |  | Frequency offset (Hz). |
| `location` | [`Location`](#ttn.lorawan.v3.Location) |  | Antenna location; injected by the Gateway Server. |
| `downlink_path_constraint` | [`DownlinkPathConstraint`](#ttn.lorawan.v3.DownlinkPathConstraint) |  | Gateway downlink path constraint; injected by the Gateway Server. |
| `uplink_token` | [`bytes`](#bytes) |  | Uplink token to be included in the Tx request in class A downlink; injected by gateway, Gateway Server or fNS. |
| `channel_index` | [`uint32`](#uint32) |  | Index of the gateway channel that received the message. |
| `hopping_width` | [`uint32`](#uint32) |  | Hopping width; a number describing the number of steps of the LR-FHSS grid. |
| `frequency_drift` | [`int32`](#int32) |  | Frequency drift in Hz between start and end of an LR-FHSS packet (signed). |
| `gps_time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp at the end of the transmission, provided by the gateway. Guaranteed to be based on a GPS PPS signal, with an accuracy of 1 millisecond. |
| `received_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp at which the Gateway Server has received the message. |
| `advanced` | [`google.protobuf.Struct`](#google.protobuf.Struct) |  | Advanced metadata fields - can be used for advanced information or experimental features that are not yet formally defined in the API - field names are written in snake_case |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_ids` | <p>`message.required`: `true`</p> |
| `downlink_path_constraint` | <p>`enum.defined_only`: `true`</p> |
| `channel_index` | <p>`uint32.lte`: `255`</p> |

### <a name="ttn.lorawan.v3.LocationSource">Enum `LocationSource`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `SOURCE_UNKNOWN` | 0 | The source of the location is not known or not set. |
| `SOURCE_GPS` | 1 | The location is determined by GPS. |
| `SOURCE_REGISTRY` | 3 | The location is set in and updated from a registry. |
| `SOURCE_IP_GEOLOCATION` | 4 | The location is estimated with IP geolocation. |
| `SOURCE_WIFI_RSSI_GEOLOCATION` | 5 | The location is estimated with WiFi RSSI geolocation. |
| `SOURCE_BT_RSSI_GEOLOCATION` | 6 | The location is estimated with BT/BLE RSSI geolocation. |
| `SOURCE_LORA_RSSI_GEOLOCATION` | 7 | The location is estimated with LoRa RSSI geolocation. |
| `SOURCE_LORA_TDOA_GEOLOCATION` | 8 | The location is estimated with LoRa TDOA geolocation. |
| `SOURCE_COMBINED_GEOLOCATION` | 9 | The location is estimated by a combination of geolocation sources. More estimation methods can be added. |

## <a name="ttn/lorawan/v3/mqtt.proto">File `ttn/lorawan/v3/mqtt.proto`</a>

### <a name="ttn.lorawan.v3.MQTTConnectionInfo">Message `MQTTConnectionInfo`</a>

The connection information of an MQTT frontend.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `public_address` | [`string`](#string) |  | The public listen address of the frontend. |
| `public_tls_address` | [`string`](#string) |  | The public listen address of the TLS frontend. |
| `username` | [`string`](#string) |  | The username to be used for authentication. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `public_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |
| `public_tls_address` | <p>`string.pattern`: `^(?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*(?:[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(?::[0-9]{1,5})?$|^$`</p> |

## <a name="ttn/lorawan/v3/networkserver.proto">File `ttn/lorawan/v3/networkserver.proto`</a>

### <a name="ttn.lorawan.v3.GenerateDevAddrResponse">Message `GenerateDevAddrResponse`</a>

Response of GenerateDevAddr.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr` | <p>`bytes.len`: `4`</p> |

### <a name="ttn.lorawan.v3.GetDefaultMACSettingsRequest">Message `GetDefaultMACSettingsRequest`</a>

Request of GetDefaultMACSettings.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency_plan_id` | [`string`](#string) |  |  |
| `lorawan_phy_version` | [`PHYVersion`](#ttn.lorawan.v3.PHYVersion) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `frequency_plan_id` | <p>`string.max_len`: `64`</p> |
| `lorawan_phy_version` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.GetDeviceAdressPrefixesResponse">Message `GetDeviceAdressPrefixesResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr_prefixes` | [`bytes`](#bytes) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr_prefixes` | <p>`repeated.items.bytes.len`: `5`</p> |

### <a name="ttn.lorawan.v3.GetNetIDResponse">Message `GetNetIDResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `net_id` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `net_id` | <p>`bytes.len`: `3`</p> |

### <a name="ttn.lorawan.v3.AsNs">Service `AsNs`</a>

The AsNs service connects an Application Server to a Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `DownlinkQueueReplace` | [`DownlinkQueueRequest`](#ttn.lorawan.v3.DownlinkQueueRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Replace the entire downlink queue with the specified messages. This can also be used to empty the queue by specifying no messages. Note that this will trigger an immediate downlink if a downlink slot is available. |
| `DownlinkQueuePush` | [`DownlinkQueueRequest`](#ttn.lorawan.v3.DownlinkQueueRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Push downlink messages to the end of the downlink queue. Note that this will trigger an immediate downlink if a downlink slot is available. |
| `DownlinkQueueList` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`ApplicationDownlinks`](#ttn.lorawan.v3.ApplicationDownlinks) | List the items currently in the downlink queue. |

### <a name="ttn.lorawan.v3.GsNs">Service `GsNs`</a>

The GsNs service connects a Gateway Server to a Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `HandleUplink` | [`UplinkMessage`](#ttn.lorawan.v3.UplinkMessage) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Called by the Gateway Server when an uplink message arrives. |
| `ReportTxAcknowledgment` | [`GatewayTxAcknowledgment`](#ttn.lorawan.v3.GatewayTxAcknowledgment) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Called by the Gateway Server when a Tx acknowledgment arrives. |

### <a name="ttn.lorawan.v3.Ns">Service `Ns`</a>

The Ns service manages the Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GenerateDevAddr` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`GenerateDevAddrResponse`](#ttn.lorawan.v3.GenerateDevAddrResponse) | GenerateDevAddr requests a device address assignment from the Network Server. |
| `GetDefaultMACSettings` | [`GetDefaultMACSettingsRequest`](#ttn.lorawan.v3.GetDefaultMACSettingsRequest) | [`MACSettings`](#ttn.lorawan.v3.MACSettings) | GetDefaultMACSettings retrieves the default MAC settings for a frequency plan. |
| `GetNetID` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`GetNetIDResponse`](#ttn.lorawan.v3.GetNetIDResponse) |  |
| `GetDeviceAddressPrefixes` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`GetDeviceAdressPrefixesResponse`](#ttn.lorawan.v3.GetDeviceAdressPrefixesResponse) |  |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GenerateDevAddr` | `GET` | `/api/v3/ns/dev_addr` |  |
| `GetDefaultMACSettings` | `GET` | `/api/v3/ns/default_mac_settings/{frequency_plan_id}/{lorawan_phy_version}` |  |
| `GetNetID` | `GET` | `/api/v3/ns/net_id` |  |
| `GetDeviceAddressPrefixes` | `GET` | `/api/v3/ns/dev_addr_prefixes` |  |

### <a name="ttn.lorawan.v3.NsEndDeviceBatchRegistry">Service `NsEndDeviceBatchRegistry`</a>

The NsEndDeviceBatchRegistry service allows clients to manage batches of end devices on the Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Delete` | [`BatchDeleteEndDevicesRequest`](#ttn.lorawan.v3.BatchDeleteEndDevicesRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a list of devices within the same application. This operation is atomic; either all devices are deleted or none. Devices not found are skipped and no error is returned. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Delete` | `DELETE` | `/api/v3/ns/applications/{application_ids.application_id}/devices/batch` |  |

### <a name="ttn.lorawan.v3.NsEndDeviceRegistry">Service `NsEndDeviceRegistry`</a>

The NsEndDeviceRegistry service allows clients to manage their end devices on the Network Server.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Get` | [`GetEndDeviceRequest`](#ttn.lorawan.v3.GetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Get returns the device that matches the given identifiers. If there are multiple matches, an error will be returned. |
| `Set` | [`SetEndDeviceRequest`](#ttn.lorawan.v3.SetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | Set creates or updates the device. |
| `ResetFactoryDefaults` | [`ResetAndGetEndDeviceRequest`](#ttn.lorawan.v3.ResetAndGetEndDeviceRequest) | [`EndDevice`](#ttn.lorawan.v3.EndDevice) | ResetFactoryDefaults resets device state to factory defaults. |
| `Delete` | [`EndDeviceIdentifiers`](#ttn.lorawan.v3.EndDeviceIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete deletes the device that matches the given identifiers. If there are multiple matches, an error will be returned. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Get` | `GET` | `/api/v3/ns/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` |  |
| `Set` | `PUT` | `/api/v3/ns/applications/{end_device.ids.application_ids.application_id}/devices/{end_device.ids.device_id}` | `*` |
| `Set` | `POST` | `/api/v3/ns/applications/{end_device.ids.application_ids.application_id}/devices` | `*` |
| `ResetFactoryDefaults` | `PATCH` | `/api/v3/ns/applications/{end_device_ids.application_ids.application_id}/devices/{end_device_ids.device_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/ns/applications/{application_ids.application_id}/devices/{device_id}` |  |

## <a name="ttn/lorawan/v3/notification_service.proto">File `ttn/lorawan/v3/notification_service.proto`</a>

### <a name="ttn.lorawan.v3.CreateNotificationRequest">Message `CreateNotificationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entity_ids` | [`EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers) |  | The entity this notification is about. |
| `notification_type` | [`string`](#string) |  | The type of this notification. |
| `data` | [`google.protobuf.Any`](#google.protobuf.Any) |  | The data related to the notification. |
| `sender_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  | If the notification was triggered by a user action, this contains the identifiers of the user that triggered the notification. |
| `receivers` | [`NotificationReceiver`](#ttn.lorawan.v3.NotificationReceiver) | repeated | Receivers of the notification. |
| `email` | [`bool`](#bool) |  | Whether an email should be sent for the notification. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `entity_ids` | <p>`message.required`: `true`</p> |
| `notification_type` | <p>`string.min_len`: `1`</p><p>`string.max_len`: `100`</p> |
| `receivers` | <p>`repeated.min_items`: `1`</p><p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.CreateNotificationResponse">Message `CreateNotificationResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.EntityStateChangedNotification">Message `EntityStateChangedNotification`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `state` | [`State`](#ttn.lorawan.v3.State) |  |  |
| `state_description` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `state` | <p>`enum.defined_only`: `true`</p> |
| `state_description` | <p>`string.max_len`: `128`</p> |

### <a name="ttn.lorawan.v3.ListNotificationsRequest">Message `ListNotificationsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `receiver_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  | The IDs of the receiving user. |
| `status` | [`NotificationStatus`](#ttn.lorawan.v3.NotificationStatus) | repeated | Select notifications with these statuses. An empty list is interpreted as "all". |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `receiver_ids` | <p>`message.required`: `true`</p> |
| `status` | <p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListNotificationsResponse">Message `ListNotificationsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `notifications` | [`Notification`](#ttn.lorawan.v3.Notification) | repeated |  |

### <a name="ttn.lorawan.v3.Notification">Message `Notification`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  | The immutable ID of the notification. Generated by the server. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | The time when the notification was triggered. |
| `entity_ids` | [`EntityIdentifiers`](#ttn.lorawan.v3.EntityIdentifiers) |  | The entity this notification is about. |
| `notification_type` | [`string`](#string) |  | The type of this notification. |
| `data` | [`google.protobuf.Any`](#google.protobuf.Any) |  | The data related to the notification. |
| `sender_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  | If the notification was triggered by a user action, this contains the identifiers of the user that triggered the notification. |
| `receivers` | [`NotificationReceiver`](#ttn.lorawan.v3.NotificationReceiver) | repeated | Relation of the notification receiver to the entity. |
| `email` | [`bool`](#bool) |  | Whether an email was sent for the notification. |
| `status` | [`NotificationStatus`](#ttn.lorawan.v3.NotificationStatus) |  | The status of the notification. |
| `status_updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | The time when the notification status was updated. |

### <a name="ttn.lorawan.v3.UpdateNotificationStatusRequest">Message `UpdateNotificationStatusRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `receiver_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  | The IDs of the receiving user. |
| `ids` | [`string`](#string) | repeated | The IDs of the notifications to update the status of. |
| `status` | [`NotificationStatus`](#ttn.lorawan.v3.NotificationStatus) |  | The status to set on the notifications. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `receiver_ids` | <p>`message.required`: `true`</p> |
| `ids` | <p>`repeated.min_items`: `1`</p><p>`repeated.max_items`: `1000`</p><p>`repeated.unique`: `true`</p><p>`repeated.items.string.len`: `36`</p> |
| `status` | <p>`enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.NotificationReceiver">Enum `NotificationReceiver`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `NOTIFICATION_RECEIVER_UNKNOWN` | 0 |  |
| `NOTIFICATION_RECEIVER_COLLABORATOR` | 1 | Notification is received by collaborators of the entity. If the collaborator is an organization, the notification is received by organization members. |
| `NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT` | 3 | Notification is received by administrative contact of the entity. If this is an organization, the notification is received by organization members. |
| `NOTIFICATION_RECEIVER_TECHNICAL_CONTACT` | 4 | Notification is received by technical contact of the entity. If this is an organization, the notification is received by organization members. |

### <a name="ttn.lorawan.v3.NotificationStatus">Enum `NotificationStatus`</a>

| Name | Number | Description |
| ---- | ------ | ----------- |
| `NOTIFICATION_STATUS_UNSEEN` | 0 |  |
| `NOTIFICATION_STATUS_SEEN` | 1 |  |
| `NOTIFICATION_STATUS_ARCHIVED` | 2 |  |

### <a name="ttn.lorawan.v3.NotificationService">Service `NotificationService`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateNotificationRequest`](#ttn.lorawan.v3.CreateNotificationRequest) | [`CreateNotificationResponse`](#ttn.lorawan.v3.CreateNotificationResponse) | Create a new notification. Can only be called by internal services using cluster auth. |
| `List` | [`ListNotificationsRequest`](#ttn.lorawan.v3.ListNotificationsRequest) | [`ListNotificationsResponse`](#ttn.lorawan.v3.ListNotificationsResponse) | List the notifications for a user or an organization. When called with user credentials and empty receiver_ids, this will list notifications for the current user and its organizations. |
| `UpdateStatus` | [`UpdateNotificationStatusRequest`](#ttn.lorawan.v3.UpdateNotificationStatusRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Batch-update multiple notifications to the same status. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `List` | `` | `/api/v3` |  |
| `List` | `GET` | `/api/v3/users/{receiver_ids.user_id}/notifications` |  |
| `UpdateStatus` | `PATCH` | `/api/v3/users/{receiver_ids.user_id}/notifications` | `*` |

## <a name="ttn/lorawan/v3/oauth.proto">File `ttn/lorawan/v3/oauth.proto`</a>

### <a name="ttn.lorawan.v3.ListOAuthAccessTokensRequest">Message `ListOAuthAccessTokensRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `client_ids` | <p>`message.required`: `true`</p> |
| `order` | <p>`string.in`: `[ created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest">Message `ListOAuthClientAuthorizationsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `order` | <p>`string.in`: `[ created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.OAuthAccessToken">Message `OAuthAccessToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `user_session_id` | [`string`](#string) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `id` | [`string`](#string) |  |  |
| `access_token` | [`string`](#string) |  |  |
| `refresh_token` | [`string`](#string) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `user_session_id` | <p>`string.max_len`: `64`</p> |
| `client_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.OAuthAccessTokenIdentifiers">Message `OAuthAccessTokenIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `client_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.OAuthAccessTokens">Message `OAuthAccessTokens`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokens` | [`OAuthAccessToken`](#ttn.lorawan.v3.OAuthAccessToken) | repeated |  |

### <a name="ttn.lorawan.v3.OAuthAuthorizationCode">Message `OAuthAuthorizationCode`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `user_session_id` | [`string`](#string) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |
| `code` | [`string`](#string) |  |  |
| `redirect_uri` | [`string`](#string) |  |  |
| `state` | [`string`](#string) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `user_session_id` | <p>`string.max_len`: `64`</p> |
| `client_ids` | <p>`message.required`: `true`</p> |
| `redirect_uri` | <p>`string.uri_ref`: `true`</p> |

### <a name="ttn.lorawan.v3.OAuthClientAuthorization">Message `OAuthClientAuthorization`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `client_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers">Message `OAuthClientAuthorizationIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `client_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.OAuthClientAuthorizations">Message `OAuthClientAuthorizations`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authorizations` | [`OAuthClientAuthorization`](#ttn.lorawan.v3.OAuthClientAuthorization) | repeated |  |

## <a name="ttn/lorawan/v3/oauth_services.proto">File `ttn/lorawan/v3/oauth_services.proto`</a>

### <a name="ttn.lorawan.v3.OAuthAuthorizationRegistry">Service `OAuthAuthorizationRegistry`</a>

The OAuthAuthorizationRegistry service, exposed by the Identity Server,
is used to manage OAuth client authorizations for users.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `List` | [`ListOAuthClientAuthorizationsRequest`](#ttn.lorawan.v3.ListOAuthClientAuthorizationsRequest) | [`OAuthClientAuthorizations`](#ttn.lorawan.v3.OAuthClientAuthorizations) | List OAuth clients that are authorized by the user. |
| `ListTokens` | [`ListOAuthAccessTokensRequest`](#ttn.lorawan.v3.ListOAuthAccessTokensRequest) | [`OAuthAccessTokens`](#ttn.lorawan.v3.OAuthAccessTokens) | List OAuth access tokens issued to the OAuth client on behalf of the user. |
| `Delete` | [`OAuthClientAuthorizationIdentifiers`](#ttn.lorawan.v3.OAuthClientAuthorizationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete (de-authorize) an OAuth client for the user. |
| `DeleteToken` | [`OAuthAccessTokenIdentifiers`](#ttn.lorawan.v3.OAuthAccessTokenIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete (invalidate) an OAuth access token. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `List` | `GET` | `/api/v3/users/{user_ids.user_id}/authorizations` |  |
| `ListTokens` | `GET` | `/api/v3/users/{user_ids.user_id}/authorizations/{client_ids.client_id}/tokens` |  |
| `Delete` | `DELETE` | `/api/v3/users/{user_ids.user_id}/authorizations/{client_ids.client_id}` |  |
| `DeleteToken` | `DELETE` | `/api/v3/users/{user_ids.user_id}/authorizations/{client_ids.client_id}/tokens/{id}` |  |

## <a name="ttn/lorawan/v3/organization.proto">File `ttn/lorawan/v3/organization.proto`</a>

### <a name="ttn.lorawan.v3.CreateOrganizationAPIKeyRequest">Message `CreateOrganizationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `name` | [`string`](#string) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `rights` | <p>`repeated.min_items`: `1`</p><p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `expires_at` | <p>`timestamp.gt_now`: `true`</p> |

### <a name="ttn.lorawan.v3.CreateOrganizationRequest">Message `CreateOrganizationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization` | [`Organization`](#ttn.lorawan.v3.Organization) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | Collaborator to grant all rights on the newly created application. NOTE: It is currently not possible to have organizations collaborating on other organizations. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteOrganizationAPIKeyRequest">Message `DeleteOrganizationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteOrganizationCollaboratorRequest">Message `DeleteOrganizationCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `collaborator_ids` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |
| `collaborator_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetOrganizationAPIKeyRequest">Message `GetOrganizationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  | Unique public identifier for the API key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetOrganizationCollaboratorRequest">Message `GetOrganizationCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | NOTE: It is currently not possible to have organizations collaborating on other organizations. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetOrganizationRequest">Message `GetOrganizationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the organization fields that should be returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.ListOrganizationAPIKeysRequest">Message `ListOrganizationAPIKeysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path. Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |
| `order` | <p>`string.in`: `[ api_key_id -api_key_id name -name created_at -created_at expires_at -expires_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListOrganizationCollaboratorsRequest">Message `ListOrganizationCollaboratorsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `order` | <p>`string.in`: `[ id -id -rights rights]`</p> |

### <a name="ttn.lorawan.v3.ListOrganizationsRequest">Message `ListOrganizationsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collaborator` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  | By default we list all organizations the caller has rights on. Set the user to instead list the organizations where the user or organization is collaborator on. NOTE: It is currently not possible to have organizations collaborating on other organizations. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the organization fields that should be returned. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted organizations. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `order` | <p>`string.in`: `[ organization_id -organization_id name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.Organization">Message `Organization`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  | The identifiers of the organization. These are public and can be seen by any authenticated user in the network. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the organization was created. This information is public and can be seen by any authenticated user in the network. |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the organization was last updated. This information is public and can be seen by any authenticated user in the network. |
| `deleted_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the organization was deleted. This information is public and can be seen by any authenticated user in the network. |
| `name` | [`string`](#string) |  | The name of the organization. This information is public and can be seen by any authenticated user in the network. |
| `description` | [`string`](#string) |  | A description for the organization. |
| `attributes` | [`Organization.AttributesEntry`](#ttn.lorawan.v3.Organization.AttributesEntry) | repeated | Key-value attributes for this organization. Typically used for organizing organizations or for storing integration-specific data. |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated | Contact information for this organization. Typically used to indicate who to contact with security/billing questions about the organization. This field is deprecated. Use administrative_contact and technical_contact instead. |
| `administrative_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `technical_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `description` | <p>`string.max_len`: `2000`</p> |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |
| `contact_info` | <p>`repeated.max_items`: `10`</p> |

### <a name="ttn.lorawan.v3.Organization.AttributesEntry">Message `Organization.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.Organizations">Message `Organizations`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organizations` | [`Organization`](#ttn.lorawan.v3.Organization) | repeated |  |

### <a name="ttn.lorawan.v3.SetOrganizationCollaboratorRequest">Message `SetOrganizationCollaboratorRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `collaborator` | [`Collaborator`](#ttn.lorawan.v3.Collaborator) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |
| `collaborator` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest">Message `UpdateOrganizationAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |
| `api_key` | [`APIKey`](#ttn.lorawan.v3.APIKey) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the api key fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization_ids` | <p>`message.required`: `true`</p> |
| `api_key` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateOrganizationRequest">Message `UpdateOrganizationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `organization` | [`Organization`](#ttn.lorawan.v3.Organization) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the organization fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `organization` | <p>`message.required`: `true`</p> |

## <a name="ttn/lorawan/v3/organization_services.proto">File `ttn/lorawan/v3/organization_services.proto`</a>

### <a name="ttn.lorawan.v3.OrganizationAccess">Service `OrganizationAccess`</a>

The OrganizationAcces service, exposed by the Identity Server, is used to manage
API keys and collaborators of organizations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListRights` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) | [`Rights`](#ttn.lorawan.v3.Rights) | List the rights the caller has on this organization. |
| `CreateAPIKey` | [`CreateOrganizationAPIKeyRequest`](#ttn.lorawan.v3.CreateOrganizationAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Create an API key scoped to this organization. Organization API keys can give access to the organization itself, as well as any application, gateway and OAuth client this organization is a collaborator of. |
| `ListAPIKeys` | [`ListOrganizationAPIKeysRequest`](#ttn.lorawan.v3.ListOrganizationAPIKeysRequest) | [`APIKeys`](#ttn.lorawan.v3.APIKeys) | List the API keys for this organization. |
| `GetAPIKey` | [`GetOrganizationAPIKeyRequest`](#ttn.lorawan.v3.GetOrganizationAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Get a single API key of this organization. |
| `UpdateAPIKey` | [`UpdateOrganizationAPIKeyRequest`](#ttn.lorawan.v3.UpdateOrganizationAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Update the rights of an API key of the organization. This method can also be used to delete the API key, by giving it no rights. The caller is required to have all assigned or/and removed rights. |
| `DeleteAPIKey` | [`DeleteOrganizationAPIKeyRequest`](#ttn.lorawan.v3.DeleteOrganizationAPIKeyRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a single API key of this organization. |
| `GetCollaborator` | [`GetOrganizationCollaboratorRequest`](#ttn.lorawan.v3.GetOrganizationCollaboratorRequest) | [`GetCollaboratorResponse`](#ttn.lorawan.v3.GetCollaboratorResponse) | Get the rights of a collaborator (member) of the organization. Pseudo-rights in the response (such as the "_ALL" right) are not expanded. |
| `SetCollaborator` | [`SetOrganizationCollaboratorRequest`](#ttn.lorawan.v3.SetOrganizationCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Set the rights of a collaborator (member) on the organization. Organization collaborators can get access to the organization itself, as well as any application, gateway and OAuth client this organization is a collaborator of. This method can also be used to delete the collaborator, by giving them no rights. The caller is required to have all assigned or/and removed rights. |
| `ListCollaborators` | [`ListOrganizationCollaboratorsRequest`](#ttn.lorawan.v3.ListOrganizationCollaboratorsRequest) | [`Collaborators`](#ttn.lorawan.v3.Collaborators) | List the collaborators on this organization. |
| `DeleteCollaborator` | [`DeleteOrganizationCollaboratorRequest`](#ttn.lorawan.v3.DeleteOrganizationCollaboratorRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | DeleteCollaborator removes a collaborator from an organization. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListRights` | `GET` | `/api/v3/organizations/{organization_id}/rights` |  |
| `CreateAPIKey` | `POST` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys` | `*` |
| `ListAPIKeys` | `GET` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys` |  |
| `GetAPIKey` | `GET` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys/{key_id}` |  |
| `UpdateAPIKey` | `PUT` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys/{api_key.id}` | `*` |
| `DeleteAPIKey` | `DELETE` | `/api/v3/organizations/{organization_ids.organization_id}/api-keys/{key_id}` |  |
| `GetCollaborator` | `` | `/api/v3` |  |
| `GetCollaborator` | `GET` | `/api/v3/organizations/{organization_ids.organization_id}/collaborator/user/{collaborator.user_ids.user_id}` |  |
| `SetCollaborator` | `PUT` | `/api/v3/organizations/{organization_ids.organization_id}/collaborators` | `*` |
| `ListCollaborators` | `GET` | `/api/v3/organizations/{organization_ids.organization_id}/collaborators` |  |
| `DeleteCollaborator` | `DELETE` | `/api/v3/organizations/{organization_ids.organization_id}/collaborators/user/{collaborator_ids.user_ids.user_id}` |  |

### <a name="ttn.lorawan.v3.OrganizationRegistry">Service `OrganizationRegistry`</a>

The OrganizationRegistry service, exposed by the Identity Server, is used to manage
organization registrations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateOrganizationRequest`](#ttn.lorawan.v3.CreateOrganizationRequest) | [`Organization`](#ttn.lorawan.v3.Organization) | Create a new organization. This also sets the given user as first collaborator with all possible rights. |
| `Get` | [`GetOrganizationRequest`](#ttn.lorawan.v3.GetOrganizationRequest) | [`Organization`](#ttn.lorawan.v3.Organization) | Get the organization with the given identifiers, selecting the fields specified in the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `List` | [`ListOrganizationsRequest`](#ttn.lorawan.v3.ListOrganizationsRequest) | [`Organizations`](#ttn.lorawan.v3.Organizations) | List organizations where the given user or organization is a direct collaborator. If no user or organization is given, this returns the organizations the caller has access to. Similar to Get, this selects the fields given by the field mask. More or less fields may be returned, depending on the rights of the caller. |
| `Update` | [`UpdateOrganizationRequest`](#ttn.lorawan.v3.UpdateOrganizationRequest) | [`Organization`](#ttn.lorawan.v3.Organization) | Update the organization, changing the fields specified by the field mask to the provided values. |
| `Delete` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the organization. This may not release the organization ID for reuse. |
| `Restore` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Restore a recently deleted organization. Deployment configuration may specify if, and for how long after deletion, entities can be restored. |
| `Purge` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Purge the organization. This will release the organization ID for reuse. The user is responsible for clearing data from any (external) integrations that may store and expose data by user or organization ID. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Create` | `POST` | `/api/v3/users/{collaborator.user_ids.user_id}/organizations` | `*` |
| `Get` | `GET` | `/api/v3/organizations/{organization_ids.organization_id}` |  |
| `List` | `GET` | `/api/v3/organizations` |  |
| `List` | `GET` | `/api/v3/users/{collaborator.user_ids.user_id}/organizations` |  |
| `Update` | `PUT` | `/api/v3/organizations/{organization.ids.organization_id}` | `*` |
| `Delete` | `DELETE` | `/api/v3/organizations/{organization_id}` |  |
| `Restore` | `POST` | `/api/v3/organizations/{organization_id}/restore` |  |
| `Purge` | `DELETE` | `/api/v3/organizations/{organization_id}/purge` |  |

## <a name="ttn/lorawan/v3/packetbrokeragent.proto">File `ttn/lorawan/v3/packetbrokeragent.proto`</a>

### <a name="ttn.lorawan.v3.ListForwarderRoutingPoliciesRequest">Message `ListForwarderRoutingPoliciesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `home_network_id` | [`PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier) |  | Packet Broker identifier of the Home Network. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

### <a name="ttn.lorawan.v3.ListHomeNetworkRoutingPoliciesRequest">Message `ListHomeNetworkRoutingPoliciesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListPacketBrokerHomeNetworksRequest">Message `ListPacketBrokerHomeNetworksRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `tenant_id_contains` | [`string`](#string) |  | Filter by tenant ID. |
| `name_contains` | [`string`](#string) |  | Filter by name. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `tenant_id_contains` | <p>`string.max_len`: `100`</p> |
| `name_contains` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.ListPacketBrokerNetworksRequest">Message `ListPacketBrokerNetworksRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `with_routing_policy` | [`bool`](#bool) |  | If true, list only the Forwarders and Home Networks with whom a routing policy has been defined in either direction. |
| `tenant_id_contains` | [`string`](#string) |  | Filter by tenant ID. |
| `name_contains` | [`string`](#string) |  | Filter by name. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `limit` | <p>`uint32.lte`: `1000`</p> |
| `tenant_id_contains` | <p>`string.max_len`: `100`</p> |
| `name_contains` | <p>`string.max_len`: `100`</p> |

### <a name="ttn.lorawan.v3.PacketBrokerAgentCompoundUplinkToken">Message `PacketBrokerAgentCompoundUplinkToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway` | [`bytes`](#bytes) |  |  |
| `forwarder` | [`bytes`](#bytes) |  |  |
| `agent` | [`PacketBrokerAgentUplinkToken`](#ttn.lorawan.v3.PacketBrokerAgentUplinkToken) |  |  |

### <a name="ttn.lorawan.v3.PacketBrokerAgentEncryptedPayload">Message `PacketBrokerAgentEncryptedPayload`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ciphertext` | [`bytes`](#bytes) |  |  |
| `nonce` | [`bytes`](#bytes) |  |  |

### <a name="ttn.lorawan.v3.PacketBrokerAgentGatewayUplinkToken">Message `PacketBrokerAgentGatewayUplinkToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_uid` | [`string`](#string) |  |  |
| `token` | [`bytes`](#bytes) |  |  |

### <a name="ttn.lorawan.v3.PacketBrokerAgentUplinkToken">Message `PacketBrokerAgentUplinkToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `forwarder_net_id` | [`bytes`](#bytes) |  |  |
| `forwarder_tenant_id` | [`string`](#string) |  |  |
| `forwarder_cluster_id` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.PacketBrokerDefaultGatewayVisibility">Message `PacketBrokerDefaultGatewayVisibility`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp when the policy got last updated. |
| `visibility` | [`PacketBrokerGatewayVisibility`](#ttn.lorawan.v3.PacketBrokerGatewayVisibility) |  |  |

### <a name="ttn.lorawan.v3.PacketBrokerDefaultRoutingPolicy">Message `PacketBrokerDefaultRoutingPolicy`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp when the policy got last updated. |
| `uplink` | [`PacketBrokerRoutingPolicyUplink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyUplink) |  | Uplink policy. |
| `downlink` | [`PacketBrokerRoutingPolicyDownlink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyDownlink) |  | Downlink policy. |

### <a name="ttn.lorawan.v3.PacketBrokerDevAddrBlock">Message `PacketBrokerDevAddrBlock`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr_prefix` | [`DevAddrPrefix`](#ttn.lorawan.v3.DevAddrPrefix) |  |  |
| `home_network_cluster_id` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.PacketBrokerGateway">Message `PacketBrokerGateway`</a>

Gateway respresentation for Packet Broker.
This is a subset and superset of the Gateway message using the same data types and field tags to achieve initial wire compatibility.
There is no (longer) wire compatibility needed; new fields may use any tag.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`PacketBrokerGateway.GatewayIdentifiers`](#ttn.lorawan.v3.PacketBrokerGateway.GatewayIdentifiers) |  |  |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated | This field is deprecated. Use administrative_contact and technical_contact instead. |
| `administrative_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `technical_contact` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `antennas` | [`GatewayAntenna`](#ttn.lorawan.v3.GatewayAntenna) | repeated |  |
| `status_public` | [`bool`](#bool) |  |  |
| `location_public` | [`bool`](#bool) |  |  |
| `frequency_plan_ids` | [`string`](#string) | repeated |  |
| `update_location_from_status` | [`bool`](#bool) |  |  |
| `online` | [`bool`](#bool) |  |  |
| `rx_rate` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Received packets rate (number of packets per hour). This field gets updated when a value is set. |
| `tx_rate` | [`google.protobuf.FloatValue`](#google.protobuf.FloatValue) |  | Transmitted packets rate (number of packets per hour). This field gets updated when a value is set. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `contact_info` | <p>`repeated.max_items`: `10`</p> |
| `antennas` | <p>`repeated.max_items`: `8`</p> |
| `frequency_plan_ids` | <p>`repeated.max_items`: `8`</p><p>`repeated.items.string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.PacketBrokerGateway.GatewayIdentifiers">Message `PacketBrokerGateway.GatewayIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway_id` | [`string`](#string) |  |  |
| `eui` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[_-]?[a-z0-9]){2,}$`</p> |
| `eui` | <p>`bytes.len`: `8`</p> |

### <a name="ttn.lorawan.v3.PacketBrokerGatewayVisibility">Message `PacketBrokerGatewayVisibility`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `location` | [`bool`](#bool) |  | Show location. |
| `antenna_placement` | [`bool`](#bool) |  | Show antenna placement (indoor/outdoor). |
| `antenna_count` | [`bool`](#bool) |  | Show antenna count. |
| `fine_timestamps` | [`bool`](#bool) |  | Show whether the gateway produces fine timestamps. |
| `contact_info` | [`bool`](#bool) |  | Show contact information. |
| `status` | [`bool`](#bool) |  | Show status (online/offline). |
| `frequency_plan` | [`bool`](#bool) |  | Show frequency plan. |
| `packet_rates` | [`bool`](#bool) |  | Show receive and transmission packet rates. |

### <a name="ttn.lorawan.v3.PacketBrokerInfo">Message `PacketBrokerInfo`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `registration` | [`PacketBrokerNetwork`](#ttn.lorawan.v3.PacketBrokerNetwork) |  | The current registration, unset if there isn't a registration. |
| `forwarder_enabled` | [`bool`](#bool) |  | Whether the server is configured as Forwarder (with gateways). |
| `home_network_enabled` | [`bool`](#bool) |  | Whether the server is configured as Home Network (with end devices). |
| `register_enabled` | [`bool`](#bool) |  | Whether the registration can be changed. |

### <a name="ttn.lorawan.v3.PacketBrokerNetwork">Message `PacketBrokerNetwork`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier) |  | Packet Broker network identifier. |
| `name` | [`string`](#string) |  | Name of the network. |
| `dev_addr_blocks` | [`PacketBrokerDevAddrBlock`](#ttn.lorawan.v3.PacketBrokerDevAddrBlock) | repeated | DevAddr blocks that are assigned to this registration. |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated | Contact information. This field is deprecated. Use administrative_contact and technical_contact instead. |
| `administrative_contact` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) |  |  |
| `technical_contact` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) |  |  |
| `listed` | [`bool`](#bool) |  | Whether the network is listed so it can be viewed by other networks. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `contact_info` | <p>`repeated.max_items`: `10`</p> |

### <a name="ttn.lorawan.v3.PacketBrokerNetworkIdentifier">Message `PacketBrokerNetworkIdentifier`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `net_id` | [`uint32`](#uint32) |  | LoRa Alliance NetID. |
| `tenant_id` | [`string`](#string) |  | Tenant identifier if the registration leases DevAddr blocks from a NetID. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `tenant_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |

### <a name="ttn.lorawan.v3.PacketBrokerNetworks">Message `PacketBrokerNetworks`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `networks` | [`PacketBrokerNetwork`](#ttn.lorawan.v3.PacketBrokerNetwork) | repeated |  |

### <a name="ttn.lorawan.v3.PacketBrokerRegisterRequest">Message `PacketBrokerRegisterRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `listed` | [`google.protobuf.BoolValue`](#google.protobuf.BoolValue) |  | Whether the network should be listed in Packet Broker. If unset, the value is taken from the registration settings. |

### <a name="ttn.lorawan.v3.PacketBrokerRoutingPolicies">Message `PacketBrokerRoutingPolicies`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `policies` | [`PacketBrokerRoutingPolicy`](#ttn.lorawan.v3.PacketBrokerRoutingPolicy) | repeated |  |

### <a name="ttn.lorawan.v3.PacketBrokerRoutingPolicy">Message `PacketBrokerRoutingPolicy`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `forwarder_id` | [`PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier) |  | Packet Broker identifier of the Forwarder. |
| `home_network_id` | [`PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier) |  | Packet Broker identifier of the Home Network. |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | Timestamp when the policy got last updated. |
| `uplink` | [`PacketBrokerRoutingPolicyUplink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyUplink) |  | Uplink policy. |
| `downlink` | [`PacketBrokerRoutingPolicyDownlink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyDownlink) |  | Downlink policy. |

### <a name="ttn.lorawan.v3.PacketBrokerRoutingPolicyDownlink">Message `PacketBrokerRoutingPolicyDownlink`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_accept` | [`bool`](#bool) |  | Allow join-accept messages. |
| `mac_data` | [`bool`](#bool) |  | Allow downlink messages with FPort of 0. |
| `application_data` | [`bool`](#bool) |  | Allow downlink messages with FPort between 1 and 255. |

### <a name="ttn.lorawan.v3.PacketBrokerRoutingPolicyUplink">Message `PacketBrokerRoutingPolicyUplink`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_request` | [`bool`](#bool) |  | Forward join-request messages. |
| `mac_data` | [`bool`](#bool) |  | Forward uplink messages with FPort of 0. |
| `application_data` | [`bool`](#bool) |  | Forward uplink messages with FPort between 1 and 255. |
| `signal_quality` | [`bool`](#bool) |  | Forward RSSI and SNR. |
| `localization` | [`bool`](#bool) |  | Forward gateway location, RSSI, SNR and fine timestamp. |

### <a name="ttn.lorawan.v3.SetPacketBrokerDefaultGatewayVisibilityRequest">Message `SetPacketBrokerDefaultGatewayVisibilityRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `visibility` | [`PacketBrokerGatewayVisibility`](#ttn.lorawan.v3.PacketBrokerGatewayVisibility) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `visibility` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.SetPacketBrokerDefaultRoutingPolicyRequest">Message `SetPacketBrokerDefaultRoutingPolicyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uplink` | [`PacketBrokerRoutingPolicyUplink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyUplink) |  | Uplink policy. |
| `downlink` | [`PacketBrokerRoutingPolicyDownlink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyDownlink) |  | Downlink policy. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `uplink` | <p>`message.required`: `true`</p> |
| `downlink` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.SetPacketBrokerRoutingPolicyRequest">Message `SetPacketBrokerRoutingPolicyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `home_network_id` | [`PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier) |  | Packet Broker identifier of the Home Network. |
| `uplink` | [`PacketBrokerRoutingPolicyUplink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyUplink) |  | Uplink policy. |
| `downlink` | [`PacketBrokerRoutingPolicyDownlink`](#ttn.lorawan.v3.PacketBrokerRoutingPolicyDownlink) |  | Downlink policy. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `uplink` | <p>`message.required`: `true`</p> |
| `downlink` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdatePacketBrokerGatewayRequest">Message `UpdatePacketBrokerGatewayRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gateway` | [`PacketBrokerGateway`](#ttn.lorawan.v3.PacketBrokerGateway) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the gateway fields that are considered for update. Online status is only updated if status_public is set. If status_public is set and false, the status will be reset. If status_public is set and true, the online status is taken from the online field. The return message contains the duration online_ttl for how long the gateway is considered online. Location is only updated if location_public is set. If location_public is set and false, the location will be reset. If location_public is set and true, the first antenna location will be used as gateway location. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `gateway` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdatePacketBrokerGatewayResponse">Message `UpdatePacketBrokerGatewayResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `online_ttl` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  | Time to live of the online status. |

### <a name="ttn.lorawan.v3.GsPba">Service `GsPba`</a>

The GsPba service connects a Gateway Server to a Packet Broker Agent.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `PublishUplink` | [`GatewayUplinkMessage`](#ttn.lorawan.v3.GatewayUplinkMessage) | [`.google.protobuf.Empty`](#google.protobuf.Empty) |  |
| `UpdateGateway` | [`UpdatePacketBrokerGatewayRequest`](#ttn.lorawan.v3.UpdatePacketBrokerGatewayRequest) | [`UpdatePacketBrokerGatewayResponse`](#ttn.lorawan.v3.UpdatePacketBrokerGatewayResponse) | Update the gateway, changing the fields specified by the field mask to the provided values. To mark a gateway as online, call this rpc setting online to true, include status_public in field_mask and keep calling this rpc before the returned online_ttl passes to keep the gateway online. |

### <a name="ttn.lorawan.v3.NsPba">Service `NsPba`</a>

The NsPba service connects a Network Server to a Packet Broker Agent.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `PublishDownlink` | [`DownlinkMessage`](#ttn.lorawan.v3.DownlinkMessage) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | PublishDownlink instructs the Packet Broker Agent to publish a downlink message to Packet Broker Router. |

### <a name="ttn.lorawan.v3.Pba">Service `Pba`</a>

The Pba service allows clients to manage peering through Packet Broker.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetInfo` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`PacketBrokerInfo`](#ttn.lorawan.v3.PacketBrokerInfo) | Get information about the Packet Broker registration. Viewing Packet Packet information requires administrative access. |
| `Register` | [`PacketBrokerRegisterRequest`](#ttn.lorawan.v3.PacketBrokerRegisterRequest) | [`PacketBrokerNetwork`](#ttn.lorawan.v3.PacketBrokerNetwork) | Register with Packet Broker. If no registration exists, it will be created. Any existing registration will be updated. Registration settings not in the request message are taken from Packet Broker Agent configuration and caller context. Packet Broker registration requires administrative access. Packet Broker registration is only supported for tenants and requires Packet Broker Agent to be configured with NetID level authentication. Use rpc GetInfo and check register_enabled to check whether this rpc is enabled. |
| `Deregister` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Deregister from Packet Broker. Packet Broker deregistration requires administrative access. Packet Broker deregistration is only supported for tenants and requires Packet Broker Agent to be configured with NetID level authentication. Use rpc GetInfo and check register_enabled to check whether this rpc is enabled. |
| `ListHomeNetworkRoutingPolicies` | [`ListHomeNetworkRoutingPoliciesRequest`](#ttn.lorawan.v3.ListHomeNetworkRoutingPoliciesRequest) | [`PacketBrokerRoutingPolicies`](#ttn.lorawan.v3.PacketBrokerRoutingPolicies) | List the routing policies that Packet Broker Agent as Forwarder configured with Home Networks. Listing routing policies requires administrative access. |
| `GetHomeNetworkRoutingPolicy` | [`PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier) | [`PacketBrokerRoutingPolicy`](#ttn.lorawan.v3.PacketBrokerRoutingPolicy) | Get the routing policy for the given Home Network. Getting routing policies requires administrative access. |
| `SetHomeNetworkRoutingPolicy` | [`SetPacketBrokerRoutingPolicyRequest`](#ttn.lorawan.v3.SetPacketBrokerRoutingPolicyRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Set the routing policy for the given Home Network. Setting routing policies requires administrative access. |
| `DeleteHomeNetworkRoutingPolicy` | [`PacketBrokerNetworkIdentifier`](#ttn.lorawan.v3.PacketBrokerNetworkIdentifier) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the routing policy for the given Home Network. Deleting routing policies requires administrative access. |
| `GetHomeNetworkDefaultRoutingPolicy` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`PacketBrokerDefaultRoutingPolicy`](#ttn.lorawan.v3.PacketBrokerDefaultRoutingPolicy) | Get the default routing policy. Getting routing policies requires administrative access. |
| `SetHomeNetworkDefaultRoutingPolicy` | [`SetPacketBrokerDefaultRoutingPolicyRequest`](#ttn.lorawan.v3.SetPacketBrokerDefaultRoutingPolicyRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Set the default routing policy. Setting routing policies requires administrative access. |
| `DeleteHomeNetworkDefaultRoutingPolicy` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Deletes the default routing policy. Deleting routing policies requires administrative access. |
| `GetHomeNetworkDefaultGatewayVisibility` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`PacketBrokerDefaultGatewayVisibility`](#ttn.lorawan.v3.PacketBrokerDefaultGatewayVisibility) | Get the default gateway visibility. Getting gateway visibilities requires administrative access. |
| `SetHomeNetworkDefaultGatewayVisibility` | [`SetPacketBrokerDefaultGatewayVisibilityRequest`](#ttn.lorawan.v3.SetPacketBrokerDefaultGatewayVisibilityRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Set the default gateway visibility. Setting gateway visibilities requires administrative access. |
| `DeleteHomeNetworkDefaultGatewayVisibility` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Deletes the default gateway visibility. Deleting gateway visibilities requires administrative access. |
| `ListNetworks` | [`ListPacketBrokerNetworksRequest`](#ttn.lorawan.v3.ListPacketBrokerNetworksRequest) | [`PacketBrokerNetworks`](#ttn.lorawan.v3.PacketBrokerNetworks) | List all listed networks. Listing networks requires administrative access. |
| `ListHomeNetworks` | [`ListPacketBrokerHomeNetworksRequest`](#ttn.lorawan.v3.ListPacketBrokerHomeNetworksRequest) | [`PacketBrokerNetworks`](#ttn.lorawan.v3.PacketBrokerNetworks) | List the listed home networks for which routing policies can be configured. Listing home networks requires administrative access. |
| `ListForwarderRoutingPolicies` | [`ListForwarderRoutingPoliciesRequest`](#ttn.lorawan.v3.ListForwarderRoutingPoliciesRequest) | [`PacketBrokerRoutingPolicies`](#ttn.lorawan.v3.PacketBrokerRoutingPolicies) | List the routing policies that Forwarders configured with Packet Broker Agent as Home Network. Listing routing policies requires administrative access. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetInfo` | `GET` | `/api/v3/pba/info` |  |
| `Register` | `PUT` | `/api/v3/pba/registration` | `*` |
| `Register` | `POST` | `/api/v3/pba/registration` | `*` |
| `Deregister` | `DELETE` | `/api/v3/pba/registration` |  |
| `ListHomeNetworkRoutingPolicies` | `GET` | `/api/v3/pba/home-networks/policies` |  |
| `GetHomeNetworkRoutingPolicy` | `GET` | `/api/v3/pba/home-networks/policies/{net_id}` |  |
| `GetHomeNetworkRoutingPolicy` | `GET` | `/api/v3/pba/home-networks/policies/{net_id}/{tenant_id}` |  |
| `SetHomeNetworkRoutingPolicy` | `PUT` | `/api/v3/pba/home-networks/policies/{home_network_id.net_id}` | `*` |
| `SetHomeNetworkRoutingPolicy` | `POST` | `/api/v3/pba/home-networks/policies/{home_network_id.net_id}` | `*` |
| `SetHomeNetworkRoutingPolicy` | `PUT` | `/api/v3/pba/home-networks/policies/{home_network_id.net_id}/{home_network_id.tenant_id}` | `*` |
| `SetHomeNetworkRoutingPolicy` | `POST` | `/api/v3/pba/home-networks/policies/{home_network_id.net_id}/{home_network_id.tenant_id}` | `*` |
| `DeleteHomeNetworkRoutingPolicy` | `DELETE` | `/api/v3/pba/home-networks/policies/{net_id}` |  |
| `DeleteHomeNetworkRoutingPolicy` | `DELETE` | `/api/v3/pba/home-networks/policies/{net_id}/{tenant_id}` |  |
| `GetHomeNetworkDefaultRoutingPolicy` | `GET` | `/api/v3/pba/home-networks/policies/default` |  |
| `SetHomeNetworkDefaultRoutingPolicy` | `PUT` | `/api/v3/pba/home-networks/policies/default` | `*` |
| `SetHomeNetworkDefaultRoutingPolicy` | `POST` | `/api/v3/pba/home-networks/policies/default` | `*` |
| `DeleteHomeNetworkDefaultRoutingPolicy` | `DELETE` | `/api/v3/pba/home-networks/policies/default` |  |
| `GetHomeNetworkDefaultGatewayVisibility` | `GET` | `/api/v3/pba/home-networks/gateway-visibilities/default` |  |
| `SetHomeNetworkDefaultGatewayVisibility` | `PUT` | `/api/v3/pba/home-networks/gateway-visibilities/default` | `*` |
| `SetHomeNetworkDefaultGatewayVisibility` | `POST` | `/api/v3/pba/home-networks/gateway-visibilities/default` | `*` |
| `DeleteHomeNetworkDefaultGatewayVisibility` | `DELETE` | `/api/v3/pba/home-networks/gateway-visibilities/default` |  |
| `ListNetworks` | `GET` | `/api/v3/pba/networks` |  |
| `ListHomeNetworks` | `GET` | `/api/v3/pba/home-networks` |  |
| `ListForwarderRoutingPolicies` | `GET` | `/api/v3/pba/forwarders/policies` |  |

## <a name="ttn/lorawan/v3/picture.proto">File `ttn/lorawan/v3/picture.proto`</a>

### <a name="ttn.lorawan.v3.Picture">Message `Picture`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `embedded` | [`Picture.Embedded`](#ttn.lorawan.v3.Picture.Embedded) |  | Embedded picture. Omitted if there are external URLs available (in sizes). |
| `sizes` | [`Picture.SizesEntry`](#ttn.lorawan.v3.Picture.SizesEntry) | repeated | URLs of the picture for different sizes, if available on a CDN. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `sizes` | <p>`map.values.string.uri_ref`: `true`</p> |

### <a name="ttn.lorawan.v3.Picture.Embedded">Message `Picture.Embedded`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mime_type` | [`string`](#string) |  | MIME type of the picture. |
| `data` | [`bytes`](#bytes) |  | Picture data. A data URI can be constructed as follows: `data:<mime_type>;base64,<data>`. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `mime_type` | <p>`string.max_len`: `32`</p> |
| `data` | <p>`bytes.max_len`: `8388608`</p> |

### <a name="ttn.lorawan.v3.Picture.SizesEntry">Message `Picture.SizesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`uint32`](#uint32) |  |  |
| `value` | [`string`](#string) |  |  |

## <a name="ttn/lorawan/v3/qrcodegenerator.proto">File `ttn/lorawan/v3/qrcodegenerator.proto`</a>

### <a name="ttn.lorawan.v3.GenerateEndDeviceQRCodeRequest">Message `GenerateEndDeviceQRCodeRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `format_id` | [`string`](#string) |  | QR code format identifier. Enumerate available formats with rpc ListFormats in the EndDeviceQRCodeGenerator service. |
| `end_device` | [`EndDevice`](#ttn.lorawan.v3.EndDevice) |  | End device to use as input to generate the QR code. |
| `image` | [`GenerateEndDeviceQRCodeRequest.Image`](#ttn.lorawan.v3.GenerateEndDeviceQRCodeRequest.Image) |  | If set, the server will render the QR code image according to these settings. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `format_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |
| `end_device` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GenerateEndDeviceQRCodeRequest.Image">Message `GenerateEndDeviceQRCodeRequest.Image`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `image_size` | [`uint32`](#uint32) |  | Requested QR code image dimension in pixels. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `image_size` | <p>`uint32.lte`: `1000`</p><p>`uint32.gte`: `10`</p> |

### <a name="ttn.lorawan.v3.GenerateQRCodeResponse">Message `GenerateQRCodeResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `text` | [`string`](#string) |  | Text representation of the QR code contents. |
| `image` | [`Picture`](#ttn.lorawan.v3.Picture) |  | QR code in PNG format, if requested. |

### <a name="ttn.lorawan.v3.GetQRCodeFormatRequest">Message `GetQRCodeFormatRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `format_id` | [`string`](#string) |  | QR code format identifier. Enumerate available formats with rpc ListFormats in the EndDeviceQRCodeGenerator service. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `format_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.ParseEndDeviceQRCodeRequest">Message `ParseEndDeviceQRCodeRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `format_id` | [`string`](#string) |  | QR code format identifier. Enumerate available formats with the rpc `ListFormats`. If this field is not specified, the server will attempt to parse the data with each known format. |
| `qr_code` | [`bytes`](#bytes) |  | Raw QR code contents. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `format_id` | <p>`string.max_len`: `36`</p><p>`string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$|^$`</p> |
| `qr_code` | <p>`bytes.min_len`: `10`</p><p>`bytes.max_len`: `1024`</p> |

### <a name="ttn.lorawan.v3.ParseEndDeviceQRCodeResponse">Message `ParseEndDeviceQRCodeResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `format_id` | [`string`](#string) |  | Identifier of the format used to successfully parse the QR code data. |
| `end_device_template` | [`EndDeviceTemplate`](#ttn.lorawan.v3.EndDeviceTemplate) |  |  |

### <a name="ttn.lorawan.v3.QRCodeFormat">Message `QRCodeFormat`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [`string`](#string) |  |  |
| `description` | [`string`](#string) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The entity fields required to generate the QR code. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `name` | <p>`string.max_len`: `100`</p> |
| `description` | <p>`string.max_len`: `200`</p> |

### <a name="ttn.lorawan.v3.QRCodeFormats">Message `QRCodeFormats`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `formats` | [`QRCodeFormats.FormatsEntry`](#ttn.lorawan.v3.QRCodeFormats.FormatsEntry) | repeated | Available formats. The map key is the format identifier. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `formats` | <p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p> |

### <a name="ttn.lorawan.v3.QRCodeFormats.FormatsEntry">Message `QRCodeFormats.FormatsEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`QRCodeFormat`](#ttn.lorawan.v3.QRCodeFormat) |  |  |

### <a name="ttn.lorawan.v3.EndDeviceQRCodeGenerator">Service `EndDeviceQRCodeGenerator`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetFormat` | [`GetQRCodeFormatRequest`](#ttn.lorawan.v3.GetQRCodeFormatRequest) | [`QRCodeFormat`](#ttn.lorawan.v3.QRCodeFormat) | Return the QR code format. |
| `ListFormats` | [`.google.protobuf.Empty`](#google.protobuf.Empty) | [`QRCodeFormats`](#ttn.lorawan.v3.QRCodeFormats) | Returns the supported formats. |
| `Generate` | [`GenerateEndDeviceQRCodeRequest`](#ttn.lorawan.v3.GenerateEndDeviceQRCodeRequest) | [`GenerateQRCodeResponse`](#ttn.lorawan.v3.GenerateQRCodeResponse) | Generates a QR code. |
| `Parse` | [`ParseEndDeviceQRCodeRequest`](#ttn.lorawan.v3.ParseEndDeviceQRCodeRequest) | [`ParseEndDeviceQRCodeResponse`](#ttn.lorawan.v3.ParseEndDeviceQRCodeResponse) | Parse QR Codes of known formats and return the information contained within. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `GetFormat` | `GET` | `/api/v3/qr-codes/end-devices/formats/{format_id}` |  |
| `ListFormats` | `GET` | `/api/v3/qr-codes/end-devices/formats` |  |
| `Generate` | `POST` | `/api/v3/qr-codes/end-devices` | `*` |
| `Parse` | `POST` | `/api/v3/qr-codes/end-devices/parse` | `*` |
| `Parse` | `POST` | `/api/v3/qr-codes/end-devices/{format_id}/parse` | `*` |

## <a name="ttn/lorawan/v3/regional.proto">File `ttn/lorawan/v3/regional.proto`</a>

### <a name="ttn.lorawan.v3.ConcentratorConfig">Message `ConcentratorConfig`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channels` | [`ConcentratorConfig.Channel`](#ttn.lorawan.v3.ConcentratorConfig.Channel) | repeated |  |
| `lora_standard_channel` | [`ConcentratorConfig.LoRaStandardChannel`](#ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel) |  |  |
| `fsk_channel` | [`ConcentratorConfig.FSKChannel`](#ttn.lorawan.v3.ConcentratorConfig.FSKChannel) |  |  |
| `lbt` | [`ConcentratorConfig.LBTConfiguration`](#ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration) |  |  |
| `ping_slot` | [`ConcentratorConfig.Channel`](#ttn.lorawan.v3.ConcentratorConfig.Channel) |  |  |
| `radios` | [`GatewayRadio`](#ttn.lorawan.v3.GatewayRadio) | repeated |  |
| `clock_source` | [`uint32`](#uint32) |  |  |

### <a name="ttn.lorawan.v3.ConcentratorConfig.Channel">Message `ConcentratorConfig.Channel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency` | [`uint64`](#uint64) |  | Frequency (Hz). |
| `radio` | [`uint32`](#uint32) |  |  |

### <a name="ttn.lorawan.v3.ConcentratorConfig.FSKChannel">Message `ConcentratorConfig.FSKChannel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency` | [`uint64`](#uint64) |  | Frequency (Hz). |
| `radio` | [`uint32`](#uint32) |  |  |

### <a name="ttn.lorawan.v3.ConcentratorConfig.LBTConfiguration">Message `ConcentratorConfig.LBTConfiguration`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rssi_target` | [`float`](#float) |  | Received signal strength (dBm). |
| `rssi_offset` | [`float`](#float) |  | Received signal strength offset (dBm). |
| `scan_time` | [`google.protobuf.Duration`](#google.protobuf.Duration) |  |  |

### <a name="ttn.lorawan.v3.ConcentratorConfig.LoRaStandardChannel">Message `ConcentratorConfig.LoRaStandardChannel`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frequency` | [`uint64`](#uint64) |  | Frequency (Hz). |
| `radio` | [`uint32`](#uint32) |  |  |
| `bandwidth` | [`uint32`](#uint32) |  | Bandwidth (Hz). |
| `spreading_factor` | [`uint32`](#uint32) |  |  |

## <a name="ttn/lorawan/v3/rights.proto">File `ttn/lorawan/v3/rights.proto`</a>

### <a name="ttn.lorawan.v3.APIKey">Message `APIKey`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [`string`](#string) |  | Immutable and unique public identifier for the API key. Generated by the Access Server. |
| `key` | [`string`](#string) |  | Immutable and unique secret value of the API key. Generated by the Access Server. |
| `name` | [`string`](#string) |  | User-defined (friendly) name for the API key. |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated | Rights that are granted to this API key. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `name` | <p>`string.max_len`: `50`</p> |
| `rights` | <p>`repeated.items.enum.defined_only`: `true`</p> |
| `expires_at` | <p>`timestamp.gt_now`: `true`</p> |

### <a name="ttn.lorawan.v3.APIKeys">Message `APIKeys`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `api_keys` | [`APIKey`](#ttn.lorawan.v3.APIKey) | repeated |  |

### <a name="ttn.lorawan.v3.Collaborator">Message `Collaborator`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `rights` | <p>`repeated.items.enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.Collaborators">Message `Collaborators`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collaborators` | [`Collaborator`](#ttn.lorawan.v3.Collaborator) | repeated |  |

### <a name="ttn.lorawan.v3.GetCollaboratorResponse">Message `GetCollaboratorResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |

### <a name="ttn.lorawan.v3.Rights">Message `Rights`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `rights` | <p>`repeated.items.enum.defined_only`: `true`</p> |

### <a name="ttn.lorawan.v3.Right">Enum `Right`</a>

Right is the enum that defines all the different rights to do something in the network.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `right_invalid` | 0 |  |
| `RIGHT_USER_INFO` | 1 | The right to view user information. |
| `RIGHT_USER_SETTINGS_BASIC` | 2 | The right to edit basic user settings. |
| `RIGHT_USER_SETTINGS_API_KEYS` | 3 | The right to view and edit user API keys. |
| `RIGHT_USER_DELETE` | 4 | The right to delete user account. |
| `RIGHT_USER_AUTHORIZED_CLIENTS` | 5 | The right to view and edit authorized OAuth clients of the user. |
| `RIGHT_USER_APPLICATIONS_LIST` | 6 | The right to list applications the user is a collaborator of. |
| `RIGHT_USER_APPLICATIONS_CREATE` | 7 | The right to create an application under the user account. |
| `RIGHT_USER_GATEWAYS_LIST` | 8 | The right to list gateways the user is a collaborator of. |
| `RIGHT_USER_GATEWAYS_CREATE` | 9 | The right to create a gateway under the account of the user. |
| `RIGHT_USER_CLIENTS_LIST` | 10 | The right to list OAuth clients the user is a collaborator of. |
| `RIGHT_USER_CLIENTS_CREATE` | 11 | The right to create an OAuth client under the account of the user. |
| `RIGHT_USER_ORGANIZATIONS_LIST` | 12 | The right to list organizations the user is a member of. |
| `RIGHT_USER_ORGANIZATIONS_CREATE` | 13 | The right to create an organization under the user account. |
| `RIGHT_USER_NOTIFICATIONS_READ` | 59 | The right to read notifications sent to the user. |
| `RIGHT_USER_ALL` | 14 | The pseudo-right for all (current and future) user rights. |
| `RIGHT_APPLICATION_INFO` | 15 | The right to view application information. |
| `RIGHT_APPLICATION_SETTINGS_BASIC` | 16 | The right to edit basic application settings. |
| `RIGHT_APPLICATION_SETTINGS_API_KEYS` | 17 | The right to view and edit application API keys. |
| `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` | 18 | The right to view and edit application collaborators. |
| `RIGHT_APPLICATION_SETTINGS_PACKAGES` | 56 | The right to view and edit application packages and associations. |
| `RIGHT_APPLICATION_DELETE` | 19 | The right to delete application. |
| `RIGHT_APPLICATION_DEVICES_READ` | 20 | The right to view devices in application. |
| `RIGHT_APPLICATION_DEVICES_WRITE` | 21 | The right to create devices in application. |
| `RIGHT_APPLICATION_DEVICES_READ_KEYS` | 22 | The right to view device keys in application. Note that keys may not be stored in a way that supports viewing them. |
| `RIGHT_APPLICATION_DEVICES_WRITE_KEYS` | 23 | The right to edit device keys in application. |
| `RIGHT_APPLICATION_TRAFFIC_READ` | 24 | The right to read application traffic (uplink and downlink). |
| `RIGHT_APPLICATION_TRAFFIC_UP_WRITE` | 25 | The right to write uplink application traffic. |
| `RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE` | 26 | The right to write downlink application traffic. |
| `RIGHT_APPLICATION_LINK` | 27 | The right to link as Application to a Network Server for traffic exchange, i.e. read uplink and write downlink (API keys only). This right is typically only given to an Application Server. This right implies RIGHT_APPLICATION_INFO, RIGHT_APPLICATION_TRAFFIC_READ, and RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE. |
| `RIGHT_APPLICATION_ALL` | 28 | The pseudo-right for all (current and future) application rights. |
| `RIGHT_CLIENT_ALL` | 29 | The pseudo-right for all (current and future) OAuth client rights. |
| `RIGHT_CLIENT_INFO` | 60 | The right to read client information. |
| `RIGHT_CLIENT_SETTINGS_BASIC` | 61 | The right to edit basic client settings. |
| `RIGHT_CLIENT_SETTINGS_COLLABORATORS` | 62 | The right to view and edit client collaborators. |
| `RIGHT_CLIENT_DELETE` | 63 | The right to delete a client. |
| `RIGHT_GATEWAY_INFO` | 30 | The right to view gateway information. |
| `RIGHT_GATEWAY_SETTINGS_BASIC` | 31 | The right to edit basic gateway settings. |
| `RIGHT_GATEWAY_SETTINGS_API_KEYS` | 32 | The right to view and edit gateway API keys. |
| `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` | 33 | The right to view and edit gateway collaborators. |
| `RIGHT_GATEWAY_DELETE` | 34 | The right to delete gateway. |
| `RIGHT_GATEWAY_TRAFFIC_READ` | 35 | The right to read gateway traffic. |
| `RIGHT_GATEWAY_TRAFFIC_DOWN_WRITE` | 36 | The right to write downlink gateway traffic. |
| `RIGHT_GATEWAY_LINK` | 37 | The right to link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink (API keys only) This right is typically only given to a gateway. This right implies RIGHT_GATEWAY_INFO. |
| `RIGHT_GATEWAY_STATUS_READ` | 38 | The right to view gateway status. |
| `RIGHT_GATEWAY_LOCATION_READ` | 39 | The right to view view gateway location. |
| `RIGHT_GATEWAY_WRITE_SECRETS` | 57 | The right to store secrets associated with this gateway. |
| `RIGHT_GATEWAY_READ_SECRETS` | 58 | The right to retrieve secrets associated with this gateway. |
| `RIGHT_GATEWAY_ALL` | 40 | The pseudo-right for all (current and future) gateway rights. |
| `RIGHT_ORGANIZATION_INFO` | 41 | The right to view organization information. |
| `RIGHT_ORGANIZATION_SETTINGS_BASIC` | 42 | The right to edit basic organization settings. |
| `RIGHT_ORGANIZATION_SETTINGS_API_KEYS` | 43 | The right to view and edit organization API keys. |
| `RIGHT_ORGANIZATION_SETTINGS_MEMBERS` | 44 | The right to view and edit organization members. |
| `RIGHT_ORGANIZATION_DELETE` | 45 | The right to delete organization. |
| `RIGHT_ORGANIZATION_APPLICATIONS_LIST` | 46 | The right to list the applications the organization is a collaborator of. |
| `RIGHT_ORGANIZATION_APPLICATIONS_CREATE` | 47 | The right to create an application under the organization. |
| `RIGHT_ORGANIZATION_GATEWAYS_LIST` | 48 | The right to list the gateways the organization is a collaborator of. |
| `RIGHT_ORGANIZATION_GATEWAYS_CREATE` | 49 | The right to create a gateway under the organization. |
| `RIGHT_ORGANIZATION_CLIENTS_LIST` | 50 | The right to list the OAuth clients the organization is a collaborator of. |
| `RIGHT_ORGANIZATION_CLIENTS_CREATE` | 51 | The right to create an OAuth client under the organization. |
| `RIGHT_ORGANIZATION_ADD_AS_COLLABORATOR` | 52 | The right to add the organization as a collaborator on an existing entity. |
| `RIGHT_ORGANIZATION_ALL` | 53 | The pseudo-right for all (current and future) organization rights. |
| `RIGHT_SEND_INVITES` | 54 | The right to send invites to new users. Note that this is not prefixed with "USER_"; it is not a right on the user entity. |
| `RIGHT_ALL` | 55 | The pseudo-right for all (current and future) possible rights. |

## <a name="ttn/lorawan/v3/search_services.proto">File `ttn/lorawan/v3/search_services.proto`</a>

### <a name="ttn.lorawan.v3.SearchAccountsRequest">Message `SearchAccountsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [`string`](#string) |  |  |
| `only_users` | [`bool`](#bool) |  |  |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `client_ids` | [`ClientIdentifiers`](#ttn.lorawan.v3.ClientIdentifiers) |  |  |
| `gateway_ids` | [`GatewayIdentifiers`](#ttn.lorawan.v3.GatewayIdentifiers) |  |  |
| `organization_ids` | [`OrganizationIdentifiers`](#ttn.lorawan.v3.OrganizationIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `query` | <p>`string.max_len`: `50`</p> |

### <a name="ttn.lorawan.v3.SearchAccountsResponse">Message `SearchAccountsResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account_ids` | [`OrganizationOrUserIdentifiers`](#ttn.lorawan.v3.OrganizationOrUserIdentifiers) | repeated |  |

### <a name="ttn.lorawan.v3.SearchApplicationsRequest">Message `SearchApplicationsRequest`</a>

This message is used for finding applications in the EntityRegistrySearch service.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [`string`](#string) |  | Find applications where the ID, name or description contains this substring. |
| `id_contains` | [`string`](#string) |  | Find applications where the ID contains this substring. |
| `name_contains` | [`string`](#string) |  | Find applications where the name contains this substring. |
| `description_contains` | [`string`](#string) |  | Find applications where the description contains this substring. |
| `attributes_contain` | [`SearchApplicationsRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchApplicationsRequest.AttributesContainEntry) | repeated | Find applications where the given attributes contain these substrings. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted applications. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `query` | <p>`string.max_len`: `50`</p> |
| `id_contains` | <p>`string.max_len`: `50`</p> |
| `name_contains` | <p>`string.max_len`: `50`</p> |
| `description_contains` | <p>`string.max_len`: `50`</p> |
| `attributes_contain` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `50`</p> |
| `order` | <p>`string.in`: `[ application_id -application_id name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SearchApplicationsRequest.AttributesContainEntry">Message `SearchApplicationsRequest.AttributesContainEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.SearchClientsRequest">Message `SearchClientsRequest`</a>

This message is used for finding OAuth clients in the EntityRegistrySearch service.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [`string`](#string) |  | Find OAuth clients where the ID, name or description contains this substring. |
| `id_contains` | [`string`](#string) |  | Find OAuth clients where the ID contains this substring. |
| `name_contains` | [`string`](#string) |  | Find OAuth clients where the name contains this substring. |
| `description_contains` | [`string`](#string) |  | Find OAuth clients where the description contains this substring. |
| `attributes_contain` | [`SearchClientsRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchClientsRequest.AttributesContainEntry) | repeated | Find OAuth clients where the given attributes contain these substrings. |
| `state` | [`State`](#ttn.lorawan.v3.State) | repeated | Find OAuth clients where the state is any of these states. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted OAuth clients. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `query` | <p>`string.max_len`: `50`</p> |
| `id_contains` | <p>`string.max_len`: `50`</p> |
| `name_contains` | <p>`string.max_len`: `50`</p> |
| `description_contains` | <p>`string.max_len`: `50`</p> |
| `attributes_contain` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `50`</p> |
| `state` | <p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `order` | <p>`string.in`: `[ client_id -client_id name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SearchClientsRequest.AttributesContainEntry">Message `SearchClientsRequest.AttributesContainEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.SearchEndDevicesRequest">Message `SearchEndDevicesRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `application_ids` | [`ApplicationIdentifiers`](#ttn.lorawan.v3.ApplicationIdentifiers) |  |  |
| `query` | [`string`](#string) |  | Find end devices where the ID, name, description or EUI contains this substring. |
| `id_contains` | [`string`](#string) |  | Find end devices where the ID contains this substring. |
| `name_contains` | [`string`](#string) |  | Find end devices where the name contains this substring. |
| `description_contains` | [`string`](#string) |  | Find end devices where the description contains this substring. |
| `attributes_contain` | [`SearchEndDevicesRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry) | repeated | Find end devices where the given attributes contain these substrings. |
| `dev_eui_contains` | [`string`](#string) |  | Find end devices where the (hexadecimal) DevEUI contains this substring. |
| `join_eui_contains` | [`string`](#string) |  | Find end devices where the (hexadecimal) JoinEUI contains this substring. |
| `dev_addr_contains` | [`string`](#string) |  | Find end devices where the (hexadecimal) DevAddr contains this substring. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `application_ids` | <p>`message.required`: `true`</p> |
| `query` | <p>`string.max_len`: `50`</p> |
| `id_contains` | <p>`string.max_len`: `50`</p> |
| `name_contains` | <p>`string.max_len`: `50`</p> |
| `description_contains` | <p>`string.max_len`: `50`</p> |
| `attributes_contain` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `50`</p> |
| `dev_eui_contains` | <p>`string.max_len`: `16`</p> |
| `join_eui_contains` | <p>`string.max_len`: `16`</p> |
| `dev_addr_contains` | <p>`string.max_len`: `8`</p> |
| `order` | <p>`string.in`: `[ device_id -device_id join_eui -join_eui dev_eui -dev_eui name -name description -description created_at -created_at last_seen_at -last_seen_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SearchEndDevicesRequest.AttributesContainEntry">Message `SearchEndDevicesRequest.AttributesContainEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.SearchGatewaysRequest">Message `SearchGatewaysRequest`</a>

This message is used for finding gateways in the EntityRegistrySearch service.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [`string`](#string) |  | Find gateways where the ID, name, description or EUI contains this substring. |
| `id_contains` | [`string`](#string) |  | Find gateways where the ID contains this substring. |
| `name_contains` | [`string`](#string) |  | Find gateways where the name contains this substring. |
| `description_contains` | [`string`](#string) |  | Find gateways where the description contains this substring. |
| `attributes_contain` | [`SearchGatewaysRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchGatewaysRequest.AttributesContainEntry) | repeated | Find gateways where the given attributes contain these substrings. |
| `eui_contains` | [`string`](#string) |  | Find gateways where the (hexadecimal) EUI contains this substring. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted gateways. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `query` | <p>`string.max_len`: `50`</p> |
| `id_contains` | <p>`string.max_len`: `50`</p> |
| `name_contains` | <p>`string.max_len`: `50`</p> |
| `description_contains` | <p>`string.max_len`: `50`</p> |
| `attributes_contain` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `50`</p> |
| `eui_contains` | <p>`string.max_len`: `16`</p> |
| `order` | <p>`string.in`: `[ gateway_id -gateway_id gateway_eui -gateway_eui name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SearchGatewaysRequest.AttributesContainEntry">Message `SearchGatewaysRequest.AttributesContainEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.SearchOrganizationsRequest">Message `SearchOrganizationsRequest`</a>

This message is used for finding organizations in the EntityRegistrySearch service.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [`string`](#string) |  | Find organizations where the ID, name or description contains this substring. |
| `id_contains` | [`string`](#string) |  | Find organizations where the ID contains this substring. |
| `name_contains` | [`string`](#string) |  | Find organizations where the name contains this substring. |
| `description_contains` | [`string`](#string) |  | Find organizations where the description contains this substring. |
| `attributes_contain` | [`SearchOrganizationsRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchOrganizationsRequest.AttributesContainEntry) | repeated | Find organizations where the given attributes contain these substrings. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted organizations. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `query` | <p>`string.max_len`: `50`</p> |
| `id_contains` | <p>`string.max_len`: `50`</p> |
| `name_contains` | <p>`string.max_len`: `50`</p> |
| `description_contains` | <p>`string.max_len`: `50`</p> |
| `attributes_contain` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `50`</p> |
| `order` | <p>`string.in`: `[ organization_id -organization_id name -name created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SearchOrganizationsRequest.AttributesContainEntry">Message `SearchOrganizationsRequest.AttributesContainEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.SearchUsersRequest">Message `SearchUsersRequest`</a>

This message is used for finding users in the EntityRegistrySearch service.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [`string`](#string) |  | Find users where the ID, name or description contains this substring. |
| `id_contains` | [`string`](#string) |  | Find users where the ID contains this substring. |
| `name_contains` | [`string`](#string) |  | Find users where the name contains this substring. |
| `description_contains` | [`string`](#string) |  | Find users where the description contains this substring. |
| `attributes_contain` | [`SearchUsersRequest.AttributesContainEntry`](#ttn.lorawan.v3.SearchUsersRequest.AttributesContainEntry) | repeated | Find users where the given attributes contain these substrings. |
| `state` | [`State`](#ttn.lorawan.v3.State) | repeated | Find users where the state is any of these states. |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted users. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `query` | <p>`string.max_len`: `50`</p> |
| `id_contains` | <p>`string.max_len`: `50`</p> |
| `name_contains` | <p>`string.max_len`: `50`</p> |
| `description_contains` | <p>`string.max_len`: `50`</p> |
| `attributes_contain` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `50`</p> |
| `state` | <p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `order` | <p>`string.in`: `[ user_id -user_id name -name primary_email_address -primary_email_address state -state admin -admin created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.SearchUsersRequest.AttributesContainEntry">Message `SearchUsersRequest.AttributesContainEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.EndDeviceRegistrySearch">Service `EndDeviceRegistrySearch`</a>

The EndDeviceRegistrySearch service indexes devices in the EndDeviceRegistry
and enables searching for them.
This service is not implemented on all deployments.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `SearchEndDevices` | [`SearchEndDevicesRequest`](#ttn.lorawan.v3.SearchEndDevicesRequest) | [`EndDevices`](#ttn.lorawan.v3.EndDevices) | Search for end devices in the given application that match the conditions specified in the request. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `SearchEndDevices` | `GET` | `/api/v3/search/applications/{application_ids.application_id}/devices` |  |

### <a name="ttn.lorawan.v3.EntityRegistrySearch">Service `EntityRegistrySearch`</a>

The EntityRegistrySearch service indexes entities in the various registries
and enables searching for them.
This service is not implemented on all deployments.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `SearchApplications` | [`SearchApplicationsRequest`](#ttn.lorawan.v3.SearchApplicationsRequest) | [`Applications`](#ttn.lorawan.v3.Applications) | Search for applications that match the conditions specified in the request. Non-admin users will only match applications that they have rights on. |
| `SearchClients` | [`SearchClientsRequest`](#ttn.lorawan.v3.SearchClientsRequest) | [`Clients`](#ttn.lorawan.v3.Clients) | Search for OAuth clients that match the conditions specified in the request. Non-admin users will only match OAuth clients that they have rights on. |
| `SearchGateways` | [`SearchGatewaysRequest`](#ttn.lorawan.v3.SearchGatewaysRequest) | [`Gateways`](#ttn.lorawan.v3.Gateways) | Search for gateways that match the conditions specified in the request. Non-admin users will only match gateways that they have rights on. |
| `SearchOrganizations` | [`SearchOrganizationsRequest`](#ttn.lorawan.v3.SearchOrganizationsRequest) | [`Organizations`](#ttn.lorawan.v3.Organizations) | Search for organizations that match the conditions specified in the request. Non-admin users will only match organizations that they have rights on. |
| `SearchUsers` | [`SearchUsersRequest`](#ttn.lorawan.v3.SearchUsersRequest) | [`Users`](#ttn.lorawan.v3.Users) | Search for users that match the conditions specified in the request. This is only available to admin users. |
| `SearchAccounts` | [`SearchAccountsRequest`](#ttn.lorawan.v3.SearchAccountsRequest) | [`SearchAccountsResponse`](#ttn.lorawan.v3.SearchAccountsResponse) | Search for accounts that match the conditions specified in the request. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `SearchApplications` | `GET` | `/api/v3/search/applications` |  |
| `SearchClients` | `GET` | `/api/v3/search/clients` |  |
| `SearchGateways` | `GET` | `/api/v3/search/gateways` |  |
| `SearchOrganizations` | `GET` | `/api/v3/search/organizations` |  |
| `SearchUsers` | `GET` | `/api/v3/search/users` |  |
| `SearchAccounts` | `GET` | `/api/v3/search/accounts` |  |
| `SearchAccounts` | `GET` | `/api/v3/applications/{application_ids.application_id}/collaborators/search` |  |
| `SearchAccounts` | `GET` | `/api/v3/clients/{client_ids.client_id}/collaborators/search` |  |
| `SearchAccounts` | `GET` | `/api/v3/gateways/{gateway_ids.gateway_id}/collaborators/search` |  |
| `SearchAccounts` | `GET` | `/api/v3/organizations/{organization_ids.organization_id}/collaborators/search` |  |

## <a name="ttn/lorawan/v3/secrets.proto">File `ttn/lorawan/v3/secrets.proto`</a>

### <a name="ttn.lorawan.v3.Secret">Message `Secret`</a>

Secret contains a secret value. It also contains the ID of the Encryption key used to encrypt it.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key_id` | [`string`](#string) |  | ID of the Key used to encrypt the secret. |
| `value` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `value` | <p>`bytes.max_len`: `2048`</p> |

## <a name="ttn/lorawan/v3/simulate.proto">File `ttn/lorawan/v3/simulate.proto`</a>

### <a name="ttn.lorawan.v3.SimulateDataUplinkParams">Message `SimulateDataUplinkParams`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `dev_addr` | [`bytes`](#bytes) |  |  |
| `f_nwk_s_int_key` | [`bytes`](#bytes) |  |  |
| `s_nwk_s_int_key` | [`bytes`](#bytes) |  |  |
| `nwk_s_enc_key` | [`bytes`](#bytes) |  |  |
| `app_s_key` | [`bytes`](#bytes) |  |  |
| `adr` | [`bool`](#bool) |  |  |
| `adr_ack_req` | [`bool`](#bool) |  |  |
| `confirmed` | [`bool`](#bool) |  |  |
| `ack` | [`bool`](#bool) |  |  |
| `f_cnt` | [`uint32`](#uint32) |  |  |
| `f_port` | [`uint32`](#uint32) |  |  |
| `frm_payload` | [`bytes`](#bytes) |  |  |
| `conf_f_cnt` | [`uint32`](#uint32) |  |  |
| `tx_dr_idx` | [`uint32`](#uint32) |  |  |
| `tx_ch_idx` | [`uint32`](#uint32) |  |  |
| `f_opts` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `dev_addr` | <p>`bytes.len`: `4`</p> |
| `f_nwk_s_int_key` | <p>`bytes.len`: `16`</p> |
| `s_nwk_s_int_key` | <p>`bytes.len`: `16`</p> |
| `nwk_s_enc_key` | <p>`bytes.len`: `16`</p> |
| `app_s_key` | <p>`bytes.len`: `16`</p> |
| `f_opts` | <p>`bytes.max_len`: `15`</p> |

### <a name="ttn.lorawan.v3.SimulateJoinRequestParams">Message `SimulateJoinRequestParams`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `join_eui` | [`bytes`](#bytes) |  |  |
| `dev_eui` | [`bytes`](#bytes) |  |  |
| `dev_nonce` | [`bytes`](#bytes) |  |  |
| `app_key` | [`bytes`](#bytes) |  |  |
| `nwk_key` | [`bytes`](#bytes) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `join_eui` | <p>`bytes.len`: `8`</p> |
| `dev_eui` | <p>`bytes.len`: `8`</p> |
| `dev_nonce` | <p>`bytes.len`: `2`</p> |
| `app_key` | <p>`bytes.len`: `16`</p> |
| `nwk_key` | <p>`bytes.len`: `16`</p> |

### <a name="ttn.lorawan.v3.SimulateMetadataParams">Message `SimulateMetadataParams`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `rssi` | [`float`](#float) |  |  |
| `snr` | [`float`](#float) |  |  |
| `timestamp` | [`uint32`](#uint32) |  |  |
| `time` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `lorawan_version` | [`MACVersion`](#ttn.lorawan.v3.MACVersion) |  |  |
| `lorawan_phy_version` | [`PHYVersion`](#ttn.lorawan.v3.PHYVersion) |  |  |
| `band_id` | [`string`](#string) |  |  |
| `frequency` | [`uint64`](#uint64) |  |  |
| `channel_index` | [`uint32`](#uint32) |  |  |
| `bandwidth` | [`uint32`](#uint32) |  |  |
| `spreading_factor` | [`uint32`](#uint32) |  |  |
| `data_rate_index` | [`uint32`](#uint32) |  |  |

## <a name="ttn/lorawan/v3/user.proto">File `ttn/lorawan/v3/user.proto`</a>

### <a name="ttn.lorawan.v3.CreateLoginTokenRequest">Message `CreateLoginTokenRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `skip_email` | [`bool`](#bool) |  | Skip sending the login token to the user by email. This field is only effective when the login token is created by an admin user. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.CreateLoginTokenResponse">Message `CreateLoginTokenResponse`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token` | [`string`](#string) |  | The token that can be used for logging in as the user. This field is only present if a token was created by an admin user for a non-admin user. |

### <a name="ttn.lorawan.v3.CreateTemporaryPasswordRequest">Message `CreateTemporaryPasswordRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.CreateUserAPIKeyRequest">Message `CreateUserAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `name` | [`string`](#string) |  |  |
| `rights` | [`Right`](#ttn.lorawan.v3.Right) | repeated |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `rights` | <p>`repeated.min_items`: `1`</p><p>`repeated.unique`: `true`</p><p>`repeated.items.enum.defined_only`: `true`</p> |
| `expires_at` | <p>`timestamp.gt_now`: `true`</p> |

### <a name="ttn.lorawan.v3.CreateUserRequest">Message `CreateUserRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user` | [`User`](#ttn.lorawan.v3.User) |  |  |
| `invitation_token` | [`string`](#string) |  | The invitation token that was sent to the user (some networks require an invitation in order to register new users). |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteInvitationRequest">Message `DeleteInvitationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `email` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `email` | <p>`string.email`: `true`</p> |

### <a name="ttn.lorawan.v3.DeleteUserAPIKeyRequest">Message `DeleteUserAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetUserAPIKeyRequest">Message `GetUserAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `key_id` | [`string`](#string) |  | Unique public identifier for the API key. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.GetUserRequest">Message `GetUserRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the user fields that should be returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.Invitation">Message `Invitation`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `email` | [`string`](#string) |  |  |
| `token` | [`string`](#string) |  |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `accepted_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `accepted_by` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `email` | <p>`string.email`: `true`</p> |

### <a name="ttn.lorawan.v3.Invitations">Message `Invitations`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `invitations` | [`Invitation`](#ttn.lorawan.v3.Invitation) | repeated |  |

### <a name="ttn.lorawan.v3.ListInvitationsRequest">Message `ListInvitationsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListUserAPIKeysRequest">Message `ListUserAPIKeysRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path. Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `order` | <p>`string.in`: `[ api_key_id -api_key_id name -name created_at -created_at expires_at -expires_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListUserSessionsRequest">Message `ListUserSessionsRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `order` | <p>`string.in`: `[ created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.ListUsersRequest">Message `ListUsersRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the user fields that should be returned. |
| `order` | [`string`](#string) |  | Order the results by this field path (must be present in the field mask). Default ordering is by ID. Prepend with a minus (-) to reverse the order. |
| `limit` | [`uint32`](#uint32) |  | Limit the number of results per page. |
| `page` | [`uint32`](#uint32) |  | Page number for pagination. 0 is interpreted as 1. |
| `deleted` | [`bool`](#bool) |  | Only return recently deleted users. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `order` | <p>`string.in`: `[ user_id -user_id name -name primary_email_address -primary_email_address state -state admin -admin created_at -created_at]`</p> |
| `limit` | <p>`uint32.lte`: `1000`</p> |

### <a name="ttn.lorawan.v3.LoginToken">Message `LoginToken`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `token` | [`string`](#string) |  |  |
| `used` | [`bool`](#bool) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.SendInvitationRequest">Message `SendInvitationRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `email` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `email` | <p>`string.email`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateUserAPIKeyRequest">Message `UpdateUserAPIKeyRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `api_key` | [`APIKey`](#ttn.lorawan.v3.APIKey) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the api key fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `api_key` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.UpdateUserPasswordRequest">Message `UpdateUserPasswordRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `new` | [`string`](#string) |  |  |
| `old` | [`string`](#string) |  |  |
| `revoke_all_access` | [`bool`](#bool) |  | Revoke active sessions and access tokens of user if true. To be used if credentials are suspected to be compromised. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `new` | <p>`string.max_len`: `1000`</p> |
| `old` | <p>`string.max_len`: `1000`</p> |

### <a name="ttn.lorawan.v3.UpdateUserRequest">Message `UpdateUserRequest`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user` | [`User`](#ttn.lorawan.v3.User) |  |  |
| `field_mask` | [`google.protobuf.FieldMask`](#google.protobuf.FieldMask) |  | The names of the user fields that should be updated. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user` | <p>`message.required`: `true`</p> |

### <a name="ttn.lorawan.v3.User">Message `User`</a>

User is the message that defines a user on the network.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  | The identifiers of the user. These are public and can be seen by any authenticated user in the network. |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the user was created. This information is public and can be seen by any authenticated user in the network. |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the user was last updated. This information is public and can be seen by any authenticated user in the network. |
| `deleted_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the user was deleted. This information is public and can be seen by any authenticated user in the network. |
| `name` | [`string`](#string) |  | The name of the user. This information is public and can be seen by any authenticated user in the network. |
| `description` | [`string`](#string) |  | A description for the user. This information is public and can be seen by any authenticated user in the network. |
| `attributes` | [`User.AttributesEntry`](#ttn.lorawan.v3.User.AttributesEntry) | repeated | Key-value attributes for this users. Typically used for storing integration-specific data. |
| `contact_info` | [`ContactInfo`](#ttn.lorawan.v3.ContactInfo) | repeated | Contact information for this user. Typically used to indicate who to contact with security/billing questions about the user. This field is deprecated. |
| `primary_email_address` | [`string`](#string) |  | Primary email address that can be used for logging in. This address is not public, use contact_info for that. |
| `primary_email_address_validated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  | When the primary email address was validated. Note that email address validation is not required on all networks. |
| `password` | [`string`](#string) |  | The password field is only considered when creating a user. It is not returned on API calls, and can not be updated by updating the User. See the UpdatePassword method of the UserRegistry service for more information. |
| `password_updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `require_password_update` | [`bool`](#bool) |  |  |
| `state` | [`State`](#ttn.lorawan.v3.State) |  | The reviewing state of the user. This information is public and can be seen by any authenticated user in the network. This field can only be modified by admins. |
| `state_description` | [`string`](#string) |  | A description for the state field. This field can only be modified by admins, and should typically only be updated when also updating `state`. |
| `admin` | [`bool`](#bool) |  | This user is an admin. This information is public and can be seen by any authenticated user in the network. This field can only be modified by other admins. |
| `temporary_password` | [`string`](#string) |  | The temporary password can only be used to update a user's password; never returned on API calls. It is not returned on API calls, and can not be updated by updating the User. See the CreateTemporaryPassword method of the UserRegistry service for more information. |
| `temporary_password_created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `temporary_password_expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `profile_picture` | [`Picture`](#ttn.lorawan.v3.Picture) |  | A profile picture for the user. This information is public and can be seen by any authenticated user in the network. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `ids` | <p>`message.required`: `true`</p> |
| `name` | <p>`string.max_len`: `50`</p> |
| `description` | <p>`string.max_len`: `2000`</p> |
| `attributes` | <p>`map.max_pairs`: `10`</p><p>`map.keys.string.max_len`: `36`</p><p>`map.keys.string.pattern`: `^[a-z0-9](?:[-]?[a-z0-9]){2,}$`</p><p>`map.values.string.max_len`: `200`</p> |
| `contact_info` | <p>`repeated.max_items`: `10`</p> |
| `primary_email_address` | <p>`string.email`: `true`</p> |
| `password` | <p>`string.max_len`: `1000`</p> |
| `state` | <p>`enum.defined_only`: `true`</p> |
| `state_description` | <p>`string.max_len`: `128`</p> |
| `temporary_password` | <p>`string.max_len`: `1000`</p> |

### <a name="ttn.lorawan.v3.User.AttributesEntry">Message `User.AttributesEntry`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [`string`](#string) |  |  |
| `value` | [`string`](#string) |  |  |

### <a name="ttn.lorawan.v3.UserSession">Message `UserSession`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `session_id` | [`string`](#string) |  |  |
| `created_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `updated_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `expires_at` | [`google.protobuf.Timestamp`](#google.protobuf.Timestamp) |  |  |
| `session_secret` | [`string`](#string) |  | The session secret is used to compose an authorization key and is never returned. |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `session_id` | <p>`string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.UserSessionIdentifiers">Message `UserSessionIdentifiers`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `user_ids` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) |  |  |
| `session_id` | [`string`](#string) |  |  |

#### Field Rules

| Field | Validations |
| ----- | ----------- |
| `user_ids` | <p>`message.required`: `true`</p> |
| `session_id` | <p>`string.max_len`: `64`</p> |

### <a name="ttn.lorawan.v3.UserSessions">Message `UserSessions`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sessions` | [`UserSession`](#ttn.lorawan.v3.UserSession) | repeated |  |

### <a name="ttn.lorawan.v3.Users">Message `Users`</a>

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `users` | [`User`](#ttn.lorawan.v3.User) | repeated |  |

## <a name="ttn/lorawan/v3/user_services.proto">File `ttn/lorawan/v3/user_services.proto`</a>

### <a name="ttn.lorawan.v3.UserAccess">Service `UserAccess`</a>

The UserAcces service, exposed by the Identity Server, is used to manage
API keys of users.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListRights` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) | [`Rights`](#ttn.lorawan.v3.Rights) | List the rights the caller has on this user. |
| `CreateAPIKey` | [`CreateUserAPIKeyRequest`](#ttn.lorawan.v3.CreateUserAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Create an API key scoped to this user. User API keys can give access to the user itself, as well as any organization, application, gateway and OAuth client this user is a collaborator of. |
| `ListAPIKeys` | [`ListUserAPIKeysRequest`](#ttn.lorawan.v3.ListUserAPIKeysRequest) | [`APIKeys`](#ttn.lorawan.v3.APIKeys) | List the API keys for this user. |
| `GetAPIKey` | [`GetUserAPIKeyRequest`](#ttn.lorawan.v3.GetUserAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Get a single API key of this user. |
| `UpdateAPIKey` | [`UpdateUserAPIKeyRequest`](#ttn.lorawan.v3.UpdateUserAPIKeyRequest) | [`APIKey`](#ttn.lorawan.v3.APIKey) | Update the rights of an API key of the user. This method can also be used to delete the API key, by giving it no rights. The caller is required to have all assigned or/and removed rights. |
| `DeleteAPIKey` | [`DeleteUserAPIKeyRequest`](#ttn.lorawan.v3.DeleteUserAPIKeyRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete a single API key of this user. |
| `CreateLoginToken` | [`CreateLoginTokenRequest`](#ttn.lorawan.v3.CreateLoginTokenRequest) | [`CreateLoginTokenResponse`](#ttn.lorawan.v3.CreateLoginTokenResponse) | Create a login token that can be used for a one-time login as a user. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `ListRights` | `GET` | `/api/v3/users/{user_id}/rights` |  |
| `CreateAPIKey` | `POST` | `/api/v3/users/{user_ids.user_id}/api-keys` | `*` |
| `ListAPIKeys` | `GET` | `/api/v3/users/{user_ids.user_id}/api-keys` |  |
| `GetAPIKey` | `GET` | `/api/v3/users/{user_ids.user_id}/api-keys/{key_id}` |  |
| `UpdateAPIKey` | `PUT` | `/api/v3/users/{user_ids.user_id}/api-keys/{api_key.id}` | `*` |
| `DeleteAPIKey` | `DELETE` | `/api/v3/users/{user_ids.user_id}/api-keys/{key_id}` |  |
| `CreateLoginToken` | `POST` | `/api/v3/users/{user_ids.user_id}/login-tokens` |  |

### <a name="ttn.lorawan.v3.UserInvitationRegistry">Service `UserInvitationRegistry`</a>

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Send` | [`SendInvitationRequest`](#ttn.lorawan.v3.SendInvitationRequest) | [`Invitation`](#ttn.lorawan.v3.Invitation) | Invite a user to join the network. |
| `List` | [`ListInvitationsRequest`](#ttn.lorawan.v3.ListInvitationsRequest) | [`Invitations`](#ttn.lorawan.v3.Invitations) | List the invitations the caller has sent. |
| `Delete` | [`DeleteInvitationRequest`](#ttn.lorawan.v3.DeleteInvitationRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete (revoke) a user invitation. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Send` | `POST` | `/api/v3/invitations` | `*` |
| `List` | `GET` | `/api/v3/invitations` |  |
| `Delete` | `DELETE` | `/api/v3/invitations` |  |

### <a name="ttn.lorawan.v3.UserRegistry">Service `UserRegistry`</a>

The UserRegistry service, exposed by the Identity Server, is used to manage
user registrations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [`CreateUserRequest`](#ttn.lorawan.v3.CreateUserRequest) | [`User`](#ttn.lorawan.v3.User) | Register a new user. This method may be restricted by network settings. |
| `Get` | [`GetUserRequest`](#ttn.lorawan.v3.GetUserRequest) | [`User`](#ttn.lorawan.v3.User) | Get the user with the given identifiers, selecting the fields given by the field mask. The method may return more or less fields, depending on the rights of the caller. |
| `List` | [`ListUsersRequest`](#ttn.lorawan.v3.ListUsersRequest) | [`Users`](#ttn.lorawan.v3.Users) | List users of the network. This method is typically restricted to admins only. |
| `Update` | [`UpdateUserRequest`](#ttn.lorawan.v3.UpdateUserRequest) | [`User`](#ttn.lorawan.v3.User) | Update the user, changing the fields specified by the field mask to the provided values. This method can not be used to change the password, see the UpdatePassword method for that. |
| `CreateTemporaryPassword` | [`CreateTemporaryPasswordRequest`](#ttn.lorawan.v3.CreateTemporaryPasswordRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Create a temporary password that can be used for updating a forgotten password. The generated password is sent to the user's email address. |
| `UpdatePassword` | [`UpdateUserPasswordRequest`](#ttn.lorawan.v3.UpdateUserPasswordRequest) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Update the password of the user. |
| `Delete` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete the user. This may not release the user ID for reuse. |
| `Restore` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Restore a recently deleted user. Deployment configuration may specify if, and for how long after deletion, entities can be restored. |
| `Purge` | [`UserIdentifiers`](#ttn.lorawan.v3.UserIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Purge the user. This will release the user ID for reuse. The user is responsible for clearing data from any (external) integrations that may store and expose data by user or organization ID. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `Create` | `POST` | `/api/v3/users` | `*` |
| `Get` | `GET` | `/api/v3/users/{user_ids.user_id}` |  |
| `List` | `GET` | `/api/v3/users` |  |
| `Update` | `PUT` | `/api/v3/users/{user.ids.user_id}` | `*` |
| `CreateTemporaryPassword` | `POST` | `/api/v3/users/{user_ids.user_id}/temporary_password` |  |
| `UpdatePassword` | `PUT` | `/api/v3/users/{user_ids.user_id}/password` | `*` |
| `Delete` | `DELETE` | `/api/v3/users/{user_id}` |  |
| `Restore` | `POST` | `/api/v3/users/{user_id}/restore` |  |
| `Purge` | `DELETE` | `/api/v3/users/{user_id}/purge` |  |

### <a name="ttn.lorawan.v3.UserSessionRegistry">Service `UserSessionRegistry`</a>

The UserSessionRegistry service, exposed by the Identity Server, is used to manage
(browser) sessions of the user.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `List` | [`ListUserSessionsRequest`](#ttn.lorawan.v3.ListUserSessionsRequest) | [`UserSessions`](#ttn.lorawan.v3.UserSessions) | List the active sessions for the given user. |
| `Delete` | [`UserSessionIdentifiers`](#ttn.lorawan.v3.UserSessionIdentifiers) | [`.google.protobuf.Empty`](#google.protobuf.Empty) | Delete (revoke) the given user session. |

#### HTTP bindings

| Method Name | Method | Pattern | Body |
| ----------- | ------ | ------- | ---- |
| `List` | `GET` | `/api/v3/users/{user_ids.user_id}/sessions` |  |
| `Delete` | `DELETE` | `/api/v3/users/{user_ids.user_id}/sessions/{session_id}` |  |

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
