---
title: "End Device APIs"
description: ""
weight: 7
---

End devices are registered in multiple registries. The Identity Server has a registry with end device metadata, the Network Server's registry contains the MAC configuration and MAC state (including network session keys), the Application Server keeps payload formatters and application session keys, the Join Server keeps the root keys.

When registering end devices, we recommend registering them in the following order:

- `EndDeviceRegistry.Create` (Identity Server)
- `JsEndDeviceRegistry.Set` (Join Server, only for OTAA devices)
- `NsEndDeviceRegistry.Set` (Network Server)
- `AsEndDeviceRegistry.Set` (Application Server)

When deleting end devices, we recommend deleting them in the reverse order

## The `EndDeviceRegistry` service

The Identity Server's `EndDeviceRegistry` is the first place to register an end device. This registry stores the following [EndDevice fields](#message:EndDevice):

- `ids` (with subfields)
- `name`
- `description`
- `attributes`
- `version_ids` (with subfields)
- `network_server_address`
- `application_server_address`
- `join_server_address` (only for OTAA devices)
- `service_profile_id`
- `locations`
- `picture`

{{< proto/method service="EndDeviceRegistry" method="Create" >}}

{{< proto/method service="EndDeviceRegistry" method="Get" >}}

{{< proto/method service="EndDeviceRegistry" method="List" >}}

{{< proto/method service="EndDeviceRegistry" method="Update" >}}

{{< proto/method service="EndDeviceRegistry" method="Delete" >}}

## The `JsEndDeviceRegistry` service

OTAA devices are registered in the Join Server's `JsEndDeviceRegistry`. This registry stores the following [EndDevice fields](#message:EndDevice):

- `ids` (with subfields)
- `provisioner_id` (when provisioning with secure elements)
- `provisioning_data` (when provisioning with secure elements)
- `resets_join_nonces`
- `root_keys`:
  - `root_key_id`
  - `app_key`
  - `nwk_key`
- `net_id`
- `network_server_address`
- `network_server_kek_label`
- `application_server_address`
- `application_server_id`
- `application_server_kek_label`
- `claim_authentication_code` (when using [end device claiming]({{< relref "end_device_claiming.md" >}}))

{{< proto/method service="JsEndDeviceRegistry" method="Set" >}}

{{< proto/method service="JsEndDeviceRegistry" method="Get" >}}

{{< proto/method service="JsEndDeviceRegistry" method="Delete" >}}

## The `NsEndDeviceRegistry` service

The Network Server's `NsEndDeviceRegistry` stores the following [EndDevice fields](#message:EndDevice):

- `ids` (with subfields)
- `frequency_plan_id`
- `lorawan_phy_version`
- `lorawan_version`
- `mac_settings` (with subfields)
- `mac_state` (with subfields)
- `supports_join`
- `multicast`
- `supports_class_b`
- `supports_class_c`
- `session.dev_addr`
- `session.keys`:
  - `session_key_id`
  - `f_nwk_s_int_key`
  - `s_nwk_s_int_key`
  - `nwk_s_enc_key`

{{< proto/method service="NsEndDeviceRegistry" method="Set" >}}

{{< proto/method service="NsEndDeviceRegistry" method="Get" >}}

{{< proto/method service="NsEndDeviceRegistry" method="Delete" >}}

## The `AsEndDeviceRegistry` service

The Network Server's `NsEndDeviceRegistry` stores the following [EndDevice fields](#message:EndDevice):

- `ids` (with subfields)
- `formatters`:
  - `up_formatter`
  - `up_formatter_parameter`
  - `down_formatter`
  - `down_formatter_parameter`
- `session.dev_addr`
- `session.keys`:
  - `session_key_id`
  - `app_s_key`

{{< proto/method service="AsEndDeviceRegistry" method="Set" >}}

{{< proto/method service="AsEndDeviceRegistry" method="Get" >}}

{{< proto/method service="AsEndDeviceRegistry" method="Delete" >}}

## Messages

{{< proto/message message="CreateEndDeviceRequest" >}}

{{< proto/message message="EndDevice" >}}

{{< proto/message message="EndDeviceAuthenticationCode" >}}

{{< proto/message message="EndDeviceIdentifiers" >}}

{{< proto/message message="EndDevices" >}}

{{< proto/message message="EndDeviceVersionIdentifiers" >}}

{{< proto/message message="GetEndDeviceRequest" >}}

{{< proto/message message="KeyEnvelope" >}}

{{< proto/message message="ListEndDevicesRequest" >}}

{{< proto/message message="MACParameters" >}}

{{< proto/message message="MACSettings" >}}

{{< proto/message message="MACState" >}}

{{< proto/message message="MessagePayloadFormatters" >}}

{{< proto/message message="RootKeys" >}}

{{< proto/message message="Session" >}}

{{< proto/message message="SessionKeys" >}}

{{< proto/message message="SetEndDeviceRequest" >}}

{{< proto/message message="UpdateEndDeviceRequest" >}}

## Enums

{{< proto/enum enum="MACVersion" >}}

{{< proto/enum enum="PHYVersion" >}}

{{< proto/enum enum="PowerState" >}}
