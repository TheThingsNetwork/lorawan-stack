---
title: "Wifx LORIX One"
description: ""
---

The LORIX One is a robust and professional grade outdoor LoRaWAN® gateway in an ultra compact form factor, designed and assembled in Switzerland. The LORIX One supports ethernet, wireless, and cellular backhauls.

This page will guide you through the steps required to connect the gateway to {{% tts %}}.

<!--more-->

![LORIX One](lorix-one.png)

For additional help and technical specifications, please refer to [Wifx's official documentation](https://iot.wifx.net/docs).

## Requirements

  1. User account on {{% tts %}} with rights to create Gateways
  2. A Wifx LORIX One running LORIX OS connected to the network
  3. A computer, tablet or mobile phone connected to the network (to configure the gateway)

If your gateway is running legacy software, please refer to [the official documentation](https://iot.wifx.net/docs) for upgrade instructions.

## Get the gateway EUI

To register the gateway, you will need its Extended Unique Identifier (Gateway EUI). This can be found either on the gateway's sticker or by software in the Manager UI.

### From the sticker

To get the Gateway EUI from the sticker, find the MAC address printed on the sticker under the gateway. The gateway EUI corresponds to the MAC address, removing the `;` and adding `FFFE` in the middle.

For example the MAC address

```
FC:C2:3D:AB:CD:EF
```

corresponds to a Gateway EUI of

```
FCC23DFFFEABCDEF
```

The full process for conversion is:
```
FC:C2:3D:AB:CD:EF => FCC23DABCDEF => FCC23D FFFE ABCDEF => FCC23DFFFEABCDEF
```

### From the Manager UI

To get the Gateway EUI from the Manager UI, connect to your gateway and check the **System > Information** page. Under the System section you will see the 'Serial number'. This serial number is the EUI.

## Registration

Create a gateway by following the instructions for [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}).

## Configuration

To connect to the LORIX One, open a web browser on your computer or device and enter the either the gateway hostname or the gateway IP address.

The hostname is `lorix-one-abcdef.local` where `abcdef` are the 6 last digits of the Gateway EUI.

> Note: hostname access is only available on networks that have mDNS enabled. On networks without mDNS, enter the IP address of the gateway in the web browser.

You will land on the login page. Log on using the following the default username **admin** and default password **lorix4u**.

{{< figure src="lorix-one-login.png" alt="LORIX One login page" >}}

### Configure the antenna type

Go to the **LoRa > Settings page > Hardware tab**.

![LORIX One LoRa hardware page](lorix-one-lora-settings-antenna.png "LORIX One LoRa hardware page")

In the **Antenna** field, select the antenna you have connected.

- 2dBi is the small antenna (~20cm)
- 4dBi is the big antenna (~40cm)

> Note: if the antenna type is not configured, the packet forwarder will fail to start.

## Configure the Packet Forwarder

After completing basic configuration, follow the instructions for connecting using [{{% lbs %}}]({{< relref "lbs" >}}) or the [UDP Packet Forwarder]({{< relref "udp" >}}).
