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

import SideNavigation from '@ttn-lw/components/navigation/side-v2'
import DedicatedEntity from '@ttn-lw/components/dedicated-entity'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectSelectedDeviceId } from '@console/store/selectors/devices'

import SideBarContext from '../context'

const AppSideNavigation = () => {
  const { layer } = React.useContext(SideBarContext)
  const appId = useSelector(selectSelectedApplicationId)
  const deviceId = useSelector(selectSelectedDeviceId)

  const entityId = React.useMemo(() => {
    if (layer.includes('/devices/')) {
      return deviceId
    }

    return appId
  }, [appId, deviceId, layer])

  return (
    <SideNavigation>
      <DedicatedEntity label={entityId} icon="arrow_left_alt" />
      <SideNavigation.Item title={sharedMessages.overview} path="" icon="overview" exact />
      <SideNavigation.Item title={sharedMessages.devices} path="devices" icon="devices" />
      <SideNavigation.Item title={sharedMessages.liveData} path="data" icon="data" />
      <SideNavigation.Item title={sharedMessages.payloadFormatters} icon="code">
        <SideNavigation.Item
          title={sharedMessages.uplink}
          path="payload-formatters/uplink"
          icon="uplink"
        />
        <SideNavigation.Item
          title={sharedMessages.downlink}
          path="payload-formatters/downlink"
          icon="downlink"
        />
      </SideNavigation.Item>
      <SideNavigation.Item title={sharedMessages.integrations} icon="integration">
        <SideNavigation.Item
          title={sharedMessages.mqtt}
          path="integrations/mqtt"
          icon="extension"
        />
        <SideNavigation.Item
          title={sharedMessages.webhooks}
          path="integrations/webhooks"
          icon="extension"
        />
        <SideNavigation.Item
          title={sharedMessages.pubsubs}
          path="integrations/pubsubs"
          icon="extension"
        />
        <SideNavigation.Item
          title={sharedMessages.loraCloud}
          path="integrations/lora-cloud"
          icon="extension"
        />
      </SideNavigation.Item>
      <SideNavigation.Item
        title={sharedMessages.collaborators}
        path="collaborators"
        icon="organization"
      />
      <SideNavigation.Item title={sharedMessages.apiKeys} path="api-keys" icon="api_keys" />
      <SideNavigation.Item
        title={sharedMessages.generalSettings}
        path="general-settings"
        icon="general_settings"
      />
    </SideNavigation>
  )
}

export default AppSideNavigation
