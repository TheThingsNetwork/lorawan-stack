---
title: "MultiTech Conduit AEP"
description: ""
---

The [MultiTech MultiConnect® Conduit™](http://www.multitech.net/developer/products/multiconnect-Conduit-platform/) is a configurable, scalable cellular communications gateway for industrial IoT applications. The technical specifications of this Conduit can be found in the software guide in the [official documentation](http://www.multitech.net/developer/products/multiconnect-conduit-platform/conduit/) page.

This guide will help you set up the Multitech Conduit gateway to communicate over {{%tts%}}.
<!--more-->
{{< figure src="001_Multitech_Conduit_AEP.png" alt="MultiTech MultiConnect Conduit AEP Gateway" >}}

## Prerequisites

1. User account on {{% tts %}} with rights to create Gateways.
2. Multitech Conduit AEP Gateway connected to the internet and running the [latest firmware](http://www.multitech.net/developer/downloads/).

> This article addresses the minimum configuration required to connect your Multitech Conduit AEP model to {{% tts %}}.
> For the first-time hardware and software setup, you can refer to the appropriate configuration guide on the [Multitech documentation](http://www.multitech.net/developer/products/multiconnect-conduit-platform/conduit/) page.

## Registration

Create a gateway by following the instructions for the [Console]({{< ref "/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/getting-started/cli#create-gateway" >}}).

> Note: The gateway EUI can be found at the bottom of the gateway under the field &quot;NODE LORA&quot;.

## Configuration using a Terminal

Firstly, you will need to generate a `global_conf.json` file required to add a custom channel plan to your gateway. Follow the steps below to generate the required json file.

> **Warning:** The manual configuration of the frequency plan - required to make it work on the The Things (Enterprise) Stack - may result in a mismatch for some of the frequency bands used for the Multitech Gateway.

The Gateway Configuration Server can be used to generate a proper `global_conf.json` configuration file for your gateway. You will need a Gateway API key with the `View gateway information` right enabled. The instructions for the same can be found in the relevant sections of the [Console]({{< ref "/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/getting-started/cli#create-gateway" >}}) getting started guides.

Open the command prompt in Windows or any Linux terminal to run a curl command (as shown below) to generate the required `global_conf.json` file in your current working directory.

Make sure you replace `thethings.example.com` with your server address:

```bash
$ curl -XGET \
    "https://thethings.example.com/api/v3/gcs/gateways/{GATEWAY_ID}/semtechudp/global_conf.json" \
    -H "Authorization: Bearer {GTW_API_KEY}" > ~/global_conf.json
```

> Note: Replace the required fields in the above command and run it.

Once the `global_conf.json` file is generated, you will need to add this to your gateway. In a web browser, open the gateway’s configuration page by navigating to its IP Address obtained from the network it is connected to. Once logged in, you can configure the gateway to connect to {{%tts%}} by following the steps below:

- Click on **LoRaWAN<sup>®</sup>** in the menu on the left. It opens the Gateway&apos;s configuration page.

{{< figure src="005_Gateway_Menu_LoRaWAN.png" alt="MultiTech Conduit gateway home page" >}}

- Under **Network Settings**, select the mode as &quot;Packet Forwarder&quot;.

{{< figure src="006_Packet_Forwarder_Gateway.png" alt="MultiTech Conduit packet forwarder settings" >}}

- On the right side of the &quot;LoRa Packet Forwarder Configuration&quot; section, you can find &quot;Manual Configuration &quot;. Click on it to setup the channel plan manually.

{{< figure src="001_Gateway_Frequency_plan_manual_config.png" alt="Switch to manual configuration mode" >}}

- The above step will lead you to the gateway configuration editor tagged as &quot;Config&quot;.

{{< figure src="002_Gateway_Frequency_plan_manual_config.png" alt="edit the configuration" >}}

- Copy the contents of the `global_conf.json` file downloaded earlier and paste them in the gateway console configuration editor.

- Once pasted, modify the value of `clksrc` to set it to `0`.

- Click on **Submit** to save the configuration.

- Now, click on **Save and Restart** from the menu.

{{< figure src="011_Gateway_Menu_Save-And-Restart.png" alt="Saving the Network Interfaces Configuration" >}}

- You will be prompted to confirm the restart. Choose **OK** to proceed.

{{< figure src="012_System_Confirmation_Prompt.png" alt="Configuration restart prompt" >}}

This will apply the custom settings and reboot the gateway. If all the steps have been followed correctly, your gateway will now connect to {{%tts%}}.

> To know more about other features of the MultiTech Conduit gateway, you can refer to the **mPower Edge AEP software guide** on the [Multitech Website](http://www.multitech.net/developer/products/multiconnect-conduit-platform/conduit/).

## Troubleshooting

### Packet Forwarder Logs

Packet forwarder logs contain the messages received from the devices and the network server. You can use the information to debug the issues.

You can find the packet forwarder logs using the gateway&apos;s web console. Do the following to download the logs onto your device:

- Login to your gateway&apos;s web console and navigate to **Administration -> Debug Options**.
- Click on the **Download Logs** button in the **Logging** section to download the logs.

{{< figure src="016_Packet_Forwarder.png" alt="Packet forwarder" >}}

Now, in the downloaded logs, open the *lora-pkt-fwd-1.log* file to see the packet forwarder logs.

### Upgrading the Firmware

If you have an issue with the current firmware version, you can use following steps to upgrade the firmware.

Refer to the top of your configuration software window to check your firmware version. You have to upgrade the device&apos;s firmware to the latest version.

You can download the firmware upgrades from the downloads section of the [MultiTech website](http://www.multitech.net/developer/downloads/).

Do the following to upgrade the firmware on your device:

- Before you upgrade your firmware, save your present configuration as a backup.
- Go to the MultiTech website, locate the firmware upgrade file you want for your device(MTCDT AEP x.x.x in the current case, where x.x.x is the latest version available on the website), and download this file to a known location.
- Navigate to **Administration -> Firmware Upgrade**.
- Click on **Choose Firmware Upgrade File** button, and
  - Click on **Browse** to find where the firmware file resides that you want to apply.
  - Select the file and click on **Open**. The file name appears next to the **Choose Firmware Upgrade File** button. Make sure you select the correct BIN file; otherwise, your device can become inoperable.
- Click on **Start Upgrade**.

{{< figure src="017_Firmware_Upgrade.png" alt="Gateway firmware upgrade window" >}}

- A message about the time needed to upgrade appears. Click on **OK**.
- A progress bar appears indicating the status of the upgrade. When the upgrade is completed, your device reboots.
- After the firmware upgrade is completed, verify your configuration to ensure that it is same as what you expected.

> Note: The new firmware is written into flash memory. It may take up to 10 minutes to upgrade the firmware. Do not interrupt the devices&apos; power or press the reset button during this time.
