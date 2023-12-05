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
import classNames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './switcher.styl'

const Switcher = ({ layer, onClick }) => {
  const overviewClassName = classNames(
    style.link,
    { [style.active]: layer === '/' || layer === '/console' },
    'p-vert-cs-s',
    'p-sides-0',
  )

  const applicationsClassName = classNames(
    style.link,
    { [style.active]: layer.includes('/applications') },
    'p-vert-cs-s',
    'p-sides-0',
  )

  const gatewaysClassName = classNames(
    style.link,
    { [style.active]: layer.includes('/gateways') },
    'p-vert-cs-s',
    'p-sides-0',
  )

  return (
    <div
      className={classNames(
        style.switcherContainer,
        'd-flex',
        'j-center',
        'gap-cs-xxs',
        'p-cs-xxs',
      )}
    >
      <NavLink to="/" onClick={onClick} className={overviewClassName}>
        <Message content={sharedMessages.overview} />
      </NavLink>
      <NavLink to="/applications" onClick={onClick} className={applicationsClassName}>
        <Message content={sharedMessages.applications} />
      </NavLink>
      <NavLink to="/gateways" onClick={onClick} className={gatewaysClassName}>
        <Message content={sharedMessages.gateways} />
      </NavLink>
    </div>
  )
}

Switcher.propTypes = {
  layer: PropTypes.string.isRequired,
  onClick: PropTypes.func.isRequired,
}

export default Switcher
