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

import PropTypes from '@ttn-lw/lib/prop-types'

import SidebarNavigation from './navigation'
import SidebarContext from './context'
import SideHeader from './header'
import SwitcherContainer from './switcher'

import style from './sidebar.styl'

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
  }, [setIsMinimized])

  // Close the drawer on navigation changes.
  useEffect(() => {
    onDrawerCloseClick()
  }, [pathname, onDrawerCloseClick])

  const onMinimizeToggle = useCallback(async () => {
    setIsMinimized(prev => !prev)
  }, [setIsMinimized])

  const sidebarClassnames = classnames(
    style.sidebar,
    'd-flex direction-column j-between c-bg-brand-extralight gap-cs-l',
    {
      [style.sidebarMinimized]: isMinimized,
      [style.sidebarOpen]: isDrawerOpen,
      'p-cs-m': !isMinimized,
      'pt-cs-m pb-cs-xs p-sides-cs-xs': isMinimized,
    },
  )

  return (
    <>
      <SidebarContext.Provider value={{ onMinimizeToggle, isMinimized, onDrawerCloseClick }}>
        <div className={sidebarClassnames} id="sidebar">
          <SideHeader />
          <div className="d-flex direction-column gap-cs-m">
            <SwitcherContainer />
            <SearchButton onClick={() => null} />
          </div>
          <SidebarNavigation />
          <SideFooter />
        </div>
      </SidebarContext.Provider>
      <div className={style.sidebarBackdrop} onClick={onDrawerCloseClick} />
    </>
  )
}

Sidebar.propTypes = {
  isDrawerOpen: PropTypes.bool.isRequired,
  onDrawerCloseClick: PropTypes.func.isRequired,
}

export default Sidebar
