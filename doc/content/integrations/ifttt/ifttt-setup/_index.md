---
title: "IFTTT Setup"
description: ""
weight: 1
---

Create an applet on IFTTT and prepare the setup by following the steps below.

<!--more-->

Log in to your IFTTT user account.

Select the **Create** button in the upper right. 

Click the **+ This** button.

Search for and choose **Webhooks** as the service.

{{< figure src="choosing-a-service.png" alt="Choosing Webhooks as a service" >}}

Select **Receive a web request** as a trigger.

Give a name to the trigger event and click the **Create trigger** button.

{{< figure src="naming-trigger.png" alt="Naming a trigger event" >}}

Next, click the **+ That** button.

There are many action services to choose from. In this guide, we will use the **Android SMS** action service to trigger an SMS when a LoRaWAN `Join accept` message is sent from {{% tts %}}.

{{< figure src="choosing-action-service.png" alt="Choosing Android SMS action service" >}}

You also need to specify the action within the action service. For the **Android SMS** action service, choose **Send an SMS** as an action. 

Complete the action fields by entering the phone number and SMS body, then click the **Create action** button.

>Note: you may also pass the decoded payload values from {{% tts %}} as `value1`, `value2` and `value3`. Learn how to implement this by following the [Node-RED Setup]({{< ref "/integrations/ifttt/node-red-setup" >}}) section.

{{< figure src="completing-action-fields.png" alt="Completing the action fields" >}}

Review your applet and select **Finish**.

After you have created your applet, make sure that its status is set to **Connected**.

Next, navigate to the [Webhooks service page](https://ifttt.com/maker_webhooks) and click the **Documentation** button in the upper right.

Here you will find the URL you will be sending the HTTP POST request to. Replace the **{event}** field with the name of your trigger event, then select and copy this URL.

{{< figure src="webhooks-documentation-page.png" alt="Webhooks service documentation page" >}}

>Note: you can test the action service by triggering it manually with the **Test It** button on the bottom of this page.