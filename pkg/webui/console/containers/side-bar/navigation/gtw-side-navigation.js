// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useContext } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import SideNavigation from '@ttn-lw/components/sidebar/side-menu'
import DedicatedEntity from '@ttn-lw/components/sidebar/dedicated-entity'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayViewGatewayInfo,
  mayViewGatewayEvents,
  mayViewOrEditGatewayLocation,
  mayViewOrEditGatewayCollaborators,
  mayViewOrEditGatewayApiKeys,
  mayEditBasicGatewayInformation,
} from '@console/lib/feature-checks'

import {
  selectSelectedGateway,
  selectSelectedGatewayId,
  selectGatewayRights,
} from '@console/store/selectors/gateways'

import SidebarContext from '../context'

const m = defineMessages({
  buttonMessage: 'Back to Gateways list',
})

const GtwSideNavigation = () => {
  const gtw = useSelector(selectSelectedGateway)
  const gtwId = useSelector(selectSelectedGatewayId)
  const rights = useSelector(selectGatewayRights)
  const { isMinimized } = useContext(SidebarContext)

  const entityId = gtw ? gtw.name ?? gtwId : gtwId

  return (
    <SideNavigation className="mb-cs-m">
      {!isMinimized && (
        <DedicatedEntity
          label={entityId}
          buttonMessage={m.buttonMessage}
          className="mt-cs-xs mb-cs-l"
          backPath="/gateways"
          path={`/gateways/${gtwId}`}
        />
      )}
      {mayViewGatewayInfo.check(rights) && (
        <SideNavigation.Item
          title={sharedMessages.gatewayOverview}
          path={`gateways/${gtwId}`}
          icon="gateway"
          exact
        />
      )}
      {mayViewGatewayEvents.check(rights) && (
        <SideNavigation.Item
          title={sharedMessages.liveData}
          path={`gateways/${gtwId}/data`}
          icon="list_alt"
        />
      )}
      {mayViewOrEditGatewayLocation.check(rights) && (
        <SideNavigation.Item
          title={sharedMessages.location}
          path={`gateways/${gtwId}/location`}
          icon="map"
        />
      )}
      {mayViewOrEditGatewayCollaborators.check(rights) && (
        <SideNavigation.Item
          title={sharedMessages.collaborators}
          path={`gateways/${gtwId}/collaborators`}
          icon="organization"
        />
      )}
      {mayViewOrEditGatewayApiKeys.check(rights) && (
        <SideNavigation.Item
          title={sharedMessages.apiKeys}
          path={`gateways/${gtwId}/api-keys`}
          icon="api_keys"
        />
      )}
      {mayEditBasicGatewayInformation.check(rights) && (
        <SideNavigation.Item
          title={sharedMessages.generalSettings}
          path={`gateways/${gtwId}/general-settings`}
          icon="general_settings"
        />
      )}
    </SideNavigation>
  )
}

export default GtwSideNavigation
