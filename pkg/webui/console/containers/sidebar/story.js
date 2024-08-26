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

import React, { useCallback, useState } from 'react'
import classnames from 'classnames'

import TtsLogo from '@assets/static/tts-logo.svg'

import {
  IconData,
  IconApiKeys,
  IconGeneralSettings,
  IconMap,
  IconOrganization,
  IconOverview,
} from '@ttn-lw/components/icon'
import Switcher from '@ttn-lw/components/sidebar/switcher'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'
import SideHeader from '@ttn-lw/components/sidebar/side-header'
import SearchButton from '@ttn-lw/components/sidebar/search-button'
import SideFooter from '@ttn-lw/components/sidebar/side-footer'

import SidebarContext from '@console/containers/sidebar/context'

import style from './sidebar.styl'

export default {
  title: 'Sidebar/Sidebar',
  decorators: [
    storyFn => {
      const [isMinimized, setIsMinimized] = useState(false)

      const onMinimizeToggle = useCallback(async () => {
        setIsMinimized(prev => !prev)
      }, [])
      return (
        <SidebarContext.Provider value={{ isMinimized, onMinimizeToggle }}>
          {storyFn()}
        </SidebarContext.Provider>
      )
    },
  ],
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/7pBLWK4tsjoAbyJq2viMAQ/2023-console-redesign?type=design&node-id=1293%3A8589&mode=design&t=Hbk2Qngeg1xqg4V3-1',
    },
  },
}

export const Default = () => (
  <div
    className={classnames(
      style.sidebar,
      'd-flex pos-relative direction-column gap-cs-s p-cs-s bg-tts-primary-050',
    )}
    style={{ width: '17rem', height: '96vh' }}
  >
    <SideHeader logo={{ src: TtsLogo, alt: 'logo' }} />
    <Switcher />
    <SearchButton />
    <SideNavigation className="mt-cs-xs">
      <SideNavigation.Item title="Overview" path="" icon={IconOverview} exact />
      <SideNavigation.Item title="Live data" path="data" icon={IconData} />
      <SideNavigation.Item title="Location" path="location" icon={IconMap} />
      <SideNavigation.Item title="Collaborators" path="collaborators" icon={IconOrganization} />
      <SideNavigation.Item title="API keys" path="api-keys" icon={IconApiKeys} />
      <SideNavigation.Item
        title="General settings"
        path="general-settings"
        icon={IconGeneralSettings}
      />
    </SideNavigation>
    <SideFooter supportLink="/support" documentationBaseUrl="/docs" statusPageBaseUrl="/status" />
  </div>
)
