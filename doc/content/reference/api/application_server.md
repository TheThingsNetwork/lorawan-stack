---
title: "Application Server APIs"
description: ""
weight: 2
---

## The `As` service

{{< proto/method service="As" method="SetLink" >}}

{{< proto/method service="As" method="GetLink" >}}

{{< proto/method service="As" method="GetLinkStats" >}}

{{< proto/method service="As" method="DeleteLink" >}}

## The `AppAs` service

{{< proto/method service="AppAs" method="DownlinkQueuePush" >}}

{{< proto/method service="AppAs" method="DownlinkQueueReplace" >}}

{{< proto/method service="AppAs" method="DownlinkQueueList" >}}

# Messages

{{< proto/message message="ApplicationDownlink" >}}

{{< proto/message message="ApplicationDownlink.ClassBC" >}}

{{< proto/message message="ApplicationDownlinks" >}}

{{< proto/message message="ApplicationIdentifiers" >}}

{{< proto/message message="ApplicationLink" >}}

{{< proto/message message="ApplicationLinkStats" >}}

{{< proto/message message="DownlinkQueueRequest" >}}

{{< proto/message message="EndDeviceIdentifiers" >}}

{{< proto/message message="GatewayAntennaIdentifiers" >}}

{{< proto/message message="GatewayIdentifiers" >}}

{{< proto/message message="GetApplicationLinkRequest" >}}

{{< proto/message message="MessagePayloadFormatters" >}}

{{< proto/message message="SetApplicationLinkRequest" >}}

# Enums

{{< proto/enum enum="PayloadFormatter" >}}

{{< proto/enum enum="TxSchedulePriority" >}}
