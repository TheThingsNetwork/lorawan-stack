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

import React, { useCallback } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './switcher.styl'

const Switcher = ({ isMinimized }) => {
  const { pathname } = useLocation()
  // <NavLink to="/"> effectively ignores the end prop and only matches when you're at the root route.
  // https://reactrouter.com/en/main/components/nav-link
  const overviewClassName = classnames(
    style.link,
    { [style.active]: !pathname.includes('/applications') && !pathname.includes('/gateways') },
    'p-vert-cs-s',
    'p-sides-0',
  )

  const className = useCallback(
    ({ isActive }) =>
      classnames(style.link, { [style.active]: isActive }, 'p-vert-cs-s', 'p-sides-0'),
    [],
  )

  return (
    <div
      className={classnames(style.switcherContainer, 'd-flex', 'j-center', 'p-cs-xxs', 'mb-cs-m', {
        'direction-column': isMinimized,
      })}
    >
      <NavLink to="/" className={overviewClassName}>
        {isMinimized ? <Icon icon="home" /> : <Message content={sharedMessages.overview} />}
      </NavLink>
      <NavLink to="/applications" className={className}>
        {isMinimized ? (
          <Icon icon="application" />
        ) : (
          <Message content={sharedMessages.applications} />
        )}
      </NavLink>
      <NavLink to="/gateways" className={className}>
        {isMinimized ? <Icon icon="gateway" /> : <Message content={sharedMessages.gateways} />}
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
