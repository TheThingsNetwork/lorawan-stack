---
title: "UDP Packet Forwarder"
description: ""
---

This section contains instructions for connecting to {{% tts %}} using the UDP Packet Forwarder.

<!--more-->

In the **Packet Forwarder** menu, click the **Plus** button to create a new server.

{{< figure src="../plus.png" alt="Create new server" >}}

In the server configuration options, check the **Enabled** box.

Choose **Semtech** as the **Type**.

For the **Server Address** choose **custom**, and enter the same as what you use instead of `thethings.example.com` in the [Getting Started guide]({{< ref "/getting-started" >}}).

Choose the appropriate **Port Up** and **Port Down** values. These are both **1700** by default in {{% tts %}}.

Click **Save** to continue.

{{< figure src="../semtech.png" alt="Semtech Configuration" >}}

If your configuration was successful, your gateway will connect to {{% tts %}} after a couple of seconds.
