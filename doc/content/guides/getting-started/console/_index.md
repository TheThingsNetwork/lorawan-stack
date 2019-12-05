---
title: "Console"
description: ""
weight: 8
---

## Login

The Console needs to be logged on in order to create gateways, applications, devices and API keys. With {{% tts %}} running, open `https://thethings.example.com` (replace with the URL of your deployment) in your browser.

{{< figure src="front.png" alt="Front page" >}}

You are now on the Console landing page. Click **Login with your {{% tts %}} Account** in order to reach the login page.

{{< figure src="login.png" alt="Login" >}}

If you do not have an account yet, you can register one by clicking **Create an account**, filling your details and clicking **Register**.

{{< figure src="register.png" alt="Create an account" >}}

When you use a new account, you may not be able to continue until you have confirmed your email address or your account is approved by an admin user.

After entering your credentials and logging in, you will reach the Console overview page.

{{< figure src="overview.png" alt="Overview" >}}

## Create Gateway

Go to **Gateways** in the top menu, and click **+ Add Gateway** to reach the gateway registration page. Fill the gateway ID, gateway EUI (if your gateway has an EUI) and frequency plan. The other fields are optional. Click **Create Gateway** to create the gateway.

{{< figure src="gateway-creation.png" alt="Gateway creation" >}}

Your gateway will be created and you will be redirected to the gateway overview page of your newly created gateway. 

{{< figure src="gateway-overview.png" alt="Gateway overview" >}}

You can now connect your gateway to {{% tts %}}.

### Create Gateway API Key

Some gateways require an API Key with Link Gateway Rights to be able to connect to {{% tts %}}. 

In order to do this, navigate the **API Keys** menu of your gateway and select **Add API Key**. Enter a name for your key, select the **Link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink** right and then press **Create API Key**.

{{< figure src="gateway-api-key-creation.png" alt="Gateway API Key creation" >}}

You will see a screen that shows your newly created API Key. You now can copy it in your clipboard by pressing the clipboard button. After saving the key in a safe place, press **I have copied the key**. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.

{{< figure src="gateway-api-key-created.png" alt="Gateway API Key created" >}}

## Create Application

Go to **Applications** in the top menu, and click **+ Add Application** to reach the application registration page. Fill the application ID. The other fields are optional. Click **Create Application** to create the application.

{{< figure src="application-creation.png" alt="Application creation" >}}

Your application will be created and you will be redirected to the application overview page of your newly created application.

{{< figure src="application-overview.png" alt="Application overview" >}}

You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic. End devices are created within applications. 

## Create End Device

Go to **Devices** in the left menu and click on **+ Add Device** to reach the end device registration page. Fill the device ID, the LoRaWAN MAC and PHY versions and the frequency plan used by the device.

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

## Working With Data

With your {{% tts %}} setup, a gateway connected and a device registered on your network, it's time to start working with data.

Learn how to work with the [Application Server MQTT server]({{< ref "/reference/application-server-data/mqtt" >}}) and [HTTP webhooks]({{< ref "/reference/application-server-data/webhooks" >}}).
