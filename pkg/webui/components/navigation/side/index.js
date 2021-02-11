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

import ReactDom from 'react-dom'
import React, { Component } from 'react'
import bind from 'autobind-decorator'
import classnames from 'classnames'
import { defineMessages } from 'react-intl'

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
export class SideNavigation extends Component {
  static propTypes = {
    appContainerId: PropTypes.string,
    children: PropTypes.node.isRequired,
    className: PropTypes.string,
    /** The header for the side navigation. */
    header: PropTypes.shape({
      title: PropTypes.string.isRequired,
      icon: PropTypes.string.isRequired,
      to: PropTypes.string.isRequired,
    }).isRequired,
    modifyAppContainerClasses: PropTypes.bool,
  }

  static defaultProps = {
    appContainerId: 'app',
    modifyAppContainerClasses: true,
    className: undefined,
  }

  state = {
    /** A flag specifying whether the side navigation is minimized or not. */
    isMinimized: getViewportWidth() <= LAYOUT.BREAKPOINTS.M,
    /** A flag specifying whether the drawer is currently open (in mobile
     * screensizes).
     */
    isDrawerOpen: false,
    /** A flag indicating whether the user has last toggled the sidebar to
     * minimized state. */
    preferMinimized: false,
  }

  @bind
  updateAppContainerClasses(initial = false) {
    const { modifyAppContainerClasses, appContainerId } = this.props
    if (!modifyAppContainerClasses) {
      return
    }
    const { isMinimized } = this.state
    const containerClasses = document.getElementById(appContainerId).classList
    containerClasses.add('with-sidebar')
    if (!initial) {
      // The transitioned class is necessary to prevent unwanted width
      // transitions during route changes.
      containerClasses.add('sidebar-transitioned')
    }
    if (isMinimized) {
      containerClasses.add('sidebar-minimized')
    } else {
      containerClasses.remove('sidebar-minimized')
    }
  }

  @bind
  removeAppContainerClasses() {
    const { modifyAppContainerClasses, appContainerId } = this.props
    if (!modifyAppContainerClasses) {
      return
    }
    document
      .getElementById(appContainerId)
      .classList.remove('with-sidebar', 'sidebar-minimized', 'sidebar-transitioned')
  }

  componentDidMount() {
    window.addEventListener('resize', this.setMinimizedState)
    this.updateAppContainerClasses(true)
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.setMinimizedState)
    this.removeAppContainerClasses()
  }

  @bind
  setMinimizedState() {
    const { isMinimized, preferMinimized } = this.state

    const viewportWidth = getViewportWidth()
    if (
      (!isMinimized && viewportWidth <= LAYOUT.BREAKPOINTS.M) ||
      (isMinimized && viewportWidth > LAYOUT.BREAKPOINTS.M)
    ) {
      this.setState({ isMinimized: getViewportWidth() <= LAYOUT.BREAKPOINTS.M || preferMinimized })
      this.updateAppContainerClasses()
    }
  }

  @bind
  async onToggle() {
    await this.setState(function (prev) {
      return { isMinimized: !prev.isMinimized, preferMinimized: !prev.isMinimized }
    })
    this.updateAppContainerClasses()
  }

  @bind
  onDrawerExpandClick() {
    const { isDrawerOpen } = this.state

    if (!isDrawerOpen) {
      this.openDrawer()
    } else {
      this.closeDrawer()
    }
  }

  @bind
  onClickOutside(e) {
    const { isDrawerOpen } = this.state
    if (isDrawerOpen && this.node && !this.node.contains(e.target)) {
      this.closeDrawer()
    }
  }

  @bind
  closeDrawer() {
    this.setState({ isDrawerOpen: false })

    // Enable body scrolling.
    document.body.classList.remove(style.scrollLock)
    document.removeEventListener('mousedown', this.onClickOutside)
  }

  @bind
  openDrawer() {
    // Disable body scrolling.
    document.body.classList.add(style.scrollLock)

    document.addEventListener('mousedown', this.onClickOutside)
    this.setState({ isDrawerOpen: true })
  }

  @bind
  onLeafItemClick() {
    const { isDrawerOpen } = this.state
    if (isDrawerOpen) {
      this.onDrawerExpandClick()
    }
  }

  @bind
  ref(node) {
    this.node = node
  }

  render() {
    const { className, header, children } = this.props
    const { isMinimized, isDrawerOpen } = this.state

    const navigationClassNames = classnames(className, style.navigation, {
      [style.navigationMinimized]: isMinimized,
    })
    const minimizeButtonClassNames = classnames(style.minimizeButton, {
      [style.minimizeButtonMinimized]: isMinimized,
    })

    const drawerClassNames = classnames(style.drawer, { [style.drawerOpen]: isDrawerOpen })

    return (
      <>
        <nav className={navigationClassNames} ref={this.ref} data-test-id="navigation-sidebar">
          <div className={style.mobileHeader} onClick={this.onDrawerExpandClick}>
            <Icon className={style.expandIcon} icon="more_vert" />
            <Icon className={style.icon} icon={header.icon} />
            <Message className={style.message} content={header.title} />
          </div>
          <div>
            <div className={drawerClassNames}>
              <Link to={header.to}>
                <div className={style.header}>
                  <Icon className={style.icon} icon={header.icon} />
                  <Message className={style.message} content={header.title} />
                </div>
              </Link>
              <SideNavigationContext.Provider
                value={{ isMinimized, onLeafItemClick: this.onLeafItemClick }}
              >
                <SideNavigationList
                  onListClick={this.onDrawerExpandClick}
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
          onClick={this.onToggle}
          data-hook="side-nav-hide-button"
        />
      </>
    )
  }
}

const PortalledSideNavigation = props =>
  ReactDom.createPortal(<SideNavigation {...props} />, document.getElementById('sidebar'))

PortalledSideNavigation.Item = SideNavigationItem

export default PortalledSideNavigation
