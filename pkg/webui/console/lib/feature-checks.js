// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { selectStackConfig } from '@ttn-lw/lib/selectors/env'

import { selectApplicationRights } from '@console/store/selectors/applications'
import { selectGatewayRights } from '@console/store/selectors/gateways'
import { selectOrganizationRights } from '@console/store/selectors/organizations'
import { selectUserRights, selectUserIsAdmin } from '@console/store/selectors/user'

const stackConfig = selectStackConfig()
const asEnabled = stackConfig.as.enabled
const gsEnabled = stackConfig.gs.enabled
const gcsEnabled = stackConfig.gcs.enabled

export const checkFromState = (featureCheck, state) =>
  featureCheck.check(featureCheck.rightsSelector(state))

// User related feature checks.
export const mayViewApplicationsOfUser = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_USER_APPLICATIONS_LIST'),
}
export const mayCreateApplications = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_USER_APPLICATIONS_CREATE'),
}
export const mayViewGatewaysOfUser = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_USER_GATEWAYS_LIST'),
}
export const mayCreateGateways = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_USER_GATEWAYS_CREATE'),
}
export const mayViewOrganizationsOfUser = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_USER_ORGANIZATIONS_LIST'),
}
export const mayCreateOrganizations = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_USER_ORGANIZATIONS_CREATE'),
}
export const mayViewOrEditApiKeys = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_USER_SETTINGS_API_KEYS'),
}
export const maySendInvites = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_SEND_INVITES'),
}

// Application related feature checks.
export const mayViewApplicationInfo = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_INFO'),
}
export const mayEditBasicApplicationInfo = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_SETTINGS_BASIC'),
}
export const mayViewApplicationLink = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_LINK') && asEnabled,
}
export const maySetApplicationPayloadFormatters = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_SETTINGS_BASIC') && asEnabled,
}
export const mayViewApplicationEvents = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_TRAFFIC_READ'),
}
export const mayViewOrEditApplicationApiKeys = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_SETTINGS_API_KEYS'),
}
export const mayViewApplicationDevices = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_DEVICES_READ'),
}
export const mayCreateOrEditApplicationDevices = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_DEVICES_WRITE'),
}
export const mayCreateOrEditApplicationIntegrations = {
  rightsSelector: selectApplicationRights,
  check: rights => mayEditBasicApplicationInfo.check(rights) && asEnabled,
}
export const mayViewMqttConnectionInfo = {
  rightsSelector: selectApplicationRights,
  check: rights => mayViewApplicationInfo.check(rights) && asEnabled,
}
export const mayViewOrEditApplicationCollaborators = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_SETTINGS_COLLABORATORS'),
}
export const mayDeleteApplication = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_DELETE'),
}
export const mayReadApplicationDeviceKeys = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_DEVICES_READ_KEYS'),
}
export const mayEditApplicationDeviceKeys = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_DEVICES_WRITE_KEYS'),
}
export const maySendUplink = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_TRAFFIC_UP_WRITE') && asEnabled,
}
export const mayScheduleDownlinks = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE') && asEnabled,
}
export const mayViewOrEditApplicationPackages = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_SETTINGS_PACKAGES') && asEnabled,
}

// Gateway related feature checks.
export const mayViewGatewayInfo = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_INFO'),
}
export const mayEditBasicGatewayInformation = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_SETTINGS_BASIC'),
}
export const mayViewOrEditGatewayApiKeys = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_SETTINGS_API_KEYS'),
}
export const mayViewOrEditGatewayCollaborators = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_SETTINGS_COLLABORATORS'),
}
export const mayDeleteGateway = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_DELETE'),
}
export const mayViewGatewayEvents = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_TRAFFIC_READ'),
}
export const mayLinkGateway = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_LINK') && gsEnabled,
}
export const mayViewGatewayStatus = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_STATUS_READ') && gsEnabled,
}
export const mayViewOrEditGatewayLocation = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_LOCATION_READ'),
}
export const mayViewGatewayConfJson = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_INFO') && gcsEnabled,
}
export const mayViewGatewaySecrets = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_READ_SECRETS'),
}
export const mayEditGatewaySecrets = {
  rightsSelector: selectGatewayRights,
  check: rights => rights.includes('RIGHT_GATEWAY_WRITE_SECRETS'),
}

// Organization related feature checks.
export const mayViewOrganizationInformation = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_INFO'),
}
export const mayEditBasicOrganizationInformation = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_SETTINGS_BASIC'),
}
export const mayViewOrEditOrganizationApiKeys = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_SETTINGS_API_KEYS'),
}
export const mayViewOrEditOrganizationCollaborators = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_SETTINGS_MEMBERS'),
}
export const mayDeleteOrganization = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_DELETE'),
}
export const mayCreateApplicationsUnderOrganization = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_APPLICATIONS_CREATE'),
}
export const mayViewApplicationsOfOrganization = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_APPLICATIONS_LIST'),
}
export const mayCreateGatewaysUnderOrganization = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_GATEWAYS_CREATE'),
}
export const mayViewGatewaysOfOrganization = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_GATEWAYS_LIST'),
}
export const mayAddOrganizationAsCollaborator = {
  rightsSelector: selectOrganizationRights,
  check: rights => rights.includes('RIGHT_ORGANIZATION_ADD_AS_COLLABORATOR'),
}

// Admin feature checks.
export const mayPerformAdminActions = {
  rightsSelector: selectUserIsAdmin,
  check: isAdmin => isAdmin,
}

export const mayManageUsers = mayPerformAdminActions
export const mayPurgeEntities = mayPerformAdminActions
export const mayConfigurePacketBroker = mayPerformAdminActions

// Composite feature checks.
export const mayViewApplications = {
  rightsSelector: state => [...selectUserRights(state), ...selectOrganizationRights(state)],
  check: rights =>
    mayViewApplicationsOfUser.check(rights) || mayViewApplicationsOfOrganization.check(rights),
}
export const mayViewGateways = {
  rightsSelector: state => [...selectUserRights(state), ...selectOrganizationRights(state)],
  check: rights =>
    mayViewApplicationsOfUser.check(rights) || mayViewApplicationsOfOrganization.check(rights),
}
export const mayWriteTraffic = {
  rightsSelector: selectApplicationRights,
  check: rights => mayScheduleDownlinks.check(rights) || maySendUplink.check(rights),
}

// Pub/Sub feature checks.
export const mayAddPubSubIntegrations = {
  check: (natsDisabled, mqttDisabled) => !natsDisabled || !mqttDisabled,
}
