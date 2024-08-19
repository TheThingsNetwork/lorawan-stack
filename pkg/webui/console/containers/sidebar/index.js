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

import React, { useCallback, useContext, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import classnames from 'classnames'
import { useDispatch } from 'react-redux'

import LAYOUT from '@ttn-lw/constants/layout'

import SearchButton from '@ttn-lw/components/sidebar/search-button'
import SideFooter from '@ttn-lw/components/sidebar/side-footer'

import PropTypes from '@ttn-lw/lib/prop-types'

import { setSearchOpen } from '@console/store/actions/search'

import SearchPanelManager from '../search-panel'

import SidebarNavigation from './navigation'
import SidebarContext from './context'
import SideHeader from './header'
import SwitcherContainer from './switcher'

import style from './sidebar.styl'

const Sidebar = ({ isDrawerOpen, isSideBarHovered, setIsHovered }) => {
  const { pathname } = useLocation()
  const { setIsMinimized, isMinimized, closeDrawer } = useContext(SidebarContext)
  const dispatch = useDispatch()

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

  useEffect(() => {
    let timer
    if (!isSideBarHovered && isMinimized && isDrawerOpen) {
      timer = setTimeout(() => {
        closeDrawer()
      }, 800)
    }
    return () => clearTimeout(timer)
  }, [isSideBarHovered, closeDrawer, isDrawerOpen, isMinimized])

  // Close the drawer on navigation changes after a small delay.
  useEffect(() => {
    const timer = setTimeout(() => {
      closeDrawer()
    }, 500)

    return () => clearTimeout(timer)
  }, [pathname, closeDrawer])

  const handleSearchClick = useCallback(() => {
    dispatch(setSearchOpen(true))
  }, [dispatch])

  const handleMouseEnter = useCallback(() => {
    setIsHovered(true)
  }, [setIsHovered])

  const handleMouseLeave = useCallback(() => {
    setIsHovered(false)
  }, [setIsHovered])

  const sidebarClassnames = classnames(
    style.sidebar,
    'd-flex direction-column j-between c-bg-brand-extralight gap-cs-l p-cs-m',
    {
      [style.sidebarMinimized]: isMinimized,
      [style.sidebarOpen]: isDrawerOpen,
    },
  )

  return (
    <>
      <div
        className={sidebarClassnames}
        id="sidebar"
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        <SideHeader />
        <div className="d-flex direction-column gap-cs-m">
          <SwitcherContainer />
          <SearchButton onClick={handleSearchClick} />
        </div>
        <SidebarNavigation />
        <SideFooter />
      </div>
      <div className={style.sidebarBackdrop} onClick={closeDrawer} />
      <SearchPanelManager />
    </>
  )
}

Sidebar.propTypes = {
  isDrawerOpen: PropTypes.bool.isRequired,
  isSideBarHovered: PropTypes.bool.isRequired,
  setIsHovered: PropTypes.func.isRequired,
}

export default Sidebar
