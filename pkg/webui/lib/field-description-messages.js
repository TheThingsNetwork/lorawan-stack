// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { defineMessages } from 'react-intl'

import GLOSSARY_IDS from '@ttn-lw/lib/constants/glossary-ids'
import TOOLTIP_IDS from '@ttn-lw/lib/constants/tooltip-ids'

import sharedMessages from './shared-messages'

const m = defineMessages({
  idLocation:
    'Enter a value using lowercase letters, numbers, and dashes. You can choose this freely.',
  freqPlanDescription:
    'A frequency plan defines data rates that your end device or gateway is setup to use. It is important that gateways and end devices within reach use the same frequency plan to be able to communicate.',
  freqPlanLocation:
    'Your end device or gateway manufacturer should provide information about the applicable frequency plan for a particular device. In some cases they are printed on the device itself but they should always be in the hardware manual or data sheet.',
  freqPlanAbsence:
    'Contact the manufacturer or reseller. Using an incorrect frequency plan will prevent traffic between devices.',
  devEuiDescription: 'A 64 bit extended unique identifier for your end device.',
  devEuiLocation:
    'It should be provided to you by the manufacturer, or printed on the end device packaging.',
  joinEuiDescription:
    'The JoinEUI (formerly called AppEUI) is a 64 bit extended unique identifier used to identify the Join Server during activation.',
  joinEuiLocation:
    'It should be provided by the end device manufacturer for pre-provisioned end devices, or by the owner of the Join Server you will use.',
  joinEuiAbsence:
    'Contact the manufacturer or your reseller. If they can not provide a JoinEUI, and your end device is programmable, it is okay to use all-zeros, but ensure that you use the same JoinEUI in your end device as you enter in The Things Stack.',
  appKeyDescription:
    'An end device specific encryption key used during OTAA to derive the AppSKey (in LoRaWAN 1.1x) or both the NwkSKey and AppSKey in LoRaWAN 1.0x.',
  appKeyLocation:
    'It is usually pre-provisioned by the end device manufacturer, but can also be created by the user.',
  nwkKeyDescription:
    'An end device specific encryption key used to derive the FNwkSIntKey, SNwkSIntKey, NwkSEncKey in LoRaWAN 1.1. When a LoRaWAN 1.1 capable device connects to a LoRaWAN 1.0x Network Server which does not support dual root keys (NwkKey and AppKey), the NwkKey value is used as the AppKey value.',
  nwkKeyLocation:
    'It is usually pre-provisioned by the end device manufacturer, but some end devices also allow using a user-defined value.',
  devIdDescription:
    'A mandatory identifier for your end device that must be unique within the application and cannot be changed after creation. It is used to reference your end device e.g. in events, webhooks and API requests.',
  joinServerDescription:
    "The Join Server's role is to store root keys, generate session keys, and to send those securely to the Network Server and Application Server of choice. The device contains the same root keys, which can be provisioned as part of assembly, distribution or upon installation.",
  joinServerLocation:
    'Contact your manufacturer or reseller to find out if your end device is pre-provisioned on an external Join Server. If not, you may use the local Join Server, or provision the end device on our Global Join Server so that you can transfer it without lock-in.',
  joinServerAbsence:
    'If the end device is pre-provisioned, you will need the keys from the manufacturer to activate it.',
  devAddrDescription:
    'A 32 bit non-unique identifier, assigned by the Network Server during end device activation.',
  devAddrLocation:
    'For Activation-By-Personalization (ABP), you must generate a device address and manually program it into the end device. You can use the generate button next to the input field to generate a device address.',
  appSKeyDescription:
    'After activation, this encryption key is used to secure messages which carry a payload.',
  appSKeyLocation:
    'For Activation-By-Personalization (ABP), you must generate a key and manually program it into your end device. You can use the generate button next to the input field to generate a key.',
  nwkSKeyDescription:
    'After activation, this encryption key is used to secure messages which do not carry a payload.',
  nwkSKeyLocation:
    'For OTAA, it is created by the Network Server. If using ABP, you must create one and manually enter it in the end device and The Things Stack.',
  loraCloudModemEncodingDescription:
    'The Application Server will parse streaming fragments using the TLV encoding and resubmit it to the Modem Services as a GNSS or WiFi payload, depending on the tag.',
  lwVersionDescription:
    'The LoRa Alliance LoRaWAN specification your end device conforms to, which defines which Media Access Control (MAC) features it supports.',
  lwVersionLocation:
    'The LoRaWAN version for your end device should be provided by the manufacturer in a datasheet as LoRaWAN version or LoRaWAN specification',
  lwVersionAbsence:
    'Contact your manufacturer or reseller, since specifying the wrong version can lead to complex issues when the Network Server provides the end device with unsupported configuration (MAC) commands.',
  regParameterDescription:
    'The Regional Parameters specify frequency, dwell time, and other communication settings for different geographical areas. The Regional Parameters version is the version of the LoRa Alliance specification which your device supports.',
  regParameterLocation:
    'The Regional Parameters version should be provided by the end device manufacturer in a datasheet.',
  regParameterAbsence:
    'Contact your manufacturer or reseller to obtain the correct specification. Specifying a wrong version can lead to complex issues when the Network Server provides the end device with unsupported configuration (MAC) commands.',
  classDescription:
    'The LoRaWAN specification defines three end device types. All LoRaWAN end devices must implement Class A, whereas Class B and Class C are extensions to the specification of Class A devices that specify different downlink reception behavior.',
  classLocation:
    'The class capabilities of your end device should be provided by the manufaturer in a datasheet.',
  classAbsence:
    'If your end device will not receive downlink messages, you can safely use Class A, which all LoRaWAN certified devices must implement. Otherwise contact your manufacturer or reseller.',
  rx1DataRateOffsetDescription:
    'The Data Rate Offset sets the offset between the uplink data rate and the downlink data rate used to communicate with the End Device during the first reception slot (RX1).',
  deviceBrandDescription: 'This is the manufacturer of your end device.',
  deviceModelDescription: 'The particular model of your end device.',
  deviceHardwareVersionDescription: 'The hardware version of your device.',
  deviceHardwareVersionLocation:
    'It should be provided by the manufacturer of your device, or printed on the device packaging.',
  deviceFirmwareVersionDescription: 'The version of firmware loaded on your device.',
  deviceFirmwareVersionLocation:
    'The firmware version should be provided by the manufacturer of your device, or printed on the device packaging. It may be possible to upgrade your device firmware to a known version.',
  activationModeDescription:
    'OTAA is the preferred and most secure way to connect a device. Devices perform a join-procedure with the network. ABP requires hardcoding the device address and security keys. Multicast is a virtual group of ABP devices which allows all devices to receive the same downlinks. Multicast groups do not support uplinks.',
  activationModeLocation: 'You decide how to activate your devices. Whenever possible, use OTAA.',
  deviceNameDescription: 'An optional human readable name to help you identify your device.',
  deviceDescDescription:
    'An optional description, which can also be used to save notes about the end device.',
  frameCounterWidthDescription:
    'Most devices use a 32 bit frame counter to prevent replay attacks. Devices with extremely limited resources are permitted to use 16 bit counters.',
  frameCounterWidthLocation: 'It should be provided by the device manufacturer.',
  frameCounterWidthAbsence:
    'Contact your manufacturer or reseller. Most devices use 32 bit counters. Selecting the wrong value will produce errors once the Up or Down frame counter exceeds 16 bits and rolls over.',
  rx2DataRateIndexDescription:
    'The data rate used for the second reception window used by this end device to receive downlinks.',
  rx2FrequencyDescription:
    'The frequency used for the second reception window used by this end device to receive downlinks.',
  gatewayIdDescription:
    'A mandatory identifier for your gateway that must be unique per network and cannot be changed after creation. It is used to reference your end device e.g. in events, webhooks and API requests.',
  gatewayEuiDescription: 'A 64 bit extended unique identifier for your gateway.',
  gatewayEuiLocation:
    'It should be provided to you by the manufacturer, or printed on the gateway packaging.',
  gatewayEuiAbsence:
    'Some gateways do not use EUIs. In that case, you can continue without EUI. If you are unsure, we recommend contacting the manufacturer or reseller.',
  gatewayNameDescription: 'An optional human readable name to help you identify your gateway.',
  gatewayDescDescription:
    'An optional description, which can also be used to save notes about the gateway.',
  requireAuthenticatedConnectionDescription:
    'This will only allow a gateway to connect if it uses a TLS enabled Basic Station or MQTT connection. It will not allow connections from UDP packet forwarders.',
  gatewayStatusDescription:
    'Setting your gateway status to public allows status information about the gateway to be shared with other users in the network, and with Packet Broker if enabled by the network operator.',
  gatewayGenerateApiKeyCUPS:
    'Use this option if you plan to use your gateway with LoRa Basics™ Station CUPS (Configuration and Update Server) to check for configuration and software updates. When checked, an appropriate API key for the CUPS service is automatically generated, so you can authorize the gateway right away.',
  gatewayGenerateApiKeyLNS:
    'Use this option if you plan to use your gateway with LoRa Basics™ Station. LNS is used to establish a connection between your gateway and LoRa Basics™ Station. If checked, an appropriate API key for LNS is automatically generated, so you can authorize the gateway for LNS usage right away.',
  gatewayStatusPublicDescription:
    "If this option is checked, the gateway status can be retrieved by other network participants as well as by peering network participants through Packet Broker (if enabled). This is useful if you would like others to benefit from your gateway's coverage by providing them with useful information to do so.",
  gatewayLocationPublicDescription:
    "If this option is checked, the gateway location can be retrieved by other network participants as well as by peering network participants through Packet Broker (if enabled). This is useful if you would like others to benefit from your gateway's coverage by providing them with the exact location.",
  gatewayAttributesDescription:
    'Attributes can be used to set arbitrary information about the entity, to be used by scripts, or simply for your own organization.',
  scheduleDownlinkLateDescription:
    'This legacy feature enables buffering of downlink messages on the network server, for gateways with no downlink queue. Scheduling consecutive downlinks on gateways with no queue will cause only the most recent downlink to be stored.',
  enforceDutyCycleDescription:
    'When checked, the Network Server will only schedule messages respecting the duty cycle limitations of the selected frequency plan.  Note that you are required by law to respect duty cycle regulations applicable to the physical location of your end device.',
  scheduleAnytimeDelayDescription:
    'Adjust the time that the Network Server schedules class C messages in advance. This is useful for gateways that have a known high latency backhaul, like 3G and satellite.',
  updateGtwLocationFromStatusDescription:
    'Instead of setting the location manually, you can alternatively choose to update the location of this gateway from status messages. This only works for gateways that send their locations within status messages while using an authenticated connection; gateways connected over UDP are not supported. Please refer to the manual of your gateway model to see whether sending location data is supported.',
  claimAuthenticationCodeDescription:
    'The claim authentication code is the proof of ownership of the gateway.',
  claimAuthenticationCodeLocation:
    'It is typically printed on the box or on the device. For example, for The Things Indoor Gateway, the claim authentication code is the WiFi password printed on the device.',
  disablePacketBrokerForwardingDescription:
    'When checked, uplink messages received from this gateway will not be forwarded to Packet Broker. This option takes effect only after the gateway reconnects.',
  rx1DelayDescription:
    'The amount of time in seconds after an uplink that the downlink window opens.',
  rx1DelayAbsence:
    'The default value of 5 seconds will be set by The Things Stack to accommodate for high-latency backhauls and/or Packet Broker.',
  pingSlotPeriodicityDescription: 'The amount of time between two receive windows (ping slots).',
  classBTimeoutDescription:
    'The amount of time after which the network server will assume a message is lost, if not confirmed. This should be set to a value less than the time between two ping slots (ping slot periodicity).',
  rx2DataRateDescription:
    'The data rate used for the RX2 window. For OTAA devices, this is configured as part of join. For ABP devices, a matching value must be programmed in the device.',
  networkRxDefaultsDescription:
    'The network uses a set of default MAC settings (e.g. Rx delays, data rates and frequencies) for the end device. These are based on the recommendations made for the respective band. In most cases these defaults will be correct for your setup. If you wish to use different settings, you can uncheck this checkbox and use custom values.',
  skipJoinServerRegistration:
    'For testing purposes, you can opt to skip registration of this end device on the Join Server. Do not enable this option if you do not understand the implications, as it will affect join capabilities of the end device.',
  factoryPresetFreqDescription:
    'Factory preset frequencies are hard-coded channel frequencies to provide for the Network Server when the end device uses frequencies that divert from the defaults of the band specification. This is uncommon but can be the case for some special devices.',
  factoryPresetFreqLocation:
    'If your device uses non-default channel frequencies, this information is likely found in the specification sheet or manual of your end device. Otherwise please contact your manufacturer or reseller.',
  factoryPresetFreqAbsence:
    'If your device uses non-default channel frequencies and these frequencies are not passed to the Network Server, the messages sent between the network and the end device on such non-default frequencies are likely to drop.',
  classCTimeoutDescription:
    'The class C timeout determines how long the network will wait for a response for downlinks that require confirmation (e.g. confirmed downlink messages or MAC commands). During that period no other downlinks will be sent.',
  classCTimeoutLocation:
    'This is dependant on your exact use case. It can help to consider the following implications: timeouts that are too long can be prone to blocking communication when replies are not sent immediately. On the other hand, timeouts that are too short might miss the reply when it arrives later.',
  classCTimeoutAbsence:
    'Leave this field empty to let the network apply a default value that should be applicable for most use cases.',
  gatewayPlacementDescription:
    'Informs whether the gateway antenna is placed inside or outside. This information is used solely for display purposes, e.g. on public gateway maps (if your gateway is set as public) and has no technical or functional effects.',
  setClaimAuthCodeDescription:
    'The claim authentication code is used to transfer ownership of an end device using a process called end device claiming. Checking this checkbox will cause a claim authentication code being generated and set for each imported end device so you can eventually transfer the ownership.',
  pingSlotDataRateDescription:
    'The class B ping slot uses a fixed frequency and data rate. This value configures the data rate to use in class B ping slots.',
  beaconFrequencyDescription:
    'The class B beacon is sent on a fixed frequency. This value changes the frequency to use in class B beacons.',
  pingSlotFrequencyDescription:
    'The class B ping slot uses a fixed frequency and data rate. This value configures the frequency to use in class B ping slots.',
  resetsFCntDescription:
    'Enable this to allow end devices to reset their frame counter (FCnt). This is not LoRaWAN compliant and this is not secure. This is to support ABP end devices that reset their frame counter on a power cycle. Do not use this setting in production.',
  resetMacDescription:
    'Resetting the session context and MAC state will reset the end device to its initial (factory) state. This includes resetting the frame counters and any other persisted MAC setting on the end device. Over The Air Activation (OTAA) end devices will also lose their session context. This means that such devices must rejoin the network to continue operation. Activation-by-personalization (ABP) end devices will only reset the MAC state, while preserving up/downlink queues.',
  useAdrDescription:
    'Controls whether the end device uses adaptive data rate. This will allow the network to adjust the employed data rate based on signal to noise ratio. This adaptively optimizes energy consumption, bandwidth and transmission power.',
  adrMarginDescription:
    'Signal-to-noise ratio (SNR) margin in dB that is taken into account in the Adaptive Data Rate (ADR) algorithm to optimize the data rate of the end device. A higher margin requires the end device to have a better SNR before the Network Server instructs a higher data rate for the end device to use.',
  adrAckLimitDescription: 'This value changes the limit value defining the ADR back-off algorithm.',
  adrAckDelayDescription: 'This value changes the delay value defining the ADR back-off algorithm.',
  maxDutyCycleDescription:
    'The maximum aggregated transmit duty cycle of the end device over all sub-bands. The allowed time-on-air is 1/N where N is the given value: 1 is 100%, 1024 is 0.97%, etc. This value is used for traffic shaping. All end devices must respect regional regulations regardless of this value.',
  statusCountPeriodicityDescription:
    'Number of uplink messages after which the end device status is requested. Set to 0 to disable requesting the device status based on the number of uplink messages.',
  statusTimePeriodicityDescription:
    'Interval to request the end device status. Set to 0 to disable requesting the end device status on an interval.',
  skipPayloadCryptoOverrideDescription:
    'Skip payload crypto disables the application layer encryption of LoRaWAN frames. This causes the Application Server to forward the messages without any processing, such as payload formatters, to the integrations.  When doing so, the integrations are responsible for decryption and processing of the binary format in order to understand the message. This application-wide setting can be overwritten per end device.',
  basicAuthDescription:
    'To increase access security, you can choose to generate a "basic auth" authorization header to be attached to the webhook requests, if the target server requires doing so. This will authenticate the webhook requests with the defined credentials.',
  downlinkQueueInvalidated:
    'This occurs only when using LoRaWAN 1.0.x because the frame counters for the application downlinks and network downlinks are shared. Network downlinks increment this frame counter, thus rendering the queued downlinks invalid (since you cannot send two messages with the same frame counter). The Application Server will automatically handle this message as long as "Skip Payload Crypto" is not enabled.',
  filterEventDataDescription:
    'By default, the data pushed to your webhook contains a vast variety details and metadata related to the event. To avoid noise and to save bandwidth, you can filter the event data by specific paths, e.g. `up.uplink_message.decoded_payload`. Your webhook will then only receive the event data that passed the filter.',
  inputMethodDescription:
    'To register the device, we need to know the exact LoRaWAN specifications that the device adheres to. To do that, you can select your device in our extensive LoRaWAN Device Repository, which is the best way to ensure proper configuration. If your device is not listed in the repository, you can also provide versions and MAC configurations manually. Please refer to the manual and/or data sheet of your device and contact your manufacturer or reseller if you are unsure about a certain config.',
  resetsJoinNoncesDescription:
    'Allowing join nonces to be reset disables any reuse checks for the device nonces and join nonces. The join requests can be replayed indefinitely when this option is enabled. This behavior is non compliant with the LoRaWAN specifications and must not be used outside of development environments.',
  resetUsedDevNoncesDescription:
    'The device nonces ensure that join requests cannot be replayed by attackers. Resetting the device nonces enables the end device to re-use a previously used nonce. Do not use this option unless you are sure that you would like the nonces to be usable again.',
  alcsyncDescription:
    'The Application Layer Clock Synchronization package is part of the LoRa TS003 specification, it synchronizes the real-time clock of an end-device to the network’s Global Positioning System (GPS) clock with near-second accuracy. It is useful for end-devices that do not have access to another accurate time source.',
  useDefaultNbTransDescription:
    'The number of retransmissions (NbTrans) controls how many times a frame will be transmitted over the air. The redundancy introduced by retransmissions improves the chances that a packet will be received, at the expense of more power usage. By default, depending on the number of missed frames, the same frame may be transmitted 3 times.',
  dataRateSpecificOverridesDescription:
    'Data rate specific overrides allow the number of transmissions to be limited on a per data rate basis. This may be used to limit power usage for low data rates.',
})

