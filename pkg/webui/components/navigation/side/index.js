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
import PropTypes from '../../../lib/prop-types'

import Button from '../../button'
import Icon from '../../icon'
import Message from '../../../lib/components/message'
import SideNavigationList from './list'
import SideNavigationItem from './item'
import SideNavigationContext from './context'

import style from './side.styl'

const m = defineMessages({
  hideSidebar: 'Hide Sidebar',
})
export class SideNavigation extends Component {
  static defaultProps = {
    className: undefined,
  }

  static propTypes = {
    children: PropTypes.node.isRequired,
    className: PropTypes.string,
    /** The header for the side navigation */
    header: PropTypes.shape({
      title: PropTypes.string,
      icon: PropTypes.string,
    }).isRequired,
  }

  state = {
    /** A flag specifying whether the side navigation is minimized or not */
    isMinimized: false,
    /** A flag specifying whether the drawer is currently open (in mobile screensizes) */
    isDrawerOpen: false,
  }

  @bind
  onToggle() {
    this.setState(function(prev) {
      return { isMinimized: !prev.isMinimized }
    })
  }

  @bind
  onDrawerExpandClick() {
    const { isDrawerOpen } = this.state

    if (!isDrawerOpen) {
      document.addEventListener('mousedown', this.onClickOutside)
      this.setState({ isDrawerOpen: true })
    } else {
      document.addEventListener('mousedown', this.onClickOutside)
      this.setState({ isDrawerOpen: false })
    }
  }

  @bind
  onClickOutside(e) {
    const { isDrawerOpen } = this.state
    if (isDrawerOpen && this.node && !this.node.contains(e.target)) {
      this.setState({ isDrawerOpen: false })
    }
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
    const headerClassNames = classnames(style.header, {
      [style.headerMinimized]: isMinimized,
    })

    const drawerClassNames = classnames(style.drawer, { [style.drawerOpen]: isDrawerOpen })

    return (
      <nav className={navigationClassNames} ref={this.ref}>
        <div className={style.mobileHeader} onClick={this.onDrawerExpandClick}>
          <Icon className={style.expandIcon} icon="more_vert" />
          <Icon className={style.icon} icon={header.icon} />
          <Message className={style.message} content={header.title} />
        </div>
        <div className={style.body}>
          <div className={drawerClassNames}>
            <div className={headerClassNames}>
              <Icon className={style.icon} icon={header.icon} />
              <Message className={style.message} content={header.title} />
            </div>
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
          <Button
            naked
            secondary
            className={style.minimizeButton}
            icon={isMinimized ? 'keyboard_arrow_right' : 'keyboard_arrow_left'}
            message={isMinimized ? null : m.hideSidebar}
            onClick={this.onToggle}
            data-hook="side-nav-hide-button"
          />
        </div>
      </nav>
    )
  }
}

const PortalledSideNavigation = props =>
  ReactDom.createPortal(<SideNavigation {...props} />, document.getElementById('sidebar'))

PortalledSideNavigation.Item = SideNavigationItem

export default PortalledSideNavigation
