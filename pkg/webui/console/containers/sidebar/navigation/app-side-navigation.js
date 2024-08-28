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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import { END_DEVICE } from '@console/constants/entities'

import {
  IconPuzzle,
  IconWebhook,
  IconApiKeys,
  IconCollaborators,
  IconDevice,
  IconDownlink,
  IconGeneralSettings,
  IconIntegration,
  IconLiveData,
  IconPayloadFormat,
  IconUplink,
  IconLayoutDashboard,
} from '@ttn-lw/components/icon'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'
import DedicatedEntity from '@ttn-lw/components/sidebar/dedicated-entity'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayViewApplicationInfo,
  mayViewApplicationEvents,
  maySetApplicationPayloadFormatters,
  mayViewApplicationDevices,
  mayCreateOrEditApplicationIntegrations,
  mayEditBasicApplicationInfo,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayViewOrEditApplicationPackages,
  mayAddPubSubIntegrations,
} from '@console/lib/feature-checks'

import {
  selectSelectedApplication,
  selectSelectedApplicationId,
  selectApplicationRights,
} from '@console/store/selectors/applications'
import {
  selectMqttProviderDisabled,
  selectNatsProviderDisabled,
} from '@console/store/selectors/application-server'
import { selectEndDeviceTopEntities } from '@console/store/selectors/top-entities'

import TopEntitiesSection from './top-entities-section'

const m = defineMessages({
  buttonMessage: 'Back to Applications list',
})

const AppSideNavigation = () => {
  const app = useSelector(selectSelectedApplication)
  const appId = useSelector(selectSelectedApplicationId)
  const rights = useSelector(selectApplicationRights)
  const natsDisabled = useSelector(selectNatsProviderDisabled)
  const mqttDisabled = useSelector(selectMqttProviderDisabled)
  const topEntityFilter = useCallback(e => e.id.startsWith(appId), [appId])
  const topEntities = useSelector(state => selectEndDeviceTopEntities(state, topEntityFilter))

  if (!app) {
    return null
  }

  const entityId = app && app.name ? app.name : appId

  return (
    <>
      <SideNavigation>
        <DedicatedEntity
          label={entityId}
          buttonMessage={m.buttonMessage}
          className="mt-cs-xs mb-cs-l"
          path={`/applications/${appId}`}
          backPath={`/applications`}
        />
        {mayViewApplicationInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.appOverview}
            path={`/applications/${appId}`}
            icon={IconLayoutDashboard}
            exact
          />
        )}
        {mayViewApplicationDevices.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.devices}
            path={`/applications/${appId}/devices`}
            icon={IconDevice}
          />
        )}
        {mayViewApplicationEvents.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.liveData}
            path={`/applications/${appId}/data`}
            icon={IconLiveData}
          />
        )}
        {maySetApplicationPayloadFormatters.check(rights) && (
          <SideNavigation.Item title={sharedMessages.payloadFormatters} icon={IconPayloadFormat}>
            <SideNavigation.Item
              title={sharedMessages.uplink}
              path={`/applications/${appId}/payload-formatters/uplink`}
              icon={IconUplink}
            />
            <SideNavigation.Item
              title={sharedMessages.downlink}
              path={`/applications/${appId}/payload-formatters/downlink`}
              icon={IconDownlink}
            />
          </SideNavigation.Item>
        )}
        {mayCreateOrEditApplicationIntegrations.check(rights) && (
          <SideNavigation.Item title={sharedMessages.integrations} icon={IconIntegration}>
            <SideNavigation.Item
              title={sharedMessages.mqtt}
              path={`/applications/${appId}/integrations/mqtt`}
              icon={IconPuzzle}
            />
            <SideNavigation.Item
              title={sharedMessages.webhooks}
              path={`/applications/${appId}/integrations/webhooks`}
              icon={IconWebhook}
            />
            {mayAddPubSubIntegrations.check(natsDisabled, mqttDisabled) && (
              <SideNavigation.Item
                title={sharedMessages.pubsubs}
                path={`/applications/${appId}/integrations/pubsubs`}
                icon={IconPuzzle}
              />
            )}
            {mayViewOrEditApplicationPackages.check(rights) && (
              <SideNavigation.Item
                title={sharedMessages.loraCloud}
                path={`/applications/${appId}/integrations/lora-cloud`}
                icon={IconPuzzle}
              />
            )}
          </SideNavigation.Item>
        )}
        {mayViewOrEditApplicationCollaborators.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            path={`/applications/${appId}/collaborators`}
            icon={IconCollaborators}
          />
        )}
        {mayViewOrEditApplicationApiKeys.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.apiKeys}
            path={`/applications/${appId}/api-keys`}
            icon={IconApiKeys}
          />
        )}
        {mayEditBasicApplicationInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.generalSettings}
            path={`/applications/${appId}/general-settings`}
            icon={IconGeneralSettings}
          />
        )}
      </SideNavigation>
      {mayViewApplicationInfo.check(rights) && (
        <TopEntitiesSection topEntities={topEntities} type={END_DEVICE} />
      )}
    </>
  )
}

export default AppSideNavigation
