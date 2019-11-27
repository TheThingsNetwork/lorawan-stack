---
title: "End Device Claiming APIs"
description: ""
weight: 8
---

## The `EndDeviceClaimingServer` service

{{< proto/method service="EndDeviceClaimingServer" method="AuthorizeApplication" >}}

{{< proto/method service="EndDeviceClaimingServer" method="UnauthorizeApplication" >}}

{{< proto/method service="EndDeviceClaimingServer" method="Claim" >}}

## Messages

{{< proto/message message="ApplicationIdentifiers" >}}

{{< proto/message message="AuthorizeApplicationRequest" >}}

{{< proto/message message="ClaimEndDeviceRequest" >}}

{{< proto/message message="ClaimEndDeviceRequest.AuthenticatedIdentifiers" >}}

{{< proto/message message="EndDeviceIdentifiers" >}}
