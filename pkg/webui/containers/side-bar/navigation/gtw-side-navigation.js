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

import { selectSelectedGateway, selectSelectedGatewayId } from '@console/store/selectors/gateways'

import SidebarContext from '../context'

const m = defineMessages({
  buttonMessage: 'Back to Gateways list',
})

const GtwSideNavigation = () => {
  const gtw = useSelector(selectSelectedGateway)
  const gtwId = useSelector(selectSelectedGatewayId)
  const { isMinimized, setLayer } = useContext(SidebarContext)
  const navigate = useNavigate()

  const entityId = gtw ? gtw.name ?? gtwId : gtwId

  const handleBackClick = useCallback(() => {
    const path = '/gateways'
    navigate(path)
    setLayer(path)
  }, [navigate, setLayer])

  return (
    <SideNavigation>
      {!isMinimized && (
        <DedicatedEntity
          label={entityId}
          buttonMessage={m.buttonMessage}
          icon="arrow_left_alt"
          className="mt-cs-xs mb-cs-m"
          onClick={handleBackClick}
        />
      )}
      <SideNavigation.Item
        title={sharedMessages.gatewayOverview}
        path={`gateways/${gtwId}`}
        icon="gateway"
        exact
      />
      <SideNavigation.Item
        title={sharedMessages.liveData}
        path={`gateways/${gtwId}/data`}
        icon="list_alt"
      />
      <SideNavigation.Item
        title={sharedMessages.location}
        path={`gateways/${gtwId}/location`}
        icon="map"
      />
      <SideNavigation.Item
        title={sharedMessages.collaborators}
        path={`gateways/${gtwId}/collaborators`}
        icon="organization"
      />
      <SideNavigation.Item
        title={sharedMessages.apiKeys}
        path={`gateways/${gtwId}/api-keys`}
        icon="api_keys"
      />
      <SideNavigation.Item
        title={sharedMessages.generalSettings}
        path={`gateways/${gtwId}/general-settings`}
        icon="general_settings"
      />
    </SideNavigation>
  )
}

export default GtwSideNavigation