const descriptions = Object.freeze({
  [TOOLTIP_IDS.FREQUENCY_PLAN]: {
    description: m.freqPlanDescription,
    location: m.freqPlanLocation,
    absence: m.freqPlanAbsence,
    glossaryId: GLOSSARY_IDS.FREQUENCY_PLAN,
  },
  [TOOLTIP_IDS.DEV_EUI]: {
    description: m.devEuiDescription,
    location: m.devEuiLocation,
    absence: sharedMessages.absenceContactManufacturer,
    glossaryId: GLOSSARY_IDS.DEV_EUI,
  },
  [TOOLTIP_IDS.JOIN_EUI]: {
    description: m.joinEuiDescription,
    location: m.joinEuiLocation,
    absence: m.joinEuiAbsence,
    glossaryId: GLOSSARY_IDS.JOIN_EUI,
  },
  [TOOLTIP_IDS.APP_KEY]: {
    description: m.appKeyDescription,
    location: m.appKeyLocation,
    absence: sharedMessages.absenceContactManufacturer,
    glossaryId: GLOSSARY_IDS.APP_KEY,
  },
  [TOOLTIP_IDS.NETWORK_KEY]: {
    description: m.nwkKeyDescription,
    location: m.nwkKeyLocation,
    absence: sharedMessages.absenceContactManufacturer,
    glossaryId: GLOSSARY_IDS.NETWORK_KEY,
  },
  [TOOLTIP_IDS.DEVICE_ID]: {
    description: m.devIdDescription,
    location: m.idLocation,
    glossaryId: GLOSSARY_IDS.DEVICE_ID,
  },
  [TOOLTIP_IDS.JOIN_SERVER]: {
    description: m.joinServerDescription,
    location: m.joinServerLocation,
    absence: m.joinServerAbsence,
    glossaryId: GLOSSARY_IDS.JOIN_SERVER,
  },
  [TOOLTIP_IDS.DEVICE_ADDRESS]: {
    description: m.devAddrDescription,
    location: m.devAddrLocation,
    glossaryId: GLOSSARY_IDS.DEVICE_ADDRESS,
  },
  [TOOLTIP_IDS.APP_SESSION_KEY]: {
    description: m.appSKeyDescription,
    location: m.appSKeyLocation,
    glossaryId: GLOSSARY_IDS.APP_SESSION_KEY,
  },
  [TOOLTIP_IDS.NETWORK_SESSION_KEY]: {
    description: m.nwkSKeyDescription,
    location: m.nwkSKeyLocation,
    glossaryId: GLOSSARY_IDS.NETWORK_SESSION_KEY,
  },
  [TOOLTIP_IDS.LORAWAN_VERSION]: {
    description: m.lwVersionDescription,
    location: m.lwVersionLocation,
    absence: m.lwVersionAbsence,
    glossaryId: GLOSSARY_IDS.LORAWAN_VERSION,
  },
  [TOOLTIP_IDS.REGIONAL_PARAMETERS]: {
    description: m.regParameterDescription,
    location: m.regParameterLocation,
    absence: m.regParameterAbsence,
    glossaryId: GLOSSARY_IDS.REGIONAL_PARAMETERS,
  },
  [TOOLTIP_IDS.CLASSES]: {
    description: m.classDescription,
    location: m.classLocation,
    absence: m.classAbsence,
    glossaryId: GLOSSARY_IDS.CLASSES,
  },
  [TOOLTIP_IDS.DATA_RATE_OFFSET]: {
    description: m.rx1DataRateOffsetDescription,
    glossaryId: GLOSSARY_IDS.DATA_RATE_OFFSET,
  },
  [TOOLTIP_IDS.DEVICE_BRAND]: {
    description: m.deviceBrandDescription,
  },
  [TOOLTIP_IDS.DEVICE_MODEL]: {
    description: m.deviceModelDescription,
  },
  [TOOLTIP_IDS.DEVICE_HARDWARE_VERSION]: {
    description: m.deviceHardwareVersionDescription,
    location: m.deviceHardwareVersionLocation,
    absence: sharedMessages.absenceContactManufacturer,
  },
  [TOOLTIP_IDS.DEVICE_FIRMWARE_VERSION]: {
    description: m.deviceFirmwareVersionDescription,
    location: m.deviceFirmwareVersionLocation,
    absence: sharedMessages.absenceContactManufacturer,
  },
  [TOOLTIP_IDS.ACTIVATION_MODE]: {
    description: m.activationModeDescription,
    location: m.activationModeLocation,
    glossaryId: GLOSSARY_IDS.ACTIVATION_MODE,
  },
  [TOOLTIP_IDS.DEVICE_NAME]: {
    description: m.deviceNameDescription,
    location: m.deviceNameLocation,
  },
  [TOOLTIP_IDS.DEVICE_DESCRIPTION]: {
    description: m.deviceDescDescription,
  },
  [TOOLTIP_IDS.FRAME_COUNTER_WIDTH]: {
    description: m.frameCounterWidthDescription,
    location: m.frameCounterWidthLocation,
    absence: m.frameCounterWidthAbsence,
  },
  [TOOLTIP_IDS.RX2_DATA_RATE_INDEX]: {
    description: m.rx2DataRateIndexDescription,
    location: m.rx2DataRateIndexLocation,
    absence: sharedMessages.absenceContactManufacturer,
  },
  [TOOLTIP_IDS.RX2_FREQUENCY]: {
    description: m.rx2FrequencyDescription,
    location: m.rx2FrequencyLocation,
    absence: sharedMessages.absenceContactManufacturer,
  },
  [TOOLTIP_IDS.GATEWAY_ID]: {
    description: m.gatewayIdDescription,
    location: m.idLocation,
    glossaryId: GLOSSARY_IDS.GATEWAY_ID,
  },
  [TOOLTIP_IDS.GATEWAY_EUI]: {
    description: m.gatewayEuiDescription,
    location: m.gatewayEuiLocation,
    absence: m.gatewayEuiAbsence,
    glossaryId: GLOSSARY_IDS.GATEWAY_EUI,
  },
  [TOOLTIP_IDS.GATEWAY_NAME]: {
    description: m.gatewayNameDescription,
    location: m.gatewayNameLocation,
  },
  [TOOLTIP_IDS.GATEWAY_DESCRIPTION]: {
    description: m.gatewayDescDescription,
  },
  [TOOLTIP_IDS.GATEWAY_GENERATE_API_KEY_CUPS]: {
    description: m.gatewayGenerateApiKeyCUPS,
  },
  [TOOLTIP_IDS.GATEWAY_GENERATE_API_KEY_LNS]: {
    description: m.gatewayGenerateApiKeyLNS,
  },
  [TOOLTIP_IDS.REQUIRE_AUTHENTICATED_CONNECTION]: {
    description: m.requireAuthenticatedConnectionDescription,
  },
  [TOOLTIP_IDS.GATEWAY_STATUS_PUBLIC]: {
    description: m.gatewayStatusPublicDescription,
  },
  [TOOLTIP_IDS.GATEWAY_LOCATION_PUBLIC]: {
    description: m.gatewayLocationPublicDescription,
  },
  [TOOLTIP_IDS.GATEWAY_ATTRIBUTES]: {
    description: m.gatewayAttributesDescription,
  },
  [TOOLTIP_IDS.SCHEDULE_DOWNLINK_LATE]: {
    description: m.scheduleDownlinkLateDescription,
  },
  [TOOLTIP_IDS.ENFORCE_DUTY_CYCLE]: {
    description: m.enforceDutyCycleDescription,
    glossaryId: GLOSSARY_IDS.ENFORCE_DUTY_CYCLE,
  },
  [TOOLTIP_IDS.SCHEDULE_ANYTIME_DELAY]: {
    description: m.scheduleAnytimeDelayDescription,
  },
  [TOOLTIP_IDS.UPDATE_LOCATION_FROM_STATUS]: {
    description: m.updateGtwLocationFromStatusDescription,
  },
  [TOOLTIP_IDS.CLAIM_AUTH_CODE]: {
    description: m.claimAuthenticationCodeDescription,
    location: m.claimAuthenticationCodeLocation,
  },
  [TOOLTIP_IDS.DISABLE_PACKET_BROKER_FORWARDING]: {
    description: m.disablePacketBrokerForwardingDescription,
  },
  [TOOLTIP_IDS.RX1_DELAY]: {
    description: m.rx1DelayDescription,
    absence: m.rx1DelayAbsence,
  },
  [TOOLTIP_IDS.PING_SLOT_PERIODICITY]: {
    description: m.pingSlotPeriodicityDescription,
  },
  [TOOLTIP_IDS.CLASS_B_TIMEOUT]: {
    description: m.classBTimeoutDescription,
  },
  [TOOLTIP_IDS.NETWORK_RX_DEFAULTS]: {
    description: m.networkRxDefaultsDescription,
  },
  [TOOLTIP_IDS.SKIP_JOIN_SERVER_REGISTRATION]: {
    description: m.skipJoinServerRegistration,
  },
  [TOOLTIP_IDS.FACTORY_PRESET_FREQUENCIES]: {
    description: m.factoryPresetFreqDescription,
    location: m.factoryPresetFreqLocation,
    absence: m.factoryPresetFreqAbsence,
  },
  [TOOLTIP_IDS.CLASS_C_TIMEOUT]: {
    description: m.classCTimeoutDescription,
    location: m.classCTimeoutLocation,
    absence: m.classCTimeoutAbsence,
  },
  [TOOLTIP_IDS.GATEWAY_PLACEMENT]: {
    description: m.gatewayPlacementDescription,
  },
  [TOOLTIP_IDS.SET_CLAIM_AUTH_CODE]: {
    description: m.setClaimAuthCodeDescription,
  },
  [TOOLTIP_IDS.PING_SLOT_DATA_RATE_INDEX]: {
    description: m.pingSlotDataRateDescription,
  },
  [TOOLTIP_IDS.BEACON_FREQUENCY]: {
    description: m.beaconFrequencyDescription,
  },
  [TOOLTIP_IDS.PING_SLOT_FREQUENCY]: {
    description: m.pingSlotFrequencyDescription,
  },
  [TOOLTIP_IDS.RESETS_F_CNT]: {
    description: m.resetsFCntDescription,
  },
  [TOOLTIP_IDS.RESET_MAC]: {
    description: m.resetMacDescription,
  },
  [TOOLTIP_IDS.ADR_USE]: {
    description: m.useAdrDescription,
  },
  [TOOLTIP_IDS.ADR_MARGIN]: {
    description: m.adrMarginDescription,
  },
  [TOOLTIP_IDS.ADR_ACK_DELAY]: {
    description: m.adrAckDelayDescription,
  },
  [TOOLTIP_IDS.ADR_ACK_LIMIT]: {
    description: m.adrAckLimitDescription,
  },
  [TOOLTIP_IDS.MAX_DUTY_CYCLE]: {
    description: m.maxDutyCycleDescription,
  },
  [TOOLTIP_IDS.STATUS_COUNT_PERIODICITY]: {
    description: m.statusCountPeriodicityDescription,
  },
  [TOOLTIP_IDS.STATUS_TIME_PERIODICITY]: {
    description: m.statusTimePeriodicityDescription,
  },
  [TOOLTIP_IDS.SKIP_PAYLOAD_CRYPTO_OVERRIDE]: {
    description: m.skipPayloadCryptoOverrideDescription,
  },
  [TOOLTIP_IDS.BASIC_AUTH]: {
    description: m.basicAuthDescription,
  },
  [TOOLTIP_IDS.DOWNLINK_QUEUE_INVALIDATED]: {
    description: m.downlinkQueueInvalidated,
  },
  [TOOLTIP_IDS.FILTER_EVENT_DATA]: {
    description: m.filterEventDataDescription,
  },
  [TOOLTIP_IDS.LORA_CLOUD_MODEM_ENCODING]: {
    description: m.loraCloudModemEncodingDescription,
  },
  [TOOLTIP_IDS.INPUT_METHOD]: {
    description: m.inputMethodDescription,
  },
  [TOOLTIP_IDS.RESETS_JOIN_NONCES]: {
    description: m.resetsJoinNoncesDescription,
  },
  [TOOLTIP_IDS.RESET_USED_DEV_NONCES]: {
    description: m.resetUsedDevNoncesDescription,
  },
  [TOOLTIP_IDS.ALCSYNC]: {
    description: m.alcsyncDescription,
  },
  [TOOLTIP_IDS.USE_DEFAULT_NB_TRANS]: {
    description: m.useDefaultNbTransDescription,
  },
  [TOOLTIP_IDS.DATA_RATE_SPECIFIC_OVERRIDES]: {
    description: m.dataRateSpecificOverridesDescription,
  },
  [TOOLTIP_IDS.GATEWAY_SHOW_PROFILES]: {},
})

