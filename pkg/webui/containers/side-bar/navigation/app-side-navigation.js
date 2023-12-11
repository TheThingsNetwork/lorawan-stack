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

import SideNavigation from '@ttn-lw/components/navigation/side-v2'
import DedicatedEntity from '@ttn-lw/components/dedicated-entity'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  selectSelectedApplication,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

import SidebarContext from '../context'

const AppSideNavigation = () => {
  const app = useSelector(selectSelectedApplication)
  const appId = useSelector(selectSelectedApplicationId)
  const { isMinimized } = useContext(SidebarContext)

  const entityId = app ? app.name ?? appId : appId

  return (
    <>
      <SideNavigation>
        {!isMinimized && (
          <DedicatedEntity label={entityId} icon="arrow_left_alt" className="mt-cs-xs mb-cs-m" />
        )}
        <SideNavigation.Item
          title={sharedMessages.appOverview}
          path={`applications/${appId}`}
          icon="group"
          exact
        />
        <SideNavigation.Item
          title={sharedMessages.devices}
          path={`applications/${appId}/devices`}
          icon="device"
        />
        <SideNavigation.Item
          title={sharedMessages.liveData}
          path={`applications/${appId}/data`}
          icon="list_alt"
        />
        {/* <SideNavigation.Item title={'Network Information Center'} path="/noc" icon="ssid_chart" /> */}
        <SideNavigation.Item title={sharedMessages.payloadFormatters} icon="developer_mode">
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
        <SideNavigation.Item title={sharedMessages.integrations} icon="integration">
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
          <SideNavigation.Item
            title={sharedMessages.pubsubs}
            path={`applications/${appId}/integrations/pubsubs`}
            icon="extension"
          />
          <SideNavigation.Item
            title={sharedMessages.loraCloud}
            path={`applications/${appId}/integrations/lora-cloud`}
            icon="extension"
          />
        </SideNavigation.Item>
        <SideNavigation.Item
          title={sharedMessages.collaborators}
          path={`applications/${appId}/collaborators`}
          icon="organization"
        />
        <SideNavigation.Item
          title={sharedMessages.apiKeys}
          path={`applications/${appId}/api-keys`}
          icon="api_keys"
        />
        <SideNavigation.Item
          title={sharedMessages.generalSettings}
          path={`applications/${appId}/general-settings`}
          icon="general_settings"
        />
      </SideNavigation>
    </>
  )
}

export default AppSideNavigation
