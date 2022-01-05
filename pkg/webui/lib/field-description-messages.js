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

const m = defineMessages({
  freqPlanDescription:
    'A Frequency Plan defines data rates and channels which comply with the LoRaWAN Regional Parameters for a band or geographical area.',
  freqPlanLocation:
    'You need to choose a Frequency Plan which adheres to the local regulations of where your end device is located. It is also important that the gateways in reach of this end device use the same Frequency Plan.',

  devEuiDescription: 'A 64 bit extended unique identifier for your end device.',
  devEuiLocation:
    'It should be provided to you by the manufacturer, or printed on the end device packaging.',
  devEuiAbsence: 'Contact the manufacturer or your reseller.',

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
  appKeyAbsence:
    'Contact the manufacturer or your reseller. If they cannot provide an AppKey, and your end device is programmable, it is okay to generate one.',

  nwkKeyDescription:
    'An end device specific encryption key used to derive the FNwkSIntKey, SNwkSIntKey, NwkSEncKey in LoRaWAN 1.1. When a LoRaWAN 1.1 capable device connects to a LoRaWAN 1.0x Network Server which does not support dual root keys (NwkKey and AppKey), the NwkKey value is used as the AppKey value.',
  nwkKeyLocation:
    'It is usually pre-provisioned by the end device manufacturer, but some end devices also allow using a user-defined value.',
  nwkKeyAbsence:
    'Contact the manufacturer or your reseller. If they cannot provide an AppKey, and your end device is programmable, it is okay to generate one.',

  devIdDescription: 'A unique, human-readable identifier for your end device.',
  devIdLocation:
    'We prefill this value using the previously entered DevEUI but you can use any other unique value you want. End device IDs can not be reused by multiple end devices within the same application.',

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
  deviceHardwareVersionAbsence:
    'Contact the manufacturer or reseller of your device. Providing an incorrect hardware version can result in unwanted device behavior.',

  deviceFirmwareVersionDescription: 'The version of firmware loaded on your device.',
  deviceFirmwareVersionLocation:
    'The firmware version should be provided by the manufacturer of your device, or printed on the device packaging. It may be possible to upgrade your device firmware to a known version.',
  deviceFirmwareVersionAbsence:
    'Contact the manufacturer or reseller of your device. Providing an incorrect hardware version can result in unwanted device behavior.',

  activationModeDescription:
    'OTAA is the preferred and most secure way to connect a device. Devices perform a join-procedure with the network. ABP requires hardcoding the device address and security keys. Multicast is a virtual group of ABP devices which allows all devices to receive the same downlinks. Multicast groups do not support uplinks.',
  activationModeLocation: 'You decide how to activate your devices. Whenever possible, use OTAA.',

  deviceNameDescription: 'An optional human readable name to help you identify your device.',
  deviceNameLocation: 'You make it up, so be creative!',

  deviceDescDescription:
    'An optional description, which can also be used to save notes about the end device.',

  frameCounterWidthDescription:
    'Most devices use a 32 bit frame counter to prevent replay attacks. Devices with extremely limited resources are permitted to use 16 bit counters.',
  frameCounterWidthLocation: 'It should be provided by the device manufacturer.',
  frameCounterWidthAbsence:
    'Contact your manufacturer or reseller. Most devices use 32 bit counters. Selecting the wrong value will produce errors once the Up or Down frame counter exceeds 16 bits and rolls over.',

  rx2DataRateIndexDescription:
    'The data rate used for the second reception window used by this end device to receive downlinks.',
  rx2DataRateIndexLocation: 'This should be provided by the device manufacturer.',
  rx2DataRateIndexAbsence: 'Contact your device manufacturer or reseller.',

  rx2FrequencyDescription:
    'The frequency used for the second reception window used by this end device to receive downlinks.',
  rx2FrequencyLocation: 'This should be provided by the device manufacturer.',
  rx2FrequencyAbsence: 'Contact your device manufacturer or reseller.',

  gatewayIdDescription: 'A unique identifier for your gateway.',
  gatewayIdLocation: 'You make it up, so be creative!',

  gatewayEuiDescription: 'A 64 bit extended unique identifier for your end device.',
  gatewayEuiLocation:
    'It should be provided to you by the manufacturer, or printed on the gateway packaging.',
  gatewayEuiAbsence: 'Contact the manufacturer or reseller.',

  gatewayNameDescription: 'An optional human readable name to help you identify your gateway.',
  gatewayNameLocation: 'You make it up, so be creative!',

  gatewayDescDescription:
    'An optional description, which can also be used to save notes about the gateway.',

  requireAuthenticatedConnectionDescription:
    'This will only allow a gateway to connect if it uses a TLS enabled Basic Station or MQTT connection. It will not allow connections from UDP packet forwarders.',

  gatewayStatusDescription:
    'Setting your gateway status to public allows status information about the gateway to be shared with other users in the network, and with Packet Broker if enabled by the network operator.',
  gatewayLocationDescription:
    'Setting your gateway location to public allows location information about the gateway to be shared with other users in the network, and with Packet Broker if enabled by the network operator.',

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

  clusterSettingsDescription:
    'By default, the server components of the current cluster are used. However, for advanced use cases, it is possible to register this end device to different Network Server and/or Join Server.',

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
    'Skip payload crypto disables the application layer encryption of LoRaWAN frames. This causes the Application Server to forward the messages without any processing, such as payload formatters, to the integrations.  When doing so, the integrations are responsible for decryption and processing of the binary format in order to understand the message. This application-wide setting can be overwritten per end device using this overwrite setting.',
  basicAuthDescription:
    'To increase access security, you can choose to generate a "basic auth" authorization header to be attached to the webhook requests, if the target server requires doing so. This will authenticate the webhook requests with the defined credentials.',
})

