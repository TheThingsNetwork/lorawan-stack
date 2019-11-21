---
title: 'Console'
description: ''
weight: 8
---

## Login

The Console needs to be logged on in order to create gateways, applications, devices and API keys. With The Things Stack running, open `https://thethings.example.com` (replace with the URL of your deployment) in your browser.

{{< figure src="front.png" alt="Front page" >}}

You are now on the Console landing page. Click **Login with your The Things Stack Account** in order to reach the login page.

{{< figure src="login.png" alt="Login" >}}

After entering your credentials and logging in, you will reach the Console overview page. You can now fully use the capabilities of the Console.

{{< figure src="overview.png" alt="Overview" >}}

## Create gateway

Go to **Gateways** in the top menu, and click **+ Add Gateway** to reach the gateway registration page. Fill the gateway ID, gateway EUI (if your gateway has an EUI) and frequency plan. The other fields are optional. Click **Create Gateway** to create the gateway.

{{< figure src="gateway-creation.png" alt="Gateway creation" >}}

Your gateway will be created and you will be redirected to the gateway overview page of your newly created gateway.

{{< figure src="gateway-overview.png" alt="Gateway overview" >}}

You can now connect your gateway to The Things Stack.

## Create application

Go to **Applications** in the top menu, and click **+ Add Application** to reach the application registration page. Fill the application ID. The other fields are optional. Click **Create Application** to create the application.

{{< figure src="application-creation.png" alt="Application creation" >}}

Your application will be created and you will be redirected to the application overview page of your newly created application.

{{< figure src="application-overview.png" alt="Application overview" >}}

Devices are created within applications.

### Link application

If you haven't unchecked the "Link automatically" checkbox during creation, your device will be automatically linked to the Application Server. You can skip this section in this case.

In order to send uplinks and receive downlinks from your device, you must link the Application Server to the Network Server. To do this, create an API key for the Application Server by going to **API keys** in the left menu of your application, and then clicking **+ Add API Key**.

In the API Key creation screen, enter a name for your linking API key and select the **Link as Application to a Network Server** right, then press **Create API Key**.

{{< figure src="api-key-creation.png" alt="Application API Key creation" >}}

You will see a screen that shows your newly created API Key. You now can copy it in your clipboard by pressing the clipboard button. After saving the key in a safe place, press **I have copied the key**. You will not be able to see this key again in the future, but if you lose it, you can create a new one to replace it.

{{< figure src="api-key-created.png" alt="Application API Key created" >}}

Now go to **Link** in the left menu of the application and enter the API key you've just created. You can leave the Network Server address empty. Press **Save Changes** to save the link settings.

{{< figure src="application-link-creation.png" alt="Application link creation" >}}

You can now see the status of the linking process appear in the right part of your screen. This also shows the statistics of the link between the Application Server and the Network Server.

Your application is now linked. You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic.

## Create end device

Go to **Devices** in the left menu and click on **+ Add Device** to reach the end device registration page. Fill the device ID, the LoRaWAN MAC and PHY versions and the frequency plan used by the device.

{{< figure src="device-creation-1.png" alt="Creating a new device" >}}

### Over-the-air-activation (OTAA) device

After filling the fields in the "General Settings" section, scroll to the lower part of the device registration page and make sure that "Over The Air Activation (OTAA)" is selected. Fill the Join EUI (App EUI in LoRaWAN versions before 1.1), and the Device EUI. Based on whether or not you're using an external Join Server, you can also set the AppKey and NwkKey, which will be generated automatically if you leave the fields blank. Press **Create Device** to create the device.

{{< figure src="device-creation-otaa.png" alt="Creating an OTAA device" >}}

You'll now reach the device overview page for your device. The end device should now be able to join the private network.

{{< figure src="device-otaa-created.png" alt="OTAA device overview" >}}

### Activation by personalization (ABP device)

After filling the fields in the "General Settings" section, scroll to the lower part of the device registration page and make sure that "Activation By Personalization (ABP)" is selected. Fill the Device Address, the FNwkSIntKey (NwkSKey in LoRaWAN versions before 1.1) and the AppSKey. The other key fields are only needed for LoRaWAN version 1.1 or later. All other fields on the page are either optional or generated automatically for you when left blank. Press **Create Device** to create the device.

{{< figure src="device-creation-abp.png" alt="Creating an ABP device" >}}

You'll now reach the device overview page for your device. The end device should now be able to communicate with the private network.

{{< figure src="device-abp-created.png" alt="ABP device overview" >}}

## Working with data

With your The Things Stack setup, a gateway connected and a device registered on your network, it's time to start working with data.

Learn how to work with the [builtin MQTT server]({{< relref "../mqtt" >}}) and [HTTP webhooks]({{< relref "../webhooks" >}}).
