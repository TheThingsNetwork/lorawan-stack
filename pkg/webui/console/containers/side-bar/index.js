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

import React, { useCallback, useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import classnames from 'classnames'

import LAYOUT from '@ttn-lw/constants/layout'

import SearchButton from '@ttn-lw/components/sidebar/search-button'
import SideFooter from '@ttn-lw/components/sidebar/side-footer'

import getCookie from '@ttn-lw/lib/cookie'
import PropTypes from '@ttn-lw/lib/prop-types'

import SidebarNavigation from './navigation'
import SidebarContext from './context'
import SideHeader from './header'
import SwitcherContainer from './switcher'

import style from './side-bar.styl'

const Sidebar = ({ isDrawerOpen, onDrawerCloseClick }) => {
  const { pathname } = useLocation()
  const [isMinimized, setIsMinimized] = useState(false)

  // Reset minimized state when screen size changes to mobile.
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth < LAYOUT.BREAKPOINTS.M) {
        setIsMinimized(false)
      }
    }

    window.addEventListener('resize', handleResize)

    return () => window.removeEventListener('resize', handleResize)
  }, [])

  // Close the drawer on navigation changes.
  useEffect(() => {
    onDrawerCloseClick()
  }, [pathname, onDrawerCloseClick])

  const onMinimizeToggle = useCallback(async () => {
    setIsMinimized(prev => !prev)
  }, [])

  const topEntitiesCookie = getCookie('topEntities')
    ? getCookie('topEntities')
        .split('_')
        .map(cookie => JSON.parse(cookie))
    : []

  const tag = pathname === '/' ? 'general' : pathname.split('/')[1]
  const topEntities = topEntitiesCookie?.filter(cookie => cookie.tag === tag)

  const sidebarClassnames = classnames(
    style.sidebar,
    'd-flex direction-column j-between gap-cs-m bg-tts-primary-050',
    {
      [style.sidebarMinimized]: isMinimized,
      [style.sidebarOpen]: isDrawerOpen,
      'p-cs-m': !isMinimized,
      'p-vert-cs-s': isMinimized,
      'p-sides-cs-xs': isMinimized,
    },
  )

  return (
    <>
      <div className={sidebarClassnames} id="sidebar">
        <SidebarContext.Provider
          value={{ topEntities, onMinimizeToggle, isMinimized, onDrawerCloseClick }}
        >
          <div className="d-flex direction-column gap-cs-l">
            <SideHeader />
            <div className="d-flex direction-column gap-cs-m">
              <SwitcherContainer />
              <SearchButton onClick={() => null} />
            </div>
            <SidebarNavigation />
          </div>
          <SideFooter />
        </SidebarContext.Provider>
      </div>
      <div className={style.sidebarBackdrop} onClick={onDrawerCloseClick} />
    </>
  )
}

Sidebar.propTypes = {
  isDrawerOpen: PropTypes.bool.isRequired,
  onDrawerCloseClick: PropTypes.func.isRequired,
}

export default Sidebar
