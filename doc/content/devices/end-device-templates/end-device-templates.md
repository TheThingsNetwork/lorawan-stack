---
title: "End Device Templates"
description: ""
summary: End device templates define the blueprint of a device. It can be used to create one device or many devices at once. Templates allow for using common values for many devices, such as a device profile.
---

## What is it?

End device templates define the blueprint of a device. It can be used to create one device or many devices at once. Templates allow for using common values for many devices, such as a device profile.

## Who is it for?

End device templates are primarily targeted at device makers and service providers who are managing and onboarding large amounts of devices.

### Typical use cases

1. Create a batch of end devices with the same profile and incrementing `DevEUI` from a range
2. Convert vendor-specific end device data, such as serial numbers and provisioning data, to a device template
3. Migrate end devices from a different LoRaWAN server stack

## How does it work?

End device templates can be used to quickly create large amounts of end devices with common settings. Templates can be created from existing devices or converted from an input file. By executing templates, you can create devices directly. See [Creating Templates]({{< relref "creating.md" >}}), [Converting Templates]({{< relref "converting.md" >}}) and [Executing Templates]({{< relref "executing.md" >}}).

Templates can also be be mapped with other templates to combine fields. For example, you can convert root key provisioning data to device templates and map that with a template containing the device profile from a device repository. See [Mapping Templates]({{< relref "mapping.md" >}}).

Tooling supports assigning the LoRaWAN `JoinEUI` and `DevEUI` from a range automatically to create a template file that can be used in mapping or executing templates to bulk create devices. See [Assigning EUIs]({{< relref "assigning-euis.md" >}}).
