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

import React from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'
import { GATEWAY } from '@console/constants/entities'

import {
  IconMap,
  IconApiKeys,
  IconGateway,
  IconGeneralSettings,
  IconLiveData,
  IconOrganization,
} from '@ttn-lw/components/icon'
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
import getCookie from '@console/lib/table-utils'

import {
  selectSelectedGateway,
  selectSelectedGatewayId,
  selectGatewayRights,
} from '@console/store/selectors/gateways'
import { selectGatewayTopEntities } from '@console/store/selectors/top-entities'

import TopEntitiesSection from './top-entities-section'

const m = defineMessages({
  buttonMessage: 'Back to Gateways list',
})

const GtwSideNavigation = () => {
  const gtw = useSelector(selectSelectedGateway)
  const gtwId = useSelector(selectSelectedGatewayId)
  const rights = useSelector(selectGatewayRights)
  const gtwPageSize = getCookie('gateways-list-page-size')
  const gtwParam = `?page-size=${gtwPageSize ? gtwPageSize : PAGE_SIZES.REGULAR}`
  const topEntities = useSelector(selectGatewayTopEntities)

  if (!gtw) {
    return null
  }

  const entityId = gtw && gtw.name ? gtw.name : gtwId

  return (
    <>
      <SideNavigation>
        <DedicatedEntity
          label={entityId}
          buttonMessage={m.buttonMessage}
          className="mt-cs-xs mb-cs-l"
          backPath={`/gateways${gtwParam}`}
          path={`/gateways/${gtwId}`}
        />
        {mayViewGatewayInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.gatewayOverview}
            path={`/gateways/${gtwId}`}
            icon={IconGateway}
            exact
          />
        )}
        {mayViewGatewayEvents.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.liveData}
            path={`/gateways/${gtwId}/data`}
            icon={IconLiveData}
          />
        )}
        {mayViewOrEditGatewayLocation.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.location}
            path={`/gateways/${gtwId}/location`}
            icon={IconMap}
          />
        )}
        {mayViewOrEditGatewayCollaborators.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            path={`/gateways/${gtwId}/collaborators`}
            icon={IconOrganization}
          />
        )}
        {mayViewOrEditGatewayApiKeys.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.apiKeys}
            path={`/gateways/${gtwId}/api-keys`}
            icon={IconApiKeys}
          />
        )}
        {mayEditBasicGatewayInformation.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.generalSettings}
            path={`/gateways/${gtwId}/general-settings`}
            icon={IconGeneralSettings}
          />
        )}
      </SideNavigation>
      {mayViewGatewayInfo.check(rights) && (
        <TopEntitiesSection topEntities={topEntities} type={GATEWAY} />
      )}
    </>
  )
}

export default GtwSideNavigation
