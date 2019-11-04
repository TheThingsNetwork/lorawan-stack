// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import { selectStackConfig } from '../../lib/selectors/env'
import { selectApplicationRights } from '../store/selectors/applications'

const stackConfig = selectStackConfig()
const asEnabled = stackConfig.as.enabled
const jsEnabled = stackConfig.js.enabled
const nsEnabled = stackConfig.ns.enabled

// Applications
export const mayViewApplicationInfo = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_INFO'),
}
export const mayEditBasicApplicationInfo = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_SETTINGS_BASIC'),
}
export const mayLinkApplication = {
  rightsSelector: selectApplicationRights,
  check: rights => rights.includes('RIGHT_APPLICATION_LINK') && asEnabled,
}
export const maySetApplicationPayloadFormatters = {
  rightsSelector: selectApplicationRights,
  check: mayLinkApplication.check,
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
