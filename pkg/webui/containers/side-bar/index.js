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

import React, { useCallback, useState } from 'react'
import { useLocation } from 'react-router-dom'
import classNames from 'classnames'

import SearchButton from '@ttn-lw/components/search-button'

import SideBarNavigation from './navigation'
import SideBarContext from './context'
import SideHeader from './header'
import SideFooter from './footer'
import getCookie from './utils'
import SwitcherContainer from './switcher'

import style from './side-bar.styl'

const SideBar = () => {
  const { pathname } = useLocation()
  const [layer, setLayer] = useState(pathname ?? '/')
  const [isMinimized, setIsMinimized] = useState(false)

  const onMinimizeToggle = useCallback(async () => {
    setIsMinimized(prev => !prev)
  }, [])

  const topEntitiesCookie = getCookie('topEntities')
    ? getCookie('topEntities')
        .split('_')
        .map(cookie => JSON.parse(cookie))
    : []

  const topEntities = topEntitiesCookie?.filter(cookie => cookie.tag === layer.split('/')[1])

  return (
    <div
      className={classNames(
        style.sidebar,
        'd-flex pos-fixed align-center direction-column gap-cs-s bg-tts-primary-050',
        { [style.sidebarMinimized]: isMinimized, 'p-cs-s': !isMinimized, 'p-cs-xs': isMinimized },
      )}
      id="sidebar-v2"
    >
      <SideBarContext.Provider
        value={{ layer, setLayer, topEntities, onMinimizeToggle, isMinimized }}
      >
        <SideHeader />
        <SwitcherContainer />
        <SearchButton />
        <SideBarNavigation />
        <SideFooter />
      </SideBarContext.Provider>
    </div>
  )
}

export default SideBar
