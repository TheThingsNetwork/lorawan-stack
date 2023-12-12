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

import React, { useCallback, useContext } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import SideNavigation from '@ttn-lw/components/navigation/side-v2'
import DedicatedEntity from '@ttn-lw/components/dedicated-entity'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import SidebarContext from '../context'

const m = defineMessages({
  buttonMessage: 'Back to Devices list',
})

const DeviceSideNavigation = () => {
  const device = useSelector(selectSelectedDevice)
  const deviceId = useSelector(selectSelectedDeviceId)
  const appId = useSelector(selectSelectedApplicationId)
  const { isMinimized, setLayer } = useContext(SidebarContext)
  const navigate = useNavigate()

  const entityId = device ? device.name ?? deviceId : deviceId

  const handleBackClick = useCallback(() => {
    const path = `/applications/${appId}/devices`
    navigate(path)
    setLayer(path)
  }, [navigate, appId, setLayer])

  return (
    <SideNavigation>
      {!isMinimized && (
        <DedicatedEntity
          label={entityId}
          entityIcon="device"
          icon="arrow_left_alt"
          className="mt-cs-xs mb-cs-m"
          buttonMessage={m.buttonMessage}
          onClick={handleBackClick}
        />
      )}
      <SideNavigation.Item
        title={sharedMessages.endDeviceOverview}
        path={`applications/${appId}/devices/${deviceId}`}
        icon="grid_view"
        exact
      />
      <SideNavigation.Item
        title={sharedMessages.liveData}
        path={`applications/${appId}/devices/${deviceId}/data`}
        icon="list_alt"
      />
      <SideNavigation.Item
        title={'Messaging'}
        path={`applications/${appId}/devices/${deviceId}/messaging`}
        icon="swap_vert"
      />
      <SideNavigation.Item
        title={sharedMessages.location}
        path={`applications/${appId}/devices/${deviceId}/location`}
        icon="map"
      />
      <SideNavigation.Item title={sharedMessages.payloadFormatters} icon="developer_mode">
        <SideNavigation.Item
          title={sharedMessages.uplink}
          path={`applications/${appId}/devices/${deviceId}/payload-formatters/uplink`}
          icon="uplink"
        />
        <SideNavigation.Item
          title={sharedMessages.downlink}
          path={`applications/${appId}/devices/${deviceId}/payload-formatters/downlink`}
          icon="downlink"
        />
      </SideNavigation.Item>
      <SideNavigation.Item
        title={sharedMessages.generalSettings}
        path={`applications/${appId}/devices/${deviceId}/general-settings`}
        icon="general_settings"
      />
    </SideNavigation>
  )
}

export default DeviceSideNavigation