const links = Object.freeze({
  [TOOLTIP_IDS.FREQUENCY_PLAN]: {
    documentationPath: '/reference/frequency-plans',
  },
  [TOOLTIP_IDS.GATEWAY_ID]: {
    documentationPath: '/reference/id-eui-constraints',
  },
  [TOOLTIP_IDS.GATEWAY_EUI]: {
    documentationPath: '/reference/id-eui-constraints',
  },
  [TOOLTIP_IDS.GATEWAY_GENERATE_API_KEY_CUPS]: {
    documentationPath: '/gateways/lora-basics-station/cups/',
  },
  [TOOLTIP_IDS.GATEWAY_GENERATE_API_KEY_LNS]: {
    documentationPath: '/gateways/lora-basics-station/lns/',
  },
  [TOOLTIP_IDS.DEVICE_ID]: {
    documentationPath: '/reference/id-eui-constraints',
  },
  [TOOLTIP_IDS.DEV_EUI]: {
    documentationPath: '/reference/id-eui-constraints',
  },
  [TOOLTIP_IDS.APPLICATION_ID]: {
    documentationPath: '/reference/id-eui-constraints',
  },
  [TOOLTIP_IDS.ACTIVATION_MODE]: {
    documentationPath: '/devices/abp-vs-otaa',
  },
  [TOOLTIP_IDS.SET_CLAIM_AUTH_CODE]: {
    documentationPath: '/devices/device-claiming',
  },
  [TOOLTIP_IDS.ADR_USE]: {
    documentationPath: '/reference/adr',
  },
  [TOOLTIP_IDS.ADR_MARGIN]: {
    documentationPath: '/reference/adr',
  },
  [TOOLTIP_IDS.FILTER_EVENT_DATA]: {
    documentationPath: '/integrations/webhooks/creating-webhooks/',
  },
  [TOOLTIP_IDS.LORA_CLOUD_MODEM_ENCODING]: {
    externalUrl:
      'https://github.com/Lora-net/lr1110_evk_demo_app/wiki/Command-tool---Node-RED-application-server-example#building-tlv-payload',
  },
  [TOOLTIP_IDS.INPUT_METHOD]: {
    documentationPath: '/devices/adding-devices/',
  },
})

export { m, descriptions, links }
