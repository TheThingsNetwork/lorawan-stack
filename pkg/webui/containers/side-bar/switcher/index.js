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

import React, { useCallback, useContext } from 'react'
import { NavLink } from 'react-router-dom'
import classNames from 'classnames'

import SideBarContext from '@ttn-lw/containers/side-bar/context'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './switcher.styl'

const Switcher = () => {
  const { layer, setLayer } = useContext(SideBarContext)

  const handleClick = useCallback(
    evt => {
      setLayer(evt.target.getAttribute('href'))
    },
    [setLayer],
  )

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
      <NavLink to="/" onClick={handleClick} className={overviewClassName}>
        <Message content={sharedMessages.overview} />
      </NavLink>
      <NavLink to="/applications" onClick={handleClick} className={applicationsClassName}>
        <Message content={sharedMessages.applications} />
      </NavLink>
      <NavLink to="/gateways" onClick={handleClick} className={gatewaysClassName}>
        <Message content={sharedMessages.gateways} />
      </NavLink>
    </div>
  )
}

export default Switcher
