---
title: "Gateway APIs"
description: ""
weight: 5
---

# The `GatewayRegistry` service

{{< proto/method service="GatewayRegistry" method="Create" >}}

{{< proto/method service="GatewayRegistry" method="Get" >}}

{{< proto/method service="GatewayRegistry" method="List" >}}

{{< proto/method service="GatewayRegistry" method="Update" >}}

{{< proto/method service="GatewayRegistry" method="Delete" >}}

# The `GatewayAccess` service

{{< proto/method service="GatewayAccess" method="ListRights" >}}

{{< proto/method service="GatewayAccess" method="CreateAPIKey" >}}

{{< proto/method service="GatewayAccess" method="ListAPIKeys" >}}

{{< proto/method service="GatewayAccess" method="GetAPIKey" >}}

{{< proto/method service="GatewayAccess" method="UpdateAPIKey" >}}

{{< proto/method service="GatewayAccess" method="GetCollaborator" >}}

{{< proto/method service="GatewayAccess" method="SetCollaborator" >}}

{{< proto/method service="GatewayAccess" method="ListCollaborators" >}}

# The `Configuration` service

The Gateway Server exposes the list of available frequency plans with the `Configuration` service.

{{< proto/method service="Configuration" method="ListFrequencyPlans" >}}

# Messages

{{< proto/message message="APIKey" >}}

{{< proto/message message="APIKeys" >}}

{{< proto/message message="Collaborator" >}}

{{< proto/message message="Collaborators" >}}

{{< proto/message message="CreateGatewayAPIKeyRequest" >}}

{{< proto/message message="CreateGatewayRequest" >}}

{{< proto/message message="FrequencyPlanDescription" >}}

{{< proto/message message="Gateway" >}}

{{< proto/message message="GatewayAntenna" >}}

{{< proto/message message="GatewayIdentifiers" >}}

{{< proto/message message="Gateways" >}}

{{< proto/message message="GatewayVersionIdentifiers" >}}

{{< proto/message message="GetCollaboratorResponse" >}}

{{< proto/message message="GetGatewayAPIKeyRequest" >}}

{{< proto/message message="GetGatewayCollaboratorRequest" >}}

{{< proto/message message="GetGatewayRequest" >}}

{{< proto/message message="ListFrequencyPlansRequest" >}}

{{< proto/message message="ListFrequencyPlansResponse" >}}

{{< proto/message message="ListGatewayAPIKeysRequest" >}}

{{< proto/message message="ListGatewayCollaboratorsRequest" >}}

{{< proto/message message="ListGatewaysRequest" >}}

{{< proto/message message="OrganizationOrUserIdentifiers" >}}

{{< proto/message message="Rights" >}}

{{< proto/message message="SetGatewayCollaboratorRequest" >}}

{{< proto/message message="UpdateGatewayAPIKeyRequest" >}}

{{< proto/message message="UpdateGatewayRequest" >}}

{{< proto/message message="UserIdentifiers" >}}

# Enums

{{< proto/enum enum="DownlinkPathConstraint" >}}
