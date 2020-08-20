---
title: "Adding Applications"
description: ""
weight: -1
aliases: [/getting-started/cli/create-application, /getting-started/console/create-application, /guides/getting-started/console/create-application]
---

This section contains instructions for creating an Application.

<!--more-->

{{< tabs/container "Console" "CLI" >}}

{{< tabs/tab "Console" >}}

## Adding Applications using the Console

Go to **Applications** in the top menu, and click **+ Add Application** to reach the application registration page. Fill the application ID. The other fields are optional. Click **Create Application** to create the application.

{{< figure src="application-creation.png" alt="Application creation" >}}

Your application will be created and you will be redirected to the application overview page of your newly created application.

{{< figure src="application-overview.png" alt="Application overview" >}}

You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic. End devices are created within applications.

### Link application

If you did not uncheck the **Link automatically** checkbox during creation, your application will be automatically linked to the Application Server. You can skip this section in this case.

In order to send uplinks and receive downlinks from your device, you must link the Application Server to the Network Server. To do this, create an API key for the Application Server by going to **API keys** in the left menu of your application, and then clicking **+ Add API Key**.

In the API Key creation screen, enter a name for your linking API key and select the **Link as Application to a Network Server** right, then press **Create API Key**.  In the API Key creation screen, enter a name for your linking API key and select the **Link as Application to a Network Server** right, then press **Create API Key**.

{{< figure src="api-key-creation.png" alt="Application API Key creation" >}}

You will see a screen that shows your newly created API Key. Copy it in your clipboard by pressing the clipboard button. After saving the key in a safe place, press **I have copied the key**. You will not be able to see this key again in the future, but if you lose it, you can create a new one to replace it.  You will see a screen that shows your newly created API Key. You now can copy it in your clipboard by pressing the clipboard button. After saving the key in a safe place, press **I have copied the key**. You will not be able to see this key again in the future, but if you lose it, you can create a new one to replace it.

{{< figure src="api-key-created.png" alt="Application API Key created" >}}

Now go to **Link** in the left menu of the application and enter the API key you've just created. You can leave the Network Server address empty. Press **Save Changes** to save the link settings.  Now go to **Link** in the left menu of the application and enter the API key you've just created. You can leave the Network Server address empty. Press **Save Changes** to save the link settings.

{{< figure src="application-link-creation.png" alt="Application link creation" >}}

You can now see the status of the linking process appear in the right part of your screen. This also shows the statistics of the link between the Application Server and the Network Server.  You can now see the status of the linking process appear in the right part of your screen. This also shows the statistics of the link between the Application Server and the Network Server.

Your application is now linked. You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic.

{{< /tabs/tab >}}

{{< tabs/tab "CLI" >}}

## Adding Applications using the CLI

Create the first application:

```bash
$ ttn-lw-cli applications create app1 --user-id admin
```

This creates an application `app1` with the `admin` user as collaborator.

Devices are created within applications.

### Link Application

In order to send uplinks and receive downlinks from your device, you must link the Application Server to the Network Server. In order to do this, create an API key for the Application Server:

```bash
$ ttn-lw-cli applications api-keys create \
  --name link \
  --application-id app1 \
  --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used for linking this application. Make sure to copy the key and save it in a safe place. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.

You can now link the Application Server to the Network Server:

```bash
$ ttn-lw-cli applications link set app1 --api-key NNSXS.VEEBURF3KR77ZR..
```

Your application is now linked. You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic.

{{< /tabs/tab >}}

{{< /tabs/container >}}
