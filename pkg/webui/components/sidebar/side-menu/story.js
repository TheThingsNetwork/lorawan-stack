// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import {
  IconDevices,
  IconCode,
  IconPuzzle,
  IconApiKeys,
  IconDownlink,
  IconGeneralSettings,
  IconIntegration,
  IconLiveData,
  IconOrganization,
  IconOverview,
  IconUplink,
} from '@ttn-lw/components/icon'

import SidebarContext from '@console/containers/sidebar/context'

import SideNavigationItem from './item'

import SideNavigation from '.'

export default {
  title: 'Navigation v2',
  component: SideNavigation,
  decorators: [
    storyFn => (
      <SidebarContext.Provider value={{ isMinimized: false }}>{storyFn()}</SidebarContext.Provider>
    ),
  ],
}

export const _SideNavigation = () => (
  <div style={{ width: '300px', height: '700px' }}>
    <SideNavigation>
      <SideNavigationItem title="Overview" path="/" icon={IconOverview} exact />
      <SideNavigationItem title="Devices" path="/devices" icon={IconDevices} />
      <SideNavigationItem title="Data" path="/data" icon={IconLiveData} />
      <SideNavigationItem title="Payload Formatters" icon={IconCode}>
        <SideNavigationItem title="Uplink" path="/payload-formatters/uplink" icon={IconUplink} />
        <SideNavigationItem
          title="Downlink"
          path="/payload-formatters/downlink"
          icon={IconDownlink}
        />
      </SideNavigationItem>
      <SideNavigationItem title="Integrations" icon={IconIntegration}>
        <SideNavigationItem title="MQTT" path="/integrations/mqtt" icon={IconPuzzle} />
        <SideNavigationItem title="Webhooks" path="/integrations/webhooks" icon={IconPuzzle} />
        <SideNavigationItem title="Pub/Subs" path="/integrations/pubsubs" icon={IconPuzzle} />
      </SideNavigationItem>
      <SideNavigationItem title="Collaborators" path="/collaborators" icon={IconOrganization} />
      <SideNavigationItem title="API keys" path="/api-keys" icon={IconApiKeys} />
      <SideNavigationItem
        title="General Settings"
        path="/general-settings"
        icon={IconGeneralSettings}
      />
    </SideNavigation>
  </div>
)
