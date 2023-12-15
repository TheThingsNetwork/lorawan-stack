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

import React from 'react'
import { NavLink } from 'react-router-dom'
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './switcher.styl'

const Switcher = ({ layer, onClick, isMinimized }) => {
  const overviewClassName = classnames(
    style.link,
    { [style.active]: !layer.includes('/applications') && !layer.includes('/gateways') },
    'p-vert-cs-s',
    'p-sides-0',
  )

  const applicationsClassName = classnames(
    style.link,
    { [style.active]: layer.includes('/applications') },
    'p-vert-cs-s',
    'p-sides-0',
  )

  const gatewaysClassName = classnames(
    style.link,
    { [style.active]: layer.includes('/gateways') },
    'p-vert-cs-s',
    'p-sides-0',
  )

  return (
    <div
      className={classnames(style.switcherContainer, 'd-flex', 'j-center', 'p-cs-xxs', {
        'direction-column': isMinimized,
      })}
    >
      <NavLink to="/" onClick={onClick} className={overviewClassName}>
        {isMinimized ? <Icon icon="home" /> : null}
        {!isMinimized ? <Message content={sharedMessages.overview} /> : null}
      </NavLink>
      <NavLink to="/applications" onClick={onClick} className={applicationsClassName}>
        {isMinimized ? <Icon icon="application" /> : null}
        {!isMinimized ? <Message content={sharedMessages.applications} /> : null}
      </NavLink>
      <NavLink to="/gateways" onClick={onClick} className={gatewaysClassName}>
        {isMinimized ? <Icon icon="gateway" /> : null}
        {!isMinimized ? <Message content={sharedMessages.gateways} /> : null}
      </NavLink>
    </div>
  )
}

Switcher.propTypes = {
  isMinimized: PropTypes.bool,
  layer: PropTypes.string.isRequired,
  onClick: PropTypes.func.isRequired,
}

Switcher.defaultProps = {
  isMinimized: false,
}

export default Switcher
