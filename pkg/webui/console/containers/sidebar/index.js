// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import { setSearchOpen } from '@console/store/actions/search'

import SearchPanelManager from '../search-panel'

import SidebarNavigation from './navigation'
import SidebarContext from './context'
import SideHeader from './header'
import SwitcherContainer from './switcher'

import style from './sidebar.styl'

const Sidebar = () => {
  const { pathname } = useLocation()
  const {
    setIsMinimized,
    isMinimized,
    setIsDrawerOpen,
    closeDrawer,
    isDrawerOpen,
    isHovered,
    setIsHovered,
  } = useContext(SidebarContext)
  const dispatch = useDispatch()
  const node = React.useRef()

  const handleMouseMove = useCallback(
    e => {
      if (e.clientX <= 20 && isMinimized) {
        // If the mouse is within 20px of the left edge
        setIsDrawerOpen(true)
      } else if (e.clientX >= 550 && isMinimized) {
        // If the mouse is within 300px of the sidebar
        setIsDrawerOpen(false)
      }
    },
    [isMinimized, setIsDrawerOpen],
  )

  useEffect(() => {
    const onClickOutside = e => {
      if (isDrawerOpen && node.current && !node.current.contains(e.target)) {
        closeDrawer()
      }
    }

    if (isMinimized) {
      document.addEventListener('mousemove', handleMouseMove)
      return () => document.removeEventListener('mousemove', handleMouseMove)
    }

    if (isDrawerOpen) {
      document.addEventListener('mousedown', onClickOutside)
      return () => document.removeEventListener('mousedown', onClickOutside)
    }
  }, [isDrawerOpen, isMinimized, handleMouseMove, closeDrawer])

  // End of mobile side menu drawer functionality

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
    if (!isHovered && isMinimized && isDrawerOpen) {
      timer = setTimeout(() => {
        closeDrawer()
      }, 800)
    }
    return () => clearTimeout(timer)
  }, [isHovered, isDrawerOpen, isMinimized, closeDrawer])

  // Close the drawer on navigation changes after a small delay.
  useEffect(() => {
    const timer = setTimeout(() => {
      closeDrawer()
    }, 500)

    return () => clearTimeout(timer)
  }, [closeDrawer, pathname])

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
    'd-flex direction-column j-between c-bg-neutral-extralight gap-cs-l p-cs-m',
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
        ref={node}
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

export default Sidebar
