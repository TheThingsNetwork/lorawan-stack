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

import ReactDom from 'react-dom'
import React, { useState, useEffect, useCallback, useRef } from 'react'
import classnames from 'classnames'
import { defineMessages, useIntl } from 'react-intl'

import LAYOUT from '@ttn-lw/constants/layout'

import Button from '@ttn-lw/components/button'
import Icon from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import SideNavigationList from './list'
import SideNavigationItem from './item'
import SideNavigationContext from './context'

import style from './side.styl'

const getViewportWidth = () =>
  Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0)

const m = defineMessages({
  hideSidebar: 'Hide sidebar',
})

const SideNavigation = ({
  appContainerId,
  modifyAppContainerClasses,
  className,
  header,
  children,
}) => {
  const [isMinimized, setIsMinimized] = useState(getViewportWidth() <= LAYOUT.BREAKPOINTS.M)
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const [preferMinimized, setPreferMinimized] = useState(false)
  const node = useRef()
  const intl = useIntl()

  const updateAppContainerClasses = useCallback(
    (initial = false) => {
      if (!modifyAppContainerClasses) {
        return
      }
      const containerClasses = document.getElementById(appContainerId).classList
      containerClasses.add('with-sidebar')
      if (!initial) {
        containerClasses.add('sidebar-transitioned')
      }
      if (isMinimized) {
        containerClasses.add('sidebar-minimized')
      } else {
        containerClasses.remove('sidebar-minimized')
      }
    },
    [modifyAppContainerClasses, appContainerId, isMinimized],
  )

  const removeAppContainerClasses = useCallback(() => {
    if (!modifyAppContainerClasses) {
      return
    }
    document
      .getElementById(appContainerId)
      .classList.remove('with-sidebar', 'sidebar-minimized', 'sidebar-transitioned')
  }, [modifyAppContainerClasses, appContainerId])

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

  const setMinimizedState = useCallback(() => {
    const viewportWidth = getViewportWidth()
    if (
      (!isMinimized && viewportWidth <= LAYOUT.BREAKPOINTS.M) ||
      (isMinimized && viewportWidth > LAYOUT.BREAKPOINTS.M)
    ) {
      setIsMinimized(getViewportWidth() <= LAYOUT.BREAKPOINTS.M || preferMinimized)
      updateAppContainerClasses()
    }
  }, [isMinimized, preferMinimized, updateAppContainerClasses])

  useEffect(() => {
    window.addEventListener('resize', setMinimizedState)
    updateAppContainerClasses(true)
    return () => {
      window.removeEventListener('resize', setMinimizedState)
      removeAppContainerClasses()
    }
  }, [removeAppContainerClasses, setMinimizedState, updateAppContainerClasses])

  const onToggle = useCallback(async () => {
    setIsMinimized(prev => !prev)
    setPreferMinimized(prev => !prev)
    updateAppContainerClasses()
  }, [updateAppContainerClasses])

  const onDrawerExpandClick = useCallback(() => {
    if (!isDrawerOpen) {
      openDrawer()
    } else {
      closeDrawer()
    }
  }, [isDrawerOpen, openDrawer, closeDrawer])

  const onLeafItemClick = useCallback(() => {
    if (isDrawerOpen) {
      onDrawerExpandClick()
    }
  }, [isDrawerOpen, onDrawerExpandClick])

  const navigationClassNames = classnames(className, style.navigation, {
    [style.navigationMinimized]: isMinimized,
  })
  const minimizeButtonClassNames = classnames(style.minimizeButton, {
    [style.minimizeButtonMinimized]: isMinimized,
  })

  const drawerClassNames = classnames(style.drawer, { [style.drawerOpen]: isDrawerOpen })

  return (
    <>
      <nav className={navigationClassNames} ref={node} data-test-id="navigation-sidebar">
        <div className={style.mobileHeader} onClick={onDrawerExpandClick}>
          <Icon className={style.expandIcon} icon="more_vert" />
          <img
            className={style.icon}
            src={header.icon}
            alt={intl.formatMessage({
              id: `${header.iconAlt}-alt`,
              defaultMessage: header.iconAlt,
            })}
          />
          <Message className={style.message} content={header.title} />
        </div>
        <div>
          <div className={drawerClassNames}>
            <Link to={header.to}>
              <div className={style.header}>
                <img
                  className={style.icon}
                  src={header.icon}
                  alt={intl.formatMessage({
                    id: `${header.iconAlt}-alt`,
                    defaultMessage: header.iconAlt,
                  })}
                />
                <Message className={style.message} content={header.title} />
              </div>
            </Link>
            <SideNavigationContext.Provider value={{ isMinimized, onLeafItemClick }}>
              <SideNavigationList
                onListClick={onDrawerExpandClick}
                isMinimized={isMinimized}
                className={style.navigationList}
              >
                {children}
              </SideNavigationList>
            </SideNavigationContext.Provider>
          </div>
        </div>
      </nav>
      <Button
        unstyled
        className={minimizeButtonClassNames}
        icon={isMinimized ? 'keyboard_arrow_right' : 'keyboard_arrow_left'}
        message={isMinimized ? null : m.hideSidebar}
        onClick={onToggle}
        data-hook="side-nav-hide-button"
      />
    </>
  )
}

SideNavigation.propTypes = {
  appContainerId: PropTypes.string,
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  /** The header for the side navigation. */
  header: PropTypes.shape({
    title: PropTypes.string.isRequired,
    icon: PropTypes.string.isRequired,
    iconAlt: PropTypes.message.isRequired,
    to: PropTypes.string.isRequired,
  }).isRequired,
  modifyAppContainerClasses: PropTypes.bool,
}

SideNavigation.defaultProps = {
  appContainerId: 'app',
  modifyAppContainerClasses: true,
  className: undefined,
}

const PortalledSideNavigation = props =>
  ReactDom.createPortal(<SideNavigation {...props} />, document.getElementById('sidebar'))

PortalledSideNavigation.Item = SideNavigationItem

export { PortalledSideNavigation as default, SideNavigation }
