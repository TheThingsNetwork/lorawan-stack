// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import {
  IconCpu,
  IconDeviceDesktopAnalytics,
  IconRouter,
  IconUsersGroup,
} from '@tabler/icons-react'

import { APPLICATION, GATEWAY, END_DEVICE, ORGANIZATION } from '@console/constants/entities'

export {
  IconLock as IconAccess,
  IconUserShield as IconAdminPanel,
  IconKey as IconApiKeys,
  IconDeviceDesktopAnalytics as IconApplication,
  IconWorld as IconCluster,
  IconUsers as IconCollaborators,
  IconCode as IconDevelop,
  IconCpu as IconDevice,
  IconArrowDown as IconDownlink,
  IconInfoCircle as IconEvent,
  IconClearAll as IconEventClearAll,
  IconTransfer as IconEventConnection,
  IconCirclePlus as IconEventCreate,
  IconTrash as IconEventDelete,
  IconArrowDown as IconEventDownlink,
  IconExclamationCircle as IconEventError,
  IconBolt as IconEventGatewayConnect,
  IconBoltOff as IconEventGatewayDisconnect,
  IconCirclesRelation as IconEventJoin,
  IconAdjustmentsHorizontal as IconEventMode,
  IconKey as IconEventRekey,
  IconHeartbeat as IconEventStatus,
  IconSwitch as IconEventSwitch,
  IconEdit as IconEventUpdate,
  IconArrowUp as IconEventUplink,
  IconArrowDown as IconExpandDown,
  IconArrowUp as IconExpandUp,
  IconRouter as IconGateway,
  IconSettings as IconGeneralSettings,
  IconPlaylistAdd as IconImportDevices,
  IconArrowMergeAltRight as IconIntegration,
  IconCirclesRelation as IconJoin,
  IconArticle as IconLiveData,
  IconUsersGroup as IconOrganization,
  IconLayoutDashboard as IconOverview,
  IconAperture as IconPacketBroker,
  IconSourceCode as IconPayloadFormat,
  IconBrandOauth as IconOauthClients,
  IconLifebuoy as IconSupport,
  IconSelector as IconSort,
  IconSortAscending as IconSortOrderAsc,
  IconSortDescending as IconSortOrderDesc,
  IconArrowUp as IconUplink,
  IconUserCog as IconUserManagement,
  IconCircleCheck as IconValid,
} from '@tabler/icons-react'

export const entityIcons = {
  [APPLICATION]: IconDeviceDesktopAnalytics,
  [GATEWAY]: IconRouter,
  [END_DEVICE]: IconCpu,
  [ORGANIZATION]: IconUsersGroup,
}
