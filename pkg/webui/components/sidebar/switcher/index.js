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

import React, { useCallback, useRef } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'
import Dropdown from '@ttn-lw/components/dropdown'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './switcher.styl'

const Switcher = ({ isMinimized }) => {
  const overviewRef = useRef(null)
  const applicationsRef = useRef(null)
  const gatewaysRef = useRef(null)
  const { pathname } = useLocation()

  const getNavLinkClass = useCallback(
    ({ isActive }) =>
      classnames(style.link, {
        [style.active]: isActive,
      }),
    [],
  )

  const getOverviewNavLinkClass = classnames(style.link, {
    [style.active]: !pathname.startsWith('/applications') && !pathname.startsWith('/gateways'),
  })

  return (
    <div
      className={classnames(style.switcherContainer, {
        [style.isMinimized]: isMinimized,
      })}
    >
      <NavLink to="/" className={getOverviewNavLinkClass} ref={overviewRef}>
        <Icon icon="home" className={style.icon} />
        <Message className={style.caption} content={sharedMessages.overview} />
        {isMinimized && (
          <Dropdown.Attached
            attachedRef={overviewRef}
            className={style.flyOutList}
            position="right"
            hover
          >
            <Dropdown.HeaderItem title={sharedMessages.overview} />
          </Dropdown.Attached>
        )}
      </NavLink>
      <NavLink to="/applications" className={getNavLinkClass} ref={applicationsRef}>
        <Icon icon="application" className={style.icon} />
        <Message className={style.caption} content={sharedMessages.applications} />
        {isMinimized && (
          <Dropdown.Attached
            attachedRef={applicationsRef}
            className={style.flyOutList}
            position="right"
            hover
          >
            <Dropdown.HeaderItem title={sharedMessages.applications} />
          </Dropdown.Attached>
        )}
      </NavLink>
      <NavLink to="/gateways" className={getNavLinkClass} ref={gatewaysRef}>
        <Icon icon="gateway" className={style.icon} />
        <Message className={style.caption} content={sharedMessages.gateways} />
        {isMinimized && (
          <Dropdown.Attached
            attachedRef={gatewaysRef}
            className={style.flyOutList}
            position="right"
            hover
          >
            <Dropdown.HeaderItem title={sharedMessages.gateways} />
          </Dropdown.Attached>
        )}
      </NavLink>
    </div>
  )
}

Switcher.propTypes = {
  isMinimized: PropTypes.bool,
}

Switcher.defaultProps = {
  isMinimized: false,
}

export default Switcher
