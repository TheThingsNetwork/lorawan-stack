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

import { selectSelectedGatewayId } from '@console/store/selectors/gateways'

const GtwSideNavigation = () => {
  const gtwId = useSelector(selectSelectedGatewayId)

  return (
    <SideNavigation>
      <DedicatedEntity label={gtwId} icon="arrow_left_alt" />
      <SideNavigation.Item title={sharedMessages.overview} path="" icon="overview" exact />
      <SideNavigation.Item title={sharedMessages.liveData} path="data" icon="data" />
      <SideNavigation.Item title={sharedMessages.location} path="location" icon="location" />
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

export default GtwSideNavigation
