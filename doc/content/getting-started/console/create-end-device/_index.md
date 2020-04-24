---
title: "Create End Device"
description: ""
weight: 4
---

Learn how to register an end device using the Console.

<!--more-->

## Step by step

Go to **Devices** in the left menu and click on **+ Add Device** to reach the end device registration page. Fill the device ID, the LoRaWAN MAC and PHY versions and the frequency plan used by the device.

> Note: The PHY version represent the revision of the LoRAWAN Regional Parameters that the device expects, and must be correlated with the MAC version.

{{< figure src="device-creation-1.png" alt="Creating a new device" >}}

### Over-The-Air-Activation (OTAA) Device

After filling the fields in the **General Settings** section, scroll to the lower part of the device registration page and make sure that "Over The Air Activation (OTAA)" is selected. Fill the JoinEUI (AppEUI in LoRaWAN versions before 1.1), the DevEUI and AppKey. The NwkKey is only needed for LoRaWAN version 1.1 or later. All other fields on the page are optional. Press **Create Device** to create the device.

{{< figure src="device-creation-otaa.png" alt="Creating an OTAA device" >}}

You will now reach the device overview page for your device. The end device should now be able to join the private network.

{{< figure src="device-otaa-created.png" alt="OTAA device overview" >}}

### Activation By Personalization (ABP) Device

After filling the fields in the "General Settings" section, scroll to the lower part of the device registration page and make sure that "Activation By Personalization (ABP)" is selected. Fill the Device Address, the FNwkSIntKey (NwkSKey in LoRaWAN versions before 1.1) and the AppSKey. The other key fields are only needed for LoRaWAN version 1.1 or later. All other fields on the page are optional. Press **Create Device** to create the device.

{{< figure src="device-creation-abp.png" alt="Creating an ABP device" >}}

You will now reach the device overview page for your device. The end device should now be able to communicate with the private network.

{{< figure src="device-abp-created.png" alt="ABP device overview" >}}