const descriptions = Object.freeze({
  [TOOLTIP_IDS.FREQUENCY_PLAN]: {
    description: m.freqPlanDescription,
    location: m.freqPlanLocation,
    glossaryId: GLOSSARY_IDS.FREQUENCY_PLAN,
  },
  [TOOLTIP_IDS.DEV_EUI]: {
    description: m.devEuiDescription,
    location: m.devEuiLocation,
    absence: m.devEuiAbsence,
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
    absence: m.appKeyAbsence,
    glossaryId: GLOSSARY_IDS.APP_KEY,
  },
  [TOOLTIP_IDS.NETWORK_KEY]: {
    description: m.nwkKeyDescription,
    location: m.nwkKeyLocation,
    absence: m.nwkKeyAbsence,
    glossaryId: GLOSSARY_IDS.NETWORK_KEY,
  },
  [TOOLTIP_IDS.DEVICE_ID]: {
    description: m.devIdDescription,
    location: m.devIdLocation,
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
    absence: m.deviceHardwareVersionAbsence,
  },
  [TOOLTIP_IDS.DEVICE_FIRMWARE_VERSION]: {
    description: m.deviceFirmwareVersionDescription,
    location: m.deviceFirmwareVersionLocation,
    absence: m.deviceFirmwareVersionAbsence,
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
    absence: m.rx2DataRateIndexAbsence,
  },
  [TOOLTIP_IDS.RX2_FREQUENCY]: {
    description: m.rx2FrequencyDescription,
    location: m.rx2FrequencyLocation,
    absence: m.rx2FrequencyAbsence,
  },
  [TOOLTIP_IDS.GATEWAY_ID]: {
    description: m.gatewayIdDescription,
    location: m.gatewayIdLocation,
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
  [TOOLTIP_IDS.REQUIRE_AUTHENTICATED_CONNECTION]: {
    description: m.requireAuthenticatedConnectionDescription,
  },
  [TOOLTIP_IDS.GATEWAY_STATUS]: {
    description: m.gatewayStatusDescription,
  },
  [TOOLTIP_IDS.GATEWAY_LOCATION]: {
    description: m.gatewayLocationDescription,
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
  [TOOLTIP_IDS.CLUSTER_SETTINGS]: {
    description: m.clusterSettingsDescription,
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
})

export { descriptions, links }
