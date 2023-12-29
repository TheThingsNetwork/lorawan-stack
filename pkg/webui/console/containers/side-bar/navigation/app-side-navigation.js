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

import SideNavigation from '@ttn-lw/components/navigation/side'
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

import SidebarContext from '../context'

const m = defineMessages({
  buttonMessage: 'Back to Applications list',
})

const AppSideNavigation = () => {
  const app = useSelector(selectSelectedApplication)
  const appId = useSelector(selectSelectedApplicationId)
  const rights = useSelector(selectApplicationRights)
  const natsDisabled = useSelector(selectNatsProviderDisabled)
  const mqttDisabled = useSelector(selectMqttProviderDisabled)
  const { isMinimized } = useContext(SidebarContext)

  const entityId = app ? app.name ?? appId : appId

  return (
    <>
      <SideNavigation className="mb-cs-m">
        {!isMinimized && (
          <DedicatedEntity
            label={entityId}
            buttonMessage={m.buttonMessage}
            icon="arrow_back"
            className="mt-cs-xs mb-cs-l"
            path={`/applications`}
          />
        )}
        {mayViewApplicationInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.appOverview}
            path={`applications/${appId}`}
            icon="group"
            exact
          />
        )}
        {mayViewApplicationDevices.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.devices}
            path={`applications/${appId}/devices`}
            icon="device"
          />
        )}
        {mayViewApplicationEvents.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.liveData}
            path={`applications/${appId}/data`}
            icon="list_alt"
          />
        )}
        {/* <SideNavigation.Item title={'Network Information Center'} path="/noc" icon="ssid_chart" /> */}
        {maySetApplicationPayloadFormatters.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.payloadFormatters}
            icon="developer_mode"
            isMinimized={isMinimized}
          >
            <SideNavigation.Item
              title={sharedMessages.uplink}
              path={`applications/${appId}/payload-formatters/uplink`}
              icon="uplink"
            />
            <SideNavigation.Item
              title={sharedMessages.downlink}
              path={`applications/${appId}/payload-formatters/downlink`}
              icon="downlink"
            />
          </SideNavigation.Item>
        )}
        {mayCreateOrEditApplicationIntegrations.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.integrations}
            icon="integration"
            isMinimized={isMinimized}
          >
            <SideNavigation.Item
              title={sharedMessages.mqtt}
              path={`applications/${appId}/integrations/mqtt`}
              icon="extension"
            />
            <SideNavigation.Item
              title={sharedMessages.webhooks}
              path={`applications/${appId}/integrations/webhooks`}
              icon="extension"
            />
            {mayAddPubSubIntegrations.check(natsDisabled, mqttDisabled) && (
              <SideNavigation.Item
                title={sharedMessages.pubsubs}
                path={`applications/${appId}/integrations/pubsubs`}
                icon="extension"
              />
            )}
            {mayViewOrEditApplicationPackages.check(rights) && (
              <SideNavigation.Item
                title={sharedMessages.loraCloud}
                path={`applications/${appId}/integrations/lora-cloud`}
                icon="extension"
              />
            )}
          </SideNavigation.Item>
        )}
        {mayViewOrEditApplicationCollaborators.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            path={`applications/${appId}/collaborators`}
            icon="organization"
          />
        )}
        {mayViewOrEditApplicationApiKeys.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.apiKeys}
            path={`applications/${appId}/api-keys`}
            icon="api_keys"
          />
        )}
        {mayEditBasicApplicationInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.generalSettings}
            path={`applications/${appId}/general-settings`}
            icon="general_settings"
          />
        )}
      </SideNavigation>
    </>
  )
}

export default AppSideNavigation
