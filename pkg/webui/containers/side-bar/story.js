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
import classNames from 'classnames'

import TtsLogo from '@assets/static/tts-logo.svg'

import Switcher from '@ttn-lw/components/switcher'
import SideNavigation from '@ttn-lw/components/navigation/side-v2'
import SideHeader from '@ttn-lw/components/side-header'
import SearchButton from '@ttn-lw/components/search-button'
import SideFooter from '@ttn-lw/components/side-footer'

import SidebarContext from '@ttn-lw/containers/side-bar/context'

import style from './side-bar.styl'

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
}

const SwictherExample = () => {
  const [layer, setLayer] = useState('/')
  const handleClick = useCallback(
    evt => {
      setLayer(evt.target.getAttribute('href'))
    },
    [setLayer],
  )

  return <Switcher layer={layer} onClick={handleClick} />
}

export const Default = () => (
  <div
    className={classNames(
      style.sidebar,
      'd-flex pos-relative direction-column gap-cs-s p-cs-s bg-tts-primary-050',
    )}
    style={{ width: '17rem', height: '96vh' }}
  >
    <SideHeader logo={{ src: TtsLogo, alt: 'logo' }} />
    <SwictherExample />
    <SearchButton />
    <SideNavigation className="mt-cs-xs">
      <SideNavigation.Item title={'Overview'} path="" icon="overview" exact />
      <SideNavigation.Item title={'Live data'} path="data" icon="data" />
      <SideNavigation.Item title={'Location'} path="location" icon="location" />
      <SideNavigation.Item title={'Collaborators'} path="collaborators" icon="organization" />
      <SideNavigation.Item title={'API keys'} path="api-keys" icon="api_keys" />
      <SideNavigation.Item
        title={'General settings'}
        path="general-settings"
        icon="general_settings"
      />
    </SideNavigation>
    <SideFooter
      supportLink={'/support'}
      documentationBaseUrl={'/docs'}
      statusPageBaseUrl={'/status'}
    />
  </div>
)
