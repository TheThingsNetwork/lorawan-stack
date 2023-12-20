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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import classnames from 'classnames'

import LAYOUT from '@ttn-lw/constants/layout'

import SearchButton from '@ttn-lw/components/sidebar/search-button'

import SidebarNavigation from './navigation'
import SidebarContext from './context'
import SideHeader from './header'
import SideFooter from './footer'
import getCookie from './utils'
import SwitcherContainer from './switcher'

import style from './side-bar.styl'

const getViewportWidth = () =>
  Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0)

const Sidebar = () => {
  const viewportWidth = getViewportWidth()
  const isMobile = viewportWidth <= LAYOUT.BREAKPOINTS.M
  const { pathname } = useLocation()
  const [layer, setLayer] = useState(pathname ?? '/')
  const [isMinimized, setIsMinimized] = useState(false)
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const node = useRef()

  const onMinimizeToggle = useCallback(async () => {
    setIsMinimized(prev => !prev)
  }, [])

  const topEntitiesCookie = getCookie('topEntities')
    ? getCookie('topEntities')
        .split('_')
        .map(cookie => JSON.parse(cookie))
    : []

  const topEntities = topEntitiesCookie?.filter(cookie => cookie.tag === layer.split('/')[1])

  const sidebarClassnames = classnames(
    style.sidebar,
    'd-flex pos-fixed direction-column j-between gap-cs-m bg-tts-primary-050',
    {
      [style.sidebarMinimized]: isMinimized,
      [style.sidebarOpen]: isMobile && isDrawerOpen,
      'p-cs-s': !isMinimized,
      'p-vert-cs-m': isMinimized,
      'p-sides-cs-xs': isMinimized,
    },
  )

  const closeDrawer = useCallback(() => {
    setIsDrawerOpen(false)
    document.body.classList.remove(style.scrollLock)
  }, [])

  const openDrawer = useCallback(() => {
    setIsDrawerOpen(true)
    document.body.classList.add(style.scrollLock)
  }, [])

  useEffect(() => {
    const onClickOutside = e => {
      if (isDrawerOpen && node.current && !node.current.contains(e.target)) {
        closeDrawer()
      }
    }

    if (isDrawerOpen) {
      document.addEventListener('mousedown', onClickOutside)
      return () => document.removeEventListener('mousedown', onClickOutside)
    }
  }, [isDrawerOpen, closeDrawer])

  const onDrawerExpandClick = useCallback(() => {
    if (!isDrawerOpen) {
      openDrawer()
    } else {
      closeDrawer()
    }
  }, [isDrawerOpen, openDrawer, closeDrawer])

  // TODO: Add this function in the header component to close and open the drawer sidebar.
  // To be done after the merge of the header component.
  const onLeafItemClick = useCallback(() => {
    if (isDrawerOpen) {
      onDrawerExpandClick()
    }
  }, [isDrawerOpen, onDrawerExpandClick])

  return (
    <div className={sidebarClassnames} id="sidebar-v2">
      <SidebarContext.Provider
        value={{ layer, setLayer, topEntities, onMinimizeToggle, isMinimized }}
      >
        <div className="d-flex direction-column gap-cs-m">
          <SideHeader />
          <div>
            <SwitcherContainer />
            <SearchButton />
          </div>
          <SidebarNavigation />
        </div>
        <SideFooter />
      </SidebarContext.Provider>
    </div>
  )
}

export default Sidebar
